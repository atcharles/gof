package j2rpc

import (
	"reflect"
)

// SnakeOption ...
var SnakeOption = &Option{SnakeNamespace: true}

// Option ...
type Option struct {
	SnakeNamespace bool
	BeforeMid      []middleInfo
}

//AddBeforeMiddleware ...
/**
 * @Description:
 * @receiver o
 * @param method
 * @param fn: //参数顺序: ctx,method,writer,request
 */
func (o *Option) AddBeforeMiddleware(fn interface{}, ls ...[]string) {
	var (
		all      bool
		method   []string
		excludes []string
	)
	switch len(ls) {
	case 0:
		all = true
	case 1:
		method = ls[0]
	default:
		method = ls[0]
		if len(ls) > 2 {
			//第一个参数是包含的路径
			//第二个参数是排除的路径
			//从第三个参数开始,后面的全部都是排除的路径
			for i := 2; i < len(ls); i++ {
				ls[1] = append(ls[1], ls[i]...)
			}
		}
		excludes = ls[1]
	}
	info := middleInfo{
		method:   method,
		excludes: excludes,
		function: fn,
		all:      all,
	}
	o.BeforeMid = append(o.BeforeMid, info)
}

// beforeMiddlewareFuncAction ...
func (o *Option) beforeMiddlewareFuncAction(fn interface{}, args ...interface{}) (err error) {
	if fn == nil {
		return
	}

	fn1 := reflect.ValueOf(fn)
	if fn1.Kind() != reflect.Func {
		return
	}

	callArgs := make([]reflect.Value, 0, len(args))
	for _, arg := range args {
		callArgs = append(callArgs, reflect.ValueOf(arg))
	}

	tt1 := fn1.Type()
	callArgs = callArgs[:tt1.NumIn()]
	for i, arg := range callArgs {
		ctp := tt1.In(i)
		if !arg.Type().ConvertibleTo(ctp) {
			addVal := reflect.Zero(ctp)
			if ctp.Kind() == reflect.Ptr {
				addVal = reflect.New(ctp).Elem()
			}
			callArgs[i] = addVal
		}
	}
	rst := fn1.Call(callArgs)
	if len(rst) == 0 {
		return
	}
	er := rst[0]
	if er.IsNil() {
		return
	}
	//只能返回error
	if isErrorType(er.Type()) {
		err, _ = er.Interface().(error)
	}
	return
}

// beforeMiddlewareAction ...
func (o *Option) beforeMiddlewareAction(args ...interface{}) (err error) {
	if len(o.BeforeMid) == 0 {
		return
	}

	method := args[1].(string)
	for _, info := range o.BeforeMid {
		if fn := info.getMatchFunction(method); fn != nil {
			if err = o.beforeMiddlewareFuncAction(fn, args...); err != nil {
				return
			}
		}
	}
	return
}
