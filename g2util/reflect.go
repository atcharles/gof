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
	v := ValueIndirect(reflect.ValueOf(bean))
	if v.IsZero() {
		panic("need not zero value")
	}
	return reflect.New(v.Type()).Interface()
}

//ObjectTagInstances ...
/**
 * @Description:根据标签获取字段实例集合
 * @param obj
 * @param tagName
 * @return []interface{}
 */
func ObjectTagInstances(obj interface{}, tagName string) []interface{} {
	data := make([]interface{}, 0)
	_f1append := func(vv reflect.Value, vf reflect.StructField) {
		_, has := vf.Tag.Lookup(tagName)
		if !has {
			return
		}
		if !(vv.CanSet() && vv.CanAddr() && !vv.IsNil() && vv.Kind() == reflect.Ptr) {
			return
		}
		data = append(data, vv.Interface())
	}
	tv1 := ValueIndirect(reflect.ValueOf(obj))
	for i := 0; i < tv1.NumField(); i++ {
		_f1append(tv1.Field(i), tv1.Type().Field(i))
	}
	return data
}
