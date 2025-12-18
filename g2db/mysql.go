package g2db

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/novalagung/gubrak/v2"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"xorm.io/xorm"
	log2 "xorm.io/xorm/log"
	"xorm.io/xorm/names"

	//load mysql driver
	_ "github.com/go-sql-driver/mysql"

	"github.com/atcharles/gof/v2/g2util"
)

// Mysql ...
type Mysql struct {
	Config    *g2util.Config   `inject:""`
	Grace     *g2util.Graceful `inject:""`
	AbFile    *g2util.AbFile   `inject:""`
	Redis     *redisObj        `inject:""`
	Cache     *cacheMem        `inject:""`
	CacheBind *cacheBind       `inject:""`

	mu     sync.RWMutex
	eg     *xorm.Engine
	out    io.Writer
	tables []interface{}
}

// Tables ...
func (m *Mysql) Tables() []interface{} { return m.tables }

// Constructor New ...
func (m *Mysql) Constructor() { m.tables = make([]interface{}, 0) }

// AfterShutdown ...
func (m *Mysql) AfterShutdown() {
	if m.eg != nil {
		_ = m.eg.Close()
	}
}

// Engine ...
func (m *Mysql) Engine() *xorm.Engine { return m.eg }

// SetEngine ......
func (m *Mysql) SetEngine(e *xorm.Engine) { m.eg = e }

// Dial MySQL连接拨号
func (m *Mysql) Dial() {
	if e := m.dial(); e != nil {
		log.Fatalf("数据库连接失败:%s\n", e.Error())
	}
	//订阅Redis
	m.Redis.Subscribe()
	m.Grace.RegProcessor(m)
	m.Grace.RegProcessor(m.Redis)
}

// Insert ...
func (m *Mysql) Insert(bean interface{}) error {
	return m.TXCallback(func(sn *xorm.Session) error { return m.Session(sn).Insert(bean) })
}

// Update ...
func (m *Mysql) Update(bean interface{}, params ...interface{}) (newBean interface{}, err error) {
	err = m.TXCallback(func(sn *xorm.Session) error {
		v, e := m.Session(sn).Update(bean, params...)
		if e != nil {
			return e
		}
		newBean = v
		return nil
	})
	return
}

// Delete ...
func (m *Mysql) Delete(bean interface{}) (err error) {
	return m.TXCallback(func(sn *xorm.Session) error { return m.Session(sn).Delete(bean) })
}

// Session ...
func (m *Mysql) Session(sn *xorm.Session) *Session { return newSession(m, sn) }

// DelCache ...
func (m *Mysql) DelCache(bean interface{}, condition ...interface{}) (err error) {
	return m.Redis.PubDelCache(m.CacheMemKeys(bean, condition...))
}

// CacheMemKeys ...
func (m *Mysql) CacheMemKeys(bean interface{}, condition ...interface{}) (list []string) {
	queryList := new(cacheBind).Values(bean, condition...)
	list = make([]string, 0)
	for _, s := range queryList {
		list = append(list, memKey(bean, s))
	}
	return
}

// CacheRestore ...
func (m *Mysql) CacheRestore(bean interface{}) (err error) {
	val, err := json.Marshal(bean)
	if err != nil {
		return
	}
	for _, k := range m.CacheMemKeys(bean) {
		m.Cache.Atomic(k, func() { err = m.Cache.Cache.Set(k, val) })
		if err != nil {
			return
		}
	}
	return
}

// CacheGet ...
func (m *Mysql) CacheGet(bean interface{}, condition ...interface{}) (err error) {
	return m.cacheGet(bean, []string{"Unscoped"}, condition...)
}

// CacheGetWrapSession ...
func (m *Mysql) CacheGetWrapSession(bean interface{}, arg interface{}, condition ...interface{}) (err error) {
	return m.cacheGet(bean, arg, condition...)
}

// ErrorMysqlNotFound ...
type ErrorMysqlNotFound string

func (e ErrorMysqlNotFound) Error() string { return string(e) }

