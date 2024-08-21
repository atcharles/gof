package g2db

import (
	"github.com/pkg/errors"
	"xorm.io/xorm"

	"github.com/atcharles/gof/v2/g2util"
)

// newSession ...
func newSession(mysql *Mysql, sn *xorm.Session) *Session {
	return &Session{mysql: mysql, sn: sn}
}

// Session ...
type Session struct {
	mysql *Mysql
	sn    *xorm.Session
}

// Delete ...
func (s *Session) Delete(bean interface{}) (err error) {
	queryList := new(cacheBind).Values(bean)
	if len(queryList) == 0 {
		return
	}
	if err = s.mysql.CacheGet(bean); err != nil {
		return
	}
	if v, ok := bean.(ItfSessionBeforeDelete); ok {
		if err = v.SessionBeforeDelete(s.sn); err != nil {
			return
		}
	}
	if _, err = s.sn.NoAutoCondition().Where(queryList[0]).Delete(bean); err != nil {
		return
	}
	if err = s.mysql.DelCache(bean); err != nil {
		return
	}
	if v, ok := bean.(ItfSessionAfterDelete); ok {
		if err = v.SessionAfterDelete(s.sn); err != nil {
			return
		}
	}
	return
}

// Insert ...
func (s *Session) Insert(bean interface{}) (err error) {
	if v1, ok := bean.(ItfSessionBeforeInsert); ok {
		if err = v1.SessionBeforeInsert(s.sn); err != nil {
			return
		}
	}
	a, err := s.sn.InsertOne(bean)
	if err != nil {
		return
	}
	if a == 0 {
		return errors.New("数据写入失败")
	}
	if err = s.mysql.DelCache(bean); err != nil {
		return
	}
	if v1, ok := bean.(ItfSessionAfterInsert); ok {
		if err = v1.SessionAfterInsert(s.sn); err != nil {
			return
		}
	}
	return
}

// Update ...
func (s *Session) Update(bean interface{}, params ...interface{}) (newBean interface{}, err error) {
	newBean, err = g2util.CopyBean(bean)
	if err != nil {
		return
	}
	if err = s.mysql.CacheGet(newBean); err != nil {
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
	if v, ok := newBean.(ItfSessionBeforeUpdate); ok {
		if err = v.SessionBeforeUpdate(s.sn); err != nil {
			return
		}
	}
	a, err := _f1(s.sn).Update(newBean)
	if err != nil {
		return
	}
	if a == 0 {
		return newBean, errors.New("未更新数据")
	}
	//更新,删除新条件缓存
	if err = s.mysql.DelCache(newBean); err != nil {
		return
	}
	if v, ok := newBean.(ItfSessionAfterUpdate); ok {
		if err = v.SessionAfterUpdate(s.sn); err != nil {
			return
		}
	}
	if queryAfter {
		_, err = s.sn.NoAutoCondition().Where(queryList[0]).Get(newBean)
		if err != nil {
			return
		}
	}
	return
}
