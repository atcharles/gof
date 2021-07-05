package g2db

import (
	"reflect"
	"strings"

	"github.com/pkg/errors"

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
	}

	//ItfMysqlAfterQueryRow ...
	ItfMysqlAfterQueryRow interface {
		MysqlAfterQueryRow()
	}
)

//QueryTableRows ...
func (m *Mysql) QueryTableRows(tableStr string, params *MysqlQueryRowsParams) (rows *MysqlRows, err error) {
	val := func() interface{} {
		m.mu.RLock()
		defer m.mu.RUnlock()
		for _, table := range m.tables {
			if tableName(table) == tableStr {
				return table
			}
		}
		return nil
	}()
	if val == nil {
		err = errors.Errorf("数据表%s不存在", tableStr)
		return
	}
	return m.QueryRows(val, params)
}

//QueryRows ...
func (m *Mysql) QueryRows(val interface{}, params *MysqlQueryRowsParams) (rows *MysqlRows, err error) {
	v1 := reflect.ValueOf(val)
	sl1 := reflect.MakeSlice(reflect.SliceOf(v1.Type()), 0, 0)
	sl := reflect.New(sl1.Type())
	sl.Elem().Set(sl1)
	data := sl.Interface()

	sq := `select * from {{.table}} where {{.condition}} order by {{.orderBy}} {{.sort}} limit {{.offsetX}},{{.pageCount}}`
	tpl := g2util.Map{
		"orderBy":   "id",
		"sort":      "desc",
		"offsetX":   params.PageCount * (params.Page - 1),
		"pageCount": params.PageCount,
	}
	tpl["table"] = tableName(val)
	condition1 := []string{"1=1"}
	condition1 = append(condition1, params.Conditions...)
	conditionStr := strings.Join(condition1, " AND ")
	tpl["condition"] = conditionStr
	if len(params.OrderBy) > 0 {
		tpl["orderBy"] = params.OrderBy
	}
	if params.Asc {
		tpl["sort"] = "asc"
	}

	db := m.Engine()
	sq = g2util.TextTemplateMustParse(sq, tpl)
	if err = db.SQL(sq).Find(data); err != nil {
		return
	}
	rows = &MysqlRows{Data: data}
	if sl.Elem().Len() == 0 {
		return
	}

	count, err := db.Where(conditionStr).Count(val)
	if err != nil {
		return
	}

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
