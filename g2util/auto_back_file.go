package g2util

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

//每秒写入到文件一次,阻塞当前数据写入
//当buffer超过指定值,写入到文件
//自动对大文件进行打包
//自动删除n天前的文件

// 常量,大小定义
const (
	_ int = 1 << (10 * iota)
	//ignore KB
	_
	MB
)

// IWriter ...
type IWriter interface {
	AfterShutdown()
	Closed() bool
	Close()
	Flush() (err error)
	Write(p []byte) (n int, err error)
	File() *os.File
}

// AbFile ...
type AbFile struct {
	Config   *Config   `inject:""`
	Graceful *Graceful `inject:""`

	mu sync.RWMutex

	files map[string]*innerIO
}

// MustLogIO ...
func (a *AbFile) MustLogIO(name string) IWriter {
	return a.MustNewIO(filepath.Join(a.Config.RootPath(), "logs", name+".log"))
}

// MustNewIO ...获取一个io对象,一旦出错,将会panic
func (a *AbFile) MustNewIO(filePath string, opts ...*ABFileOption) IWriter {
	o, err := a.newIO(filePath, opts...)
	if err != nil {
		panic(err)
	}

	a.Graceful.RegProcessor(o)
	return o
}

// Constructor ...
func (a *AbFile) Constructor() { a.constructor() }

func (a *AbFile) constructor() *AbFile { a.files = make(map[string]*innerIO); return a }

// newIO ...
func (a *AbFile) newIO(filePath string, opts ...*ABFileOption) (out *innerIO, err error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	var has bool
	out, has = a.files[filePath]
	if has {
		return
	}
	_ = os.MkdirAll(filepath.Dir(filePath), 0755)
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		return
	}
	out = &innerIO{
		buf: bytes.NewBuffer([]byte{}),
		f:   file,
		opt: &ABFileOption{
			FileMaxSize:     2 * MB,
			SaveDays:        3,
			BufferSize:      1 * MB,
			AutoFlushPeriod: time.Second * 2,
			AutoDelPeriod:   time.Minute * 60,
		},
		closed:     false,
		close:      make(chan struct{}),
		dateFormat: "20060102",
	}
	if len(opts) > 0 && opts[0] != nil {
		out.opt = opts[0]
	}
	go out.run()
	a.files[filePath] = out
	return
}

type (
	//ABFileOption 参数配置
	ABFileOption struct {
		//文件最大值,超过该值自动备份,备份文件大小略大于该值
		FileMaxSize int
		//保存目录天数
		SaveDays int
		//writer buffer size,内存缓冲区大小
		BufferSize int
		//自动刷新数据到磁盘的周期,默认值为2秒
		AutoFlushPeriod time.Duration
		//检查删除过期文件目录间隔
		AutoDelPeriod time.Duration
	}
	innerIO struct {
		sync.RWMutex
		buf        *bytes.Buffer
		f          *os.File
		opt        *ABFileOption
		closed     bool
		close      chan struct{}
		dateFormat string
	}
)

// File ...
func (i *innerIO) File() *os.File {
	i.RLock()
	defer i.RUnlock()
	return i.f
}

func (i *innerIO) AfterShutdown() {
	tf := TimeNow().Format("2006-01-02 15:04:05.000000")
	_, _ = i.Write([]byte(fmt.Sprintf("[%s] 👉 grace shutdown, flush data\n", tf)))
	i.Close()
}

// Closed 是否关闭
func (i *innerIO) Closed() bool {
	i.RLock()
	defer i.RUnlock()
	return i.closed
}

// Close 关闭io
func (i *innerIO) Close() {
	if i.Closed() {
		return
	}
	i.Lock()
	i.closed = true
	close(i.close)
	i.Unlock()
}

// run ...
func (i *innerIO) run() {
	tk := time.NewTicker(i.opt.AutoFlushPeriod)
	tk2 := time.NewTicker(i.opt.AutoDelPeriod)
	defer func() {
		tk.Stop()
		tk2.Stop()
	}()
	var err error
	for {
		select {
		case <-tk2.C:
			err = i.lockDelDir()
		case <-tk.C:
			err = i.Flush()
		case <-i.close:
			_ = i.flushAndCloseFile()
			return
		}
		if err != nil {
			log.Printf("abio error: %s\n", err.Error())
		}
	}
}

// lockDelDir ...对文件加锁读取,定时删除目录
func (i *innerIO) lockDelDir() (err error) {
	i.Lock()
	fName := i.f.Name()
	err = i.autoDelDir(fName)
	i.Unlock()
	return
}

// flushAndCloseFile ...在close channel 触发
func (i *innerIO) flushAndCloseFile() (err error) {
	i.Lock()
	defer i.Unlock()
	if err = i.flush(); err != nil {
		return
	}
	return i.f.Close()
}

