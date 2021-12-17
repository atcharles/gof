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

//Mysql ...
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

//Tables ...
func (m *Mysql) Tables() []interface{} { return m.tables }

// Constructor New ...
func (m *Mysql) Constructor() { m.tables = make([]interface{}, 0) }

//AfterShutdown ...
func (m *Mysql) AfterShutdown() {
	if m.eg != nil {
		_ = m.eg.Close()
	}
}

//Engine ...
func (m *Mysql) Engine() *xorm.Engine { return m.eg }

//Dial MySQL连接拨号
func (m *Mysql) Dial() {
	if e := m.dial(); e != nil {
		log.Fatalf("数据库连接失败:%s\n", e.Error())
	}
	//订阅Redis
	m.Redis.Subscribe()
	m.Grace.RegProcessor(m)
	m.Grace.RegProcessor(m.Redis)
}

//Insert ...
func (m *Mysql) Insert(bean interface{}) error {
	return m.TXCallback(func(sn *xorm.Session) (err error) {
		if v1, ok := bean.(ItfSessionBeforeInsert); ok {
			if err = v1.SessionBeforeInsert(sn); err != nil {
				return
			}
		}
		a, err := sn.InsertOne(bean)
		if err != nil {
			return
		}
		if a == 0 {
			return errors.New("数据写入失败")
		}
		if err = m.DelCache(bean); err != nil {
			return
		}
		if v1, ok := bean.(ItfSessionAfterInsert); ok {
			if err = v1.SessionAfterInsert(sn); err != nil {
				return
			}
		}
		return
	})
}

//Update ...
//@Description:
//@receiver m
//@param bean
//@param params
//@return newBean
//@return err
func (m *Mysql) Update(bean interface{}, params ...interface{}) (newBean interface{}, err error) {
	newBean, err = g2util.CopyBean(bean)
	if err != nil {
		return
	}
	if err = m.CacheGet(newBean); err != nil {
		return
	}

	if err = g2util.MergeBeans(newBean, bean); err != nil {
		return
	}

	var (
		cols       []string
		queryAfter bool
	)
	if len(params) > 0 {
		for _, param := range params {
			switch val := param.(type) {
			case g2util.Map:
				if err = val.Merge2Bean(newBean); err != nil {
					return
				}
			case []string:
				cols = val
			case bool:
				queryAfter = val
			}
		}
	}

	queryList := new(cacheBind).Values(newBean)
	if len(queryList) == 0 {
		return
	}

	_f1 := func(sn *xorm.Session) *xorm.Session {
		sn = sn.NoAutoCondition().Where(queryList[0])
		if len(cols) == 0 {
			return sn.UseBool().AllCols()
		}
		return sn.Cols(cols...)
	}
	err = m.TXCallback(func(sn *xorm.Session) error {
		var e error
		if v, ok := newBean.(ItfSessionBeforeUpdate); ok {
			if e = v.SessionBeforeUpdate(sn); e != nil {
				return e
			}
		}
		a, e := _f1(sn).Update(newBean)
		if e != nil {
			return e
		}
		if a == 0 {
			return errors.New("未更新数据")
		}
		//更新,删除新条件缓存
		if e = m.DelCache(newBean); e != nil {
			return e
		}
		if v, ok := newBean.(ItfSessionAfterUpdate); ok {
			if e = v.SessionAfterUpdate(sn); e != nil {
				return e
			}
		}
		if queryAfter {
			_, e = sn.NoAutoCondition().Where(queryList[0]).Get(newBean)
			if e != nil {
				return e
			}
		}
		return e
	})
	return
}

//Delete ...
func (m *Mysql) Delete(bean interface{}) (err error) {
	queryList := new(cacheBind).Values(bean)
	if len(queryList) == 0 {
		return
	}
	if err = m.CacheGet(bean); err != nil {
		return
	}
	return m.TXCallback(func(sn *xorm.Session) error {
		var e error
		if v, ok := bean.(ItfSessionBeforeDelete); ok {
			if e = v.SessionBeforeDelete(sn); e != nil {
				return e
			}
		}
		if _, e = sn.NoAutoCondition().Where(queryList[0]).Delete(bean); e != nil {
			return e
		}
		if e = m.DelCache(bean); e != nil {
			return e
		}
		if v, ok := bean.(ItfSessionAfterDelete); ok {
			if e = v.SessionAfterDelete(sn); e != nil {
				return e
			}
		}
		return e
	})
}

//DelCache ...
func (m *Mysql) DelCache(bean interface{}, condition ...interface{}) (err error) {
	return m.Redis.PubDelCache(m.cacheMemKeys(bean, condition...))
}

//cacheMemKeys ...
func (m *Mysql) cacheMemKeys(bean interface{}, condition ...interface{}) (list []string) {
	queryList := new(cacheBind).Values(bean, condition...)
	list = make([]string, 0)
	for _, s := range queryList {
		list = append(list, memKey(bean, s))
	}
	return
}

//CacheGet ...
func (m *Mysql) CacheGet(bean interface{}, condition ...interface{}) (err error) {
	return m.cacheGet(bean, []string{"Unscoped"}, condition...)
}

//CacheGetWrapSession ...
func (m *Mysql) CacheGetWrapSession(bean interface{}, arg interface{}, condition ...interface{}) (err error) {
	return m.cacheGet(bean, arg, condition...)
}

