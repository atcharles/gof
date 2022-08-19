package g2db

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"xorm.io/xorm"

	"github.com/atcharles/gof/v2/g2util"
)

type (
	//MysqlQueryRowsParams ...
	MysqlQueryRowsParams struct {
		Page        int      `json:"page,omitempty"`
		PageCount   int      `json:"page_count,omitempty"`
		Conditions  []string `json:"conditions,omitempty"`
		OrderBy     string   `json:"order_by,omitempty"`
		Asc         bool     `json:"asc,omitempty"`
		TimeColumn  string   `json:"time_column"`
		TimeBetween string   `json:"time_between"`
	}
	//MysqlRows ...
	MysqlRows struct {
		Pages int         `json:"pages,omitempty"`
		Data  interface{} `json:"data,omitempty"`
		Count int64       `json:"count,omitempty"`
	}

	//ItfMysqlAfterQueryRow ...
	ItfMysqlAfterQueryRow interface {
		MysqlAfterQueryRow()
	}
)

// GetBeanByTableName ...
func (m *Mysql) GetBeanByTableName(tableStr string) (bean interface{}, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, table := range m.tables {
		if tableName(table) == tableStr {
			bean = table
			return
		}
	}
	err = errors.Errorf("数据表%s不存在", tableStr)
	return
}

// QueryTableRows ...查询经过注册的表
func (m *Mysql) QueryTableRows(tableStr string, params *MysqlQueryRowsParams) (rows *MysqlRows, err error) {
	val, err := m.GetBeanByTableName(tableStr)
	if err != nil {
		return
	}
	return m.QueryRows(val, params)
}

// QueryRows ...分页查询,可以指定表名
func (m *Mysql) QueryRows(val interface{}, params *MysqlQueryRowsParams) (rows *MysqlRows, err error) {
	return NewQuery(m.Engine()).QueryRows(val, params)
}

// Query ...
// new(Query).SetDb(*xorm.Engine).QueryRows(val, params)
// new(Query).SetDb(*xorm.Engine).SetTable(string).QueryRows(val, params)
type Query struct {
	db *xorm.Engine

	table string
}

// SetTable ...
func (q *Query) SetTable(table string) *Query { q.table = table; return q }

// QueryRows ...
func (q *Query) QueryRows(val interface{}, params *MysqlQueryRowsParams) (rows *MysqlRows, err error) {
	db := q.db
	v1 := reflect.ValueOf(val)
	sl1 := reflect.MakeSlice(reflect.SliceOf(v1.Type()), 0, 0)
	sl := reflect.New(sl1.Type())
	sl.Elem().Set(sl1)
	data := sl.Interface()

	if params.Page == 0 {
		params.Page = 1
	}
	if params.PageCount == 0 {
		params.PageCount = 10
	}
	const maxCount = 100
	if params.PageCount > maxCount {
		params.PageCount = maxCount
	}
	if len(params.OrderBy) == 0 {
		params.OrderBy = "id"
	}
	if len(params.TimeColumn) == 0 {
		params.TimeColumn = "created"
	}

	sq := `SELECT * FROM {{.table}} WHERE ({{.condition}}) ` +
		`ORDER BY {{.orderBy}} {{.sort}} LIMIT {{.offsetX}},{{.pageCount}}`
	tpl := g2util.Map{
		"orderBy":   db.Quote(params.OrderBy),
		"sort":      "DESC",
		"offsetX":   params.PageCount * (params.Page - 1),
		"pageCount": params.PageCount,
	}
	if params.Asc {
		tpl["sort"] = "ASC"
	}

	tableStr := func() string {
		if len(q.table) == 0 {
			return tableName(val)
		}
		return q.table
	}()
	tpl["table"] = db.Quote(tableStr)

	condition1 := []string{"1=1"}
	condition1 = append(condition1, params.Conditions...)
	if len(params.TimeBetween) > 0 {
		ts := strings.Split(params.TimeBetween, ",")
		if len(ts) == 2 {
			condition1 = append(
				condition1,
				fmt.Sprintf("(`%s` BETWEEN '%s' AND '%s')", params.TimeColumn, ts[0], ts[1]),
			)
		}
	}
	for i, s := range condition1 {
		condition1[i] = fmt.Sprintf("(%s)", s)
	}
	conditionStr := strings.Join(condition1, " AND ")
	tpl["condition"] = conditionStr

	sq = g2util.TextTemplateMustParse(sq, tpl)
	if err = db.SQL(sq).Find(data); err != nil {
		return
	}
	rows = &MysqlRows{Data: data}
	if sl.Elem().Len() == 0 {
		return
	}

	count, err := db.Table(tableStr).Where(conditionStr).Count()
	if err != nil {
		return
	}
	rows.Count = count
	rows.Pages = int(count) / params.PageCount
	if int(count)%params.PageCount > 0 {
		rows.Pages++
	}

	ss1 := sl.Elem()
	for i := 0; i < ss1.Len(); i++ {
		switch vv := ss1.Index(i).Interface().(type) {
		case ItfMysqlAfterQueryRow:
			vv.MysqlAfterQueryRow()
		}
	}
	return
}

// NewQuery ...
func NewQuery(engine *xorm.Engine) *Query { return &Query{db: engine} }