func (m *Mysql) cacheGet(bean interface{}, arg interface{}, condition ...interface{}) (err error) {
	queryList := m.CacheBind.Values(bean, condition...)
	if len(queryList) == 0 {
		return errors.New("查询条件为空")
	}
	query := queryList[0]
	key := memKey(bean, query)
	bts, err := m.Cache.GetOrStore(key, func() (b []byte, e error) {
		vb := g2util.NewValue(bean)
		sn := m.eg.Context(context.Background())

		if arg != nil {
			switch _vv := arg.(type) {
			case []string:
				for _, _s := range _vv {
					switch _s {
					case "Unscoped":
						sn = sn.Unscoped()
					case "NoCache":
						sn = sn.NoCache()
					case "NoLog":
						sn = sn.MustLogSQL(false)
					}
				}
			case func(*xorm.Session) *xorm.Session:
				sn = _vv(sn)
			}
		}

		has, e := sn.NoAutoCondition().Where(query).Get(vb)
		if e != nil {
			return
		}
		if !has {
			b = []byte(cacheNULL)
			return
		}
		return json.Marshal(vb)
	})
	if err != nil {
		return
	}
	if string(bts) == cacheNULL {
		return ErrorMysqlNotFound(fmt.Sprintf("数据不存在: %s", key))
	}
	return json.Unmarshal(bts, bean)
}

// Migrate ...数据库初始化
func (m *Mysql) Migrate() {
	d := g2util.TimeExcWrap(func() {
		if e := m.migrate(); e != nil {
			log.Fatalln(e)
		}
	})
	log.Printf("数据初始化完成,use:%s\n", d)
}

func (m *Mysql) migrate() (err error) {
	if err = m.createDB(); err != nil {
		return
	}
	if err = m.dial(); err != nil {
		return
	}
	if err = m.sync(); err != nil {
		return
	}
	_ = m.eg.Close()
	return
}

// TXCallback ...
func (m *Mysql) TXCallback(fn func(sn *xorm.Session) (err error)) (err error) {
	sn := m.eg.NewSession()
	defer func() { _ = sn.Close() }() //不管是否存在err,总是close
	if err = sn.Begin(); err != nil {
		return
	}
	if err = fn(sn); err != nil {
		e := sn.Rollback()
		if e != nil {
			return fmt.Errorf("%s,tx Rollback:%s", err.Error(), e.Error())
		}
		return
	}
	if err = sn.Commit(); err != nil {
		return fmt.Errorf("tx Commit:%s", err.Error())
	}
	return
}

// GetOut ...
func (m *Mysql) GetOut() io.Writer { return m.getOut() }
func (m *Mysql) getOut() io.Writer {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.out == nil {
		m.out = m.AbFile.MustLogIO("sql")
	}
	return m.out
}

// DbName ...
func (m *Mysql) DbName() string {
	v := m.Config.Viper()
	db := v.GetString("mysql.db")
	if len(db) == 0 {
		db = m.Config.AppName()
	}
	return db
}

// getDataSource ...
func (m *Mysql) getDataSource(args ...bool) string {
	//withDB, 当需要将链接db去除时(创建数据库),设置为false
	withDB := true
	if len(args) > 0 {
		withDB = args[0]
	}

	v := m.Config.Viper()
	dsn := v.GetString("mysql.dsn")
	db := m.DbName()
	dsn = strings.Replace(dsn, "{host}", v.GetString("global.host"), -1)
	if withDB {
		return strings.Replace(dsn, "{db}", db, -1)
	}
	return strings.Replace(dsn, "{db}", "", -1)
}

func (m *Mysql) dial() (err error) {
	dataSource := m.getDataSource()
	e, err := xorm.NewEngine("mysql", dataSource)
	if err != nil {
		return
	}

	v := m.Config.Viper()
	valMap := v.GetStringMapString("mysql")

	e.SetDisableGlobalCache(!cast.ToBool(valMap["use_cache"]))
	e.SetMapper(names.LintGonicMapper)

	//e.SetLogLevel(log2.LOG_OFF)
	e.SetLogger(log2.NewSimpleLogger(m.GetOut()))
	_logLevel := func() log2.LogLevel {
		lvl := v.GetString("mysql.log_level")
		switch lvl {
		case "warn", "warning":
			return log2.LOG_WARNING
		case "error":
			return log2.LOG_ERR
		case "debug":
			return log2.LOG_DEBUG
		case "info":
			return log2.LOG_INFO
		default:
			return log2.LOG_OFF
		}
	}
	e.SetLogLevel(_logLevel())
	e.ShowSQL(cast.ToBool(valMap["show_sql"]))

	e.SetConnMaxLifetime(cast.ToDuration(valMap["max_conn_lifetime_seconds"]) * time.Second)
	e.SetMaxIdleConns(cast.ToInt(valMap["max_idle_connections"]))
	e.SetMaxOpenConns(cast.ToInt(valMap["max_open_connections"]))
	if err = e.Unscoped().MustLogSQL(false).Ping(); err != nil {
		return
	}

	//go g2util.Ticker(time.Second*30, func() { _ = e.Unscoped().MustLogSQL(false).Ping() })

	m.eg = e
	return
}

