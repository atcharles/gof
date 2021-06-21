package g2util

import (
	"reflect"
)

//ValueIndirect ...值类型
func ValueIndirect(val reflect.Value) reflect.Value {
	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	return val
}

//NewValue ...
func NewValue(bean interface{}) (val interface{}) {
	return reflect.New(ValueIndirect(reflect.ValueOf(bean)).Type()).Interface()
}
