package j2rpc

import (
	"reflect"
)

//SnakeOption ...
var SnakeOption = &Option{SnakeNamespace: true}

//Option ...
type Option struct {
	SnakeNamespace bool

	BeforeMid []middleInfo
}

//AddBeforeMiddleware ...
func (o *Option) AddBeforeMiddleware(method []string, fn interface{}) {
	o.BeforeMid = append(o.BeforeMid, middleInfo{method: method, function: fn})
}

//BeforeMiddlewareAction ...
//参数顺序: ctx,method,writer,request
func (o *Option) BeforeMiddlewareAction(args ...interface{}) (err error) {
	if len(o.BeforeMid) == 0 {
		return
	}

	method := args[1].(string)
	_f1 := func() interface{} {
		for _, info := range o.BeforeMid {
			if fn := info.getMatchFunction(method); fn != nil {
				return fn
			}
		}
		return nil
	}

	func1 := _f1()
	if func1 == nil {
		return
	}

	fn1 := reflect.ValueOf(func1)
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
	//只能返回error
	if isErrorType(rst[0].Type()) {
		err = rst[0].Interface().(error)
	}
	return
}