// DropDatabase ...
func (m *Mysql) DropDatabase() (err error) {
	log.Println("删除数据库")
	sq := fmt.Sprintf(`DROP DATABASE IF EXISTS %s;`, m.DbName())
	return m.ExecSqOnNewEngine(sq)
}

func (m *Mysql) createDB() (err error) {
	log.Println("创建数据库")
	defer func() {
		if err == nil {
			log.Printf("数据库创建完成,使用数据库: %s\n", m.DbName())
		}
	}()
	createDBSql := fmt.Sprintf(`CREATE DATABASE IF NOT EXISTS %s;`, m.DbName())
	return m.ExecSqOnNewEngine(createDBSql)
}

// ExecSqOnNewEngine ...
func (m *Mysql) ExecSqOnNewEngine(sq string) (err error) {
	dataSource := m.getDataSource(false)
	dba, err := xorm.NewEngine("mysql", dataSource)
	if err != nil {
		return
	}
	defer func() { _ = dba.Close() }()
	_, err = dba.Exec(sq)
	return
}

// DialWithMysql ......
func (m *Mysql) DialWithMysql(fn func(x *xorm.Engine) error) (err error) {
	dataSource := m.getDataSource(false)
	dba, err := xorm.NewEngine("mysql", dataSource)
	if err != nil {
		return
	}
	defer func() { _ = dba.Close() }()
	return fn(dba)
}

