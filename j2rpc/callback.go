package j2rpc

import (
	"context"
	"reflect"
)

type callback struct {
	server     *server
	methodName string
	//receiver object of method, set if fn is method
	rcv reflect.Value
	//the function
	fn reflect.Value
	//input argument types
	argTypes []reflect.Type
	//method's first argument is a context (not included in argTypes)
	hasCtx bool
	//err return idx, of -1 when method cannot return error
	errPos int
}

// makeArgTypes ...
func (c *callback) makeArgTypes() bool {
	fnt := c.fn.Type()

	outs := make([]reflect.Type, fnt.NumOut())
	for i := 0; i < fnt.NumOut(); i++ {
		outs[i] = fnt.Out(i)
	}
	//A maximum of two values can be returned.
	if len(outs) > 2 {
		return false
	}
	//If an error is returned, it must be the last returned value.
	switch {
	case len(outs) == 1 && isErrorType(outs[0]):
		c.errPos = 0
	case len(outs) == 2:
		if isErrorType(outs[0]) || !isErrorType(outs[1]) {
			return false
		}
		c.errPos = 1
	}

	firstArg := 0
	if c.rcv.IsValid() {
		firstArg++
	}

	if fnt.NumIn() > firstArg && fnt.In(firstArg).Implements(contextType) {
		c.hasCtx = true
		firstArg++
	}
	//Add all remaining parameters.
	c.argTypes = make([]reflect.Type, fnt.NumIn()-firstArg)
	for i := firstArg; i < fnt.NumIn(); i++ {
		c.argTypes[i-firstArg] = fnt.In(i)
	}
	return true
}

// call invokes the callback.
func (c *callback) call(ctx context.Context, args []reflect.Value) (res interface{}, err error) {
	//Create the argument slice.
	fullArgs := make([]reflect.Value, 0, 2+len(args))
	if c.rcv.IsValid() {
		fullArgs = append(fullArgs, c.rcv)
	}
	if c.hasCtx {
		fullArgs = append(fullArgs, reflect.ValueOf(ctx))
	}
	fullArgs = append(fullArgs, args...)

	//Catch panic while running the callback.
	defer func() {
		if p := recover(); p != nil {
			err = c.server.stack(p, c.methodName)
			return
		}
	}()

	//Run the callback.
	results := c.fn.Call(fullArgs)
	if len(results) == 0 {
		return
	}
	if c.errPos >= 0 {
		err = value2err(results[c.errPos])
		if err != nil {
			return
		}
	}
	rv := results[0]
	if !rv.IsValid() {
		return
	}
	return rv.Interface(), err
}

func value2err(val reflect.Value) error {
	if !isErrorType(val.Type()) {
		return nil
	}
	if !val.IsValid() {
		return nil
	}
	if val.IsZero() {
		return nil
	}
	return val.Interface().(error)
}
