package g2db

import (
	"xorm.io/xorm"
)

// ItfSessionBeforeInsert ...
type ItfSessionBeforeInsert interface {
	SessionBeforeInsert(sn *xorm.Session) (err error)
}

// ItfSessionBeforeUpdate ...
type ItfSessionBeforeUpdate interface {
	SessionBeforeUpdate(sn *xorm.Session) (err error)
}

// ItfSessionBeforeDelete ...
type ItfSessionBeforeDelete interface {
	SessionBeforeDelete(sn *xorm.Session) (err error)
}

// ItfSessionAfterInsert ...
type ItfSessionAfterInsert interface {
	SessionAfterInsert(sn *xorm.Session) (err error)
}

// ItfSessionAfterUpdate ...
type ItfSessionAfterUpdate interface {
	SessionAfterUpdate(sn *xorm.Session) (err error)
}

// ItfSessionAfterDelete ...
type ItfSessionAfterDelete interface {
	SessionAfterDelete(sn *xorm.Session) (err error)
}

// ItfAfterSync ...
type ItfAfterSync interface {
	AfterSync(sn *xorm.Session) error
}