//ErrorMysqlNotFound ...
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

//Migrate ...数据库初始化
func (m *Mysql) Migrate() {
	d := g2util.TimeExcWrap(func() {
		if e := m.migrate(); e != nil {
			log.Fatalln(e)
		}
	})
	log.Printf("数据初始化完成,use:%s", d)
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

//TXCallback ...
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

//GetOut ...
func (m *Mysql) GetOut() io.Writer { return m.getOut() }
func (m *Mysql) getOut() io.Writer {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.out == nil {
		m.out = m.AbFile.MustLogIO("sql")
	}
	return m.out
}

//dbName ...
func (m *Mysql) dbName() string {
	v := m.Config.Viper()
	db := v.GetString("mysql.db")
	if len(db) == 0 {
		db = m.Config.AppName()
	}
	return db
}

//getDataSource ...
func (m *Mysql) getDataSource(args ...bool) string {
	//withDB, 当需要将链接db去除时(创建数据库),设置为false
	withDB := true
	if len(args) > 0 {
		withDB = args[0]
	}

	v := m.Config.Viper()
	dsn := v.GetString("mysql.dsn")
	db := m.dbName()
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

	e.SetLogLevel(log2.LOG_OFF)
	if cast.ToBool(valMap["show_sql"]) {
		e.SetLogger(log2.NewSimpleLogger(m.GetOut()))
		e.ShowSQL(true)
		_logLevel := func() log2.LogLevel {
			lvl := v.GetString("mysql.log_level")
			switch lvl {
			case "warn", "warning":
				return log2.LOG_WARNING
			case "error":
				return log2.LOG_ERR
			case "debug":
				return log2.LOG_DEBUG
			default:
				return log2.LOG_INFO
			}
		}
		e.SetLogLevel(_logLevel())
	}
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

func (m *Mysql) createDB() (err error) {
	dataSource := m.getDataSource(false)
	dba, err := xorm.NewEngine("mysql", dataSource)
	if err != nil {
		return
	}
	defer func() {
		_ = dba.Close()
	}()
	db := m.dbName()
	createDBSql := fmt.Sprintf(`CREATE DATABASE IF NOT EXISTS %s;`, db)
	_, err = dba.Exec(createDBSql)
	return
}

//TableRegister ...注册表,用于同步数据表 ... 等
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

//Sync ...初始化数据表,结构,数据等
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
						return
					}
				}
			}

			if obj, ok := table.(ItfInitData); ok {
				if e = initializeTables(obj, sn); e != nil {
					return
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
	ItfInitData interface{ InitData() []interface{} }
)

//SetSn ...
func (c *CompoundIndex) SetSn(sn *xorm.Session) { c.sn = sn }

//makeQuery ...
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

//execCreate ...创建索引
func (c *CompoundIndex) execCreate(table string) (err error) {
	if c.sn == nil {
		panic("orm session is nil!")
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
	sq2 := `ALTER TABLE {{.table}} ADD {{.unique}}INDEX {{.index_name}} ({{.indexes}})`
	sq2 = g2util.TextTemplateMustParse(sq2, mp1)

	indexCount, err := c.sn.SQL(sq1).Count()
	if err != nil || indexCount > 0 {
		return
	}

	_, err = c.sn.Exec(sq2)
	return
}

//TableName ...
func TableName(bean interface{}) string { return tableName(bean) }

//FieldName ...
func FieldName(field string) (name string) { return names.LintGonicMapper.Obj2Table(field) }

//tableName CacheTableName ...
//redis-key = /h/tableName
//2020/3/30 22:11 -- Author:charles
//func CacheTableName(table interface{}) string { return tableName(table) }
func tableName(table interface{}) string {
	/*tnBean, ok := table.(names.TableName)
	if ok {
		return tnBean.TableName()
	}
	val1 := reflect.Indirect(reflect.ValueOf(table))
	return names.LintGonicMapper.Obj2Table(val1.Type().Name())*/
	return names.GetTableName(names.LintGonicMapper, reflect.ValueOf(table))
}

//fieldName 获取模型对象字段 => 数据库的字段名
//参与缓存的字段,tag中不能有自定义的 name (字段名)
func fieldName(field string) (name string) { return names.LintGonicMapper.Obj2Table(field) }

//memKey ...
func memKey(b interface{}, key string) string { return fmt.Sprintf("%s::%s", tableName(b), key) }

//MyBase1 id and created
type MyBase1 struct {
	ID      int64            `json:"id,omitempty" xorm:"pk autoincr"`
	Created *g2util.JSONTime `json:"created,omitempty" xorm:"notnull default CURRENT_TIMESTAMP created index comment('创建时间')"`
}

//MyBase xorm MySQL model base
type MyBase struct {
	MyBase1 `xorm:"extends"`
	Updated *g2util.JSONTime `json:"updated,omitempty" xorm:"notnull default CURRENT_TIMESTAMP updated comment('更新时间')"`
	Version int64            `json:"version,omitempty" xorm:"notnull default 1 version comment('乐观锁')"`
}

//ClearColumns ...
func (m *MyBase) ClearColumns() {
	m.Version = 0
	m.Created = nil
	m.Updated = nil
}
