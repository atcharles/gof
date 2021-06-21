package g2db

import (
	"fmt"

	"github.com/pkg/errors"
	"xorm.io/xorm"
)

//initializeTables 初始化数据表
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
		return _err
	}
	if count > 0 {
		return
	}

	sq1 := fmt.Sprintf(`truncate table %s;`, tbName)
	if _, err = sn.Exec(sq1); err != nil {
		return
	}
	if _, err = sn.Insert(beans...); err != nil {
		err = errors.Errorf("[InitData] [%s] error: %s", tbName, err.Error())
		return
	}
	return
}
