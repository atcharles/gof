package g2util

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/andeya/goutil"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

// Config ...
type Config struct {
	rootPath string
	filename string
	viper    *viper.Viper
	mutex    sync.RWMutex
}

// LockFunc ...
func (c *Config) LockFunc(fn func(v *viper.Viper) (e error)) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return fn(c.viper)
}

// AppName ...
func (c *Config) AppName() string { return c.viper.GetString("name") }

// SetFilename ...
func (c *Config) SetFilename(filename string) { c.filename = filename }

// RootPath ...
func (c *Config) RootPath() string { return c.rootPath }

// SetRootPath ...
func (c *Config) SetRootPath(rootPath string) { c.rootPath = rootPath }

// Viper ...
func (c *Config) Viper() *viper.Viper { return c.viper }

// Constructor ...初始化
func (c *Config) Constructor() { c.viper = viper.New(); c.rootPath = goutil.SelfDir() }

// Load ...
func (c *Config) Load(args ...interface{}) {
	if e := c.load(args...); e != nil {
		log.Fatalln("load conf file: ", e)
	}
}

//load ...
/**
 * @Description:
 * @receiver c
 * @param args rootDir,configName,absConfigFile
 * @return err
 */
func (c *Config) load(args ...interface{}) (err error) {
	if c.viper == nil {
		c.Constructor()
	}
	v := c.viper
	if len(args) > 2 {
		if f1 := cast.ToString(args[2]); len(f1) > 0 {
			v.SetConfigFile(f1)
		}
	}
	v.SetConfigType("yml")

	filename := "conf"
	if len(c.filename) > 0 {
		filename = c.filename
	}
	if len(args) > 1 {
		if f1 := cast.ToString(args[1]); len(f1) > 0 {
			filename = f1
		}
	}
	v.SetConfigName(filename)

	dirName := c.rootPath
	if len(args) > 0 {
		if f1 := cast.ToString(args[0]); len(f1) > 0 {
			dirName = f1
			c.rootPath = dirName
		}
	}
	_fn1entry := func(dirName string) {
		v.AddConfigPath(dirName)
		var entries []fs.DirEntry
		entries, err = os.ReadDir(dirName)
		if err != nil {
			return
		}
		for _, e := range entries {
			if e.IsDir() {
				v.AddConfigPath(filepath.Join(dirName, e.Name()))
			}
		}
	}
	_fn1entry(dirName)
	_fn1entry(filepath.Dir("."))
	if err = v.ReadInConfig(); err != nil {
		return
	}
	if err = c.LoadFromTemp(); err != nil {
		return
	}
	c.viper.WatchConfig()
	return
}

func getRoot() string {
	ex, _ := os.Executable()
	if strings.Contains(ex, "go-build") {
		return "."
	}
	return filepath.Dir(ex)
}

// TempFile ......
func (c *Config) TempFile() string {
	s := filepath.Join(getRoot(), "data", "conf_temp.yml")
	_ = os.MkdirAll(filepath.Dir(s), 0755)
	return s
}

// WriteToTemp ......
func (c *Config) WriteToTemp(fn func(v *viper.Viper)) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	fn(c.viper)
	return c.viper.WriteConfigAs(c.TempFile())
}

// LoadFromTemp ......
func (c *Config) LoadFromTemp() error {
	tempFile := c.TempFile()
	if _, e := os.Stat(tempFile); e != nil {
		return nil
	}
	v := viper.New()
	v.SetConfigFile(tempFile)
	if e := v.ReadInConfig(); e != nil {
		return e
	}
	if c.viper == nil {
		return nil
	}
	if e := c.viper.MergeConfigMap(v.AllSettings()); e != nil {
		return e
	}
	if e := c.viper.WriteConfig(); e != nil {
		return e
	}
	_ = os.Remove(tempFile)
	return nil
}