// Flush 将数据刷入到磁盘
func (i *innerIO) Flush() (err error) {
	i.Lock()
	err = i.flush()
	i.Unlock()
	return
}

// flush ...从缓存中将数据刷入到文件中
func (i *innerIO) flush() (err error) {
	if i.buf.Len() == 0 {
		return
	}
	_, err = i.f.Write(i.buf.Bytes())
	if err != nil {
		return
	}
	i.buf.Reset()
	if i.getCurrentFileSize() >= i.opt.FileMaxSize {
		if err = i.fileCopyWithIndex(); err != nil {
			return
		}
	}
	return
}

// getCurrentFileSize ...获取当前打开的文件的大小
func (i *innerIO) getCurrentFileSize() (fSize int) {
	fInfo, _ := i.f.Stat()
	fSize = int(fInfo.Size())
	return
}

func (i *innerIO) Write(p []byte) (n int, err error) {
	if i.Closed() {
		return
	}
	i.Lock()
	defer i.Unlock()
	_, err = i.buf.Write(p)
	if err != nil {
		return
	}
	if i.buf.Len() < i.opt.BufferSize {
		return
	}
	if err = i.flush(); err != nil {
		return
	}
	return
}

// todayStr 获取当前日期字符串
func (i *innerIO) todayStr() string { return TimeNow().Format(i.dateFormat) }

// 获取最近...天的日期字符串
func (i *innerIO) latestNDayStr() []string {
	days := make([]string, 0)
	for a := 0; a < i.opt.SaveDays; a++ {
		d := TimeNow().Add(24 * time.Hour * time.Duration(a) * -1)
		days = append(days, d.Format(i.dateFormat))
	}
	return days
}

// 获取文件名命名的目录+日期后缀的目录名
func (*innerIO) fileDirWithDaySuffix(filePath, day string) string {
	dir := filepath.Dir(filePath)
	f1 := strings.Replace(filepath.Base(filePath), filepath.Ext(filePath), "", -1)
	return filepath.Join(dir, f1+day)
}

func (*innerIO) fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		return !os.IsNotExist(err)
	}
	return true
}

// fileCopyWithIndex ...
// 获取目录中的文件数目,填充6位数字,加上原文件扩展名,并将原文件完整拷贝到新文件
// 创建或者获取一个新的目录,目录加上文件名前缀,以当日日期结尾
// 遍历目录,找到指定拷贝的文件后缀类型的文件个数
// 拷贝文件
func (i *innerIO) fileCopyWithIndex() (err error) {
	fileSrc := i.f
	fileStat, err := fileSrc.Stat()
	if err != nil {
		return
	}
	if fileStat.IsDir() {
		return errors.New("拷贝文件,文件不能是一个目录")
	}
	dir, err := filepath.Abs(i.fileDirWithDaySuffix(fileSrc.Name(), i.todayStr()))
	if err != nil {
		return
	}
	_ = os.MkdirAll(dir, 0755)
	//计数,目录下文件个数
	fileExt := filepath.Ext(fileSrc.Name())
	countFile := 0
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(info.Name()) == fileExt {
			countFile++
		}
		return nil
	})
	if err != nil {
		return
	}
	//新文件名
	newFilePathName := filepath.Join(dir, fmt.Sprintf("%06d%s", countFile, fileExt))
	//copy file
	fNew, err := os.Create(newFilePathName)
	if err != nil {
		return
	}
	defer func() {
		_ = fNew.Close()
	}()
	_, _ = fileSrc.Seek(0, 0)
	if _, err = io.Copy(fNew, fileSrc); err != nil {
		return
	}
	_ = fileSrc.Truncate(0)
	return
}

// autoDelDir ...自动删除过期日志目录
func (i *innerIO) autoDelDir(rootFilePath string) (err error) {
	if !i.fileExists(rootFilePath) {
		return
	}
	dir, err := filepath.Abs(filepath.Dir(rootFilePath))
	if err != nil {
		return
	}
	fileAbsName := strings.Replace(filepath.Base(rootFilePath), filepath.Ext(rootFilePath), "", -1)
	var unDelDirs []string
	for _, s := range i.latestNDayStr() {
		unDelDirs = append(unDelDirs, filepath.Join(dir, fmt.Sprintf("%s%s", fileAbsName, s)))
	}
	delList := make([]string, 0)
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			return nil
		}
		if path == dir {
			return nil
		}
		rgp := regexp.MustCompile(fmt.Sprintf(`^%s\d+$`, fileAbsName))
		if !rgp.MatchString(info.Name()) {
			return nil
		}
		for _, n1 := range unDelDirs {
			if n1 == path {
				return nil
			}
		}
		delList = append(delList, path)
		return nil
	})
	for _, s := range delList {
		_ = os.RemoveAll(s)
	}
	return
}
