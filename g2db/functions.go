package g2db

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-redis/redis/v8"
	"github.com/gookit/goutil/dump"
	"github.com/pkg/errors"
	"xorm.io/xorm"
)

var (
	_ = spew.Dump
	_ = dump.P
)

// initializeTables 初始化数据表
func initializeTables(table ItfInitData, sn *xorm.Session) (err error) {
	tbName := tableName(table)
	beans := table.InitData()
	if len(beans) == 0 {
		return
	}

	//如果数据表不为空,不进行初始化
	//2021/6/22 2:13 上午 -- Author:charles
	count, _err := sn.Count(table)
	if _err != nil {
		return fmt.Errorf("[InitData] [%s] Count error: %w", tbName, _err)
	}
	if count > 0 {
		return
	}

	sq1 := fmt.Sprintf(`truncate table %s;`, tbName)
	_, _ = sn.Exec(sq1)
	if _, err = sn.Insert(beans...); err != nil {
		err = errors.Errorf("[InitData] [%s] error: %s", tbName, err.Error())
		return
	}
	return
}

// HasError ...
func HasError(e error) (bool, error) {
	if e == nil {
		return true, e
	}
	var errorMysqlNotFound ErrorMysqlNotFound
	if errors.As(e, &errorMysqlNotFound) || errors.Is(e, redis.Nil) {
		return false, nil
	}
	return false, e
}