// TableRegister ...注册表,用于同步数据表 ... 等
func (m *Mysql) TableRegister(tables ...interface{}) {
	_f1hasTable := func(tb interface{}) bool {
		for _, t1 := range m.tables {
			if tableName(t1) == tableName(tb) {
				return true
			}
		}
		return false
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, table := range tables {
		if _f1hasTable(table) {
			return
		}
		m.tables = append(m.tables, table)
	}
}

// Sync ...初始化数据表,结构,数据等
func (m *Mysql) Sync() (err error) { return m.sync() }
func (m *Mysql) sync() (err error) {
	return m.TXCallback(func(sn *xorm.Session) (e error) {
		if len(m.tables) == 0 {
			return
		}
		if e = sn.Sync2(m.tables...); e != nil {
			return
		}
		for _, table := range m.tables {
			tbName := tableName(table)
			sq1 := fmt.Sprintf("ALTER TABLE `%s` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;", tbName)
			if _, e = sn.Exec(sq1); e != nil {
				return
			}

			//复合索引
			if obj, ok := table.(ItfCompoundIndex); ok {
				for _, c1 := range obj.CompoundIndexes() {
					c1.SetSn(sn)
					if e = c1.execCreate(tbName); e != nil {
						return fmt.Errorf("[CompoundIndex] [%s] error: %w", tbName, e)
					}
				}
			}

			if obj, ok := table.(ItfInitData); ok {
				if e = initializeTables(obj, sn); e != nil {
					return fmt.Errorf("[InitData] [%s] error: %w", tbName, e)
				}
			}

			if obj, ok := table.(ItfAfterSync); ok {
				if e = obj.AfterSync(sn); e != nil {
					return fmt.Errorf("[AfterSync] [%s] error: %w", tbName, e)
				}
			}
		}
		return
	})
}

type (
	//ItfCompoundIndex 复合索引接口
	ItfCompoundIndex interface {
		CompoundIndexes() []*CompoundIndex
	}
	//CompoundIndex 复合索引
	CompoundIndex struct {
		//索引字段
		Columns map[string]interface{}
		//是否是唯一索引
		Unique bool

		sn *xorm.Session
	}

	//ItfInitData ...
	ItfInitData interface {
		InitData() []interface{}
	}
)

// SetSn ...
func (c *CompoundIndex) SetSn(sn *xorm.Session) { c.sn = sn }

// makeQuery ...
func (c *CompoundIndex) makeQuery() string {
	if !c.Unique {
		return ""
	}
	list := make([]string, 0)
	for k, v := range c.Columns {
		vv := reflect.ValueOf(v)
		if vv.IsValid() && !vv.IsZero() {
			list = append(list, fmt.Sprintf("%s=%s", fieldName(k), new(cacheBind).cacheValString(vv)))
		}
	}
	sort.Strings(list)
	list = gubrak.From(list).OrderBy(func(each string) int { return len(each) }, true).Result().([]string)
	return strings.Join(list, " AND ")
}

// execCreate ...创建索引
func (c *CompoundIndex) execCreate(table string) (err error) {
	if c.sn == nil {
		panic("orm Session is nil!")
	}

	cs1 := func() []string {
		ss1 := make([]string, 0)
		for k := range c.Columns {
			ss1 = append(ss1, fieldName(k))
		}
		return ss1
	}()
	if len(cs1) < 2 {
		return
	}

	_unique := func() string {
		if c.Unique {
			return "UNIQUE "
		}
		return ""
	}

	_indexName := func() string {
		prefix := "CUK"
		if !c.Unique {
			prefix = "CIX"
		}
		return fmt.Sprintf("%s_%s_%s", prefix, table, strings.Join(cs1, "_"))
	}

	mp1 := g2util.Map{
		"table":      table,
		"unique":     _unique(),
		"index_name": _indexName(),
		"indexes":    strings.Join(cs1, ","),
	}

	sq1 := `SHOW INDEX FROM {{.table}} WHERE Key_name = '{{.index_name}}'`
	sq1 = g2util.TextTemplateMustParse(sq1, mp1)
	rs1, err := c.sn.MustLogSQL(true).SQL(sq1).Query()
	if err != nil {
		return fmt.Errorf("查询复合索引失败: %w,\n%s", err, sq1)
	}
	if len(rs1) > 0 {
		return
	}
	sq2 := `ALTER TABLE {{.table}} ADD {{.unique}}INDEX {{.index_name}} ({{.indexes}})`
	sq2 = g2util.TextTemplateMustParse(sq2, mp1)
	_, err = c.sn.Exec(sq2)
	return
}

// TableName ...
func TableName(bean interface{}) string { return tableName(bean) }

// FieldName ...
func FieldName(field string) (name string) { return names.LintGonicMapper.Obj2Table(field) }

// tableName CacheTableName ...
// redis-key = /h/tableName
// 2020/3/30 22:11 -- Author:charles
// func CacheTableName(table interface{}) string { return tableName(table) }
func tableName(table interface{}) string {
	/*tnBean, ok := table.(names.TableName)
	if ok {
		return tnBean.TableName()
	}
	val1 := reflect.Indirect(reflect.ValueOf(table))
	return names.LintGonicMapper.Obj2Table(val1.Type().Name())*/
	return names.GetTableName(names.LintGonicMapper, reflect.ValueOf(table))
}

// fieldName 获取模型对象字段 => 数据库的字段名
// 参与缓存的字段,tag中不能有自定义的 name (字段名)
func fieldName(field string) (name string) { return names.LintGonicMapper.Obj2Table(field) }

// memKey ...
func memKey(b interface{}, key string) string { return fmt.Sprintf("%s::%s", tableName(b), key) }

// MyBase1 id and created
type MyBase1 struct {
	ID      int64            `json:"id,omitempty" xorm:"pk autoincr"`
	Created *g2util.JSONTime `json:"created,omitempty" xorm:"notnull default CURRENT_TIMESTAMP created index comment('创建时间')"`
}

// MyBase xorm MySQL model base
type MyBase struct {
	MyBase1 `xorm:"extends"`
	Updated *g2util.JSONTime `json:"updated,omitempty" xorm:"notnull default CURRENT_TIMESTAMP updated comment('更新时间')"`
	Version int64            `json:"version,omitempty" xorm:"notnull default 1 version comment('乐观锁')"`
}

// ClearColumns ...
func (m *MyBase) ClearColumns() {
	m.Version = 0
	m.Created = nil
	m.Updated = nil
}
