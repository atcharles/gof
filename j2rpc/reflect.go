package j2rpc

import (
	"reflect"
	"runtime"
	"strings"
)

// FuncName ...
func FuncName(fn interface{}) string {
	name := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	sl := strings.Split(name, ".")
	return strings.Replace(sl[len(sl)-1], "-fm", "", -1)
}
