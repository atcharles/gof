package g2gin

import (
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
)

// JSONResponse ...
type JSONResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

// JSONP ...
func JSONP(c *gin.Context, code int, val ...interface{}) {
	var msg string
	var data interface{}
	if len(val) > 0 {
		_fn1 := func() {
			if code != 200 {
				msg = cast.ToString(val[0])
				return
			}
			data = val[0]
		}
		_fn1()
	}
	c.JSONP(code, &JSONResponse{Code: code, Msg: msg, Data: data})
}

var (
	ginContextType = reflect.TypeOf((*gin.Context)(nil))
	errorType      = reflect.TypeOf((*error)(nil)).Elem()
)

// GinHandler ...
func GinHandler(fn interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		val, err := parseArgumentFromContext(fn, c)
		if err != nil {
			_ = c.Error(err)
			JSONP(c, 400, err.Error())
			return
		}
		JSONP(c, 200, val)
	}
}

func parseArgumentFromContext(fn interface{}, c *gin.Context) (val interface{}, err error) {
	fnv := reflect.ValueOf(fn)
	if fnv.Kind() != reflect.Func {
		return nil, errors.Errorf("need func type")
	}
	numIn := fnv.Type().NumIn()
	if numIn > 2 {
		return nil, errors.Errorf("too many arguments, want at most 2")
	}
	args := make([]reflect.Value, 0)
	for i := 0; i < fnv.Type().NumIn(); i++ {
		arg1 := fnv.Type().In(i)
		if arg1.AssignableTo(ginContextType) {
			args = append(args, reflect.ValueOf(c))
			continue
		}
		switch arg1.Kind() {
		case reflect.Ptr:
			agv := reflect.New(arg1.Elem())
			if err = c.ShouldBind(agv.Interface()); err != nil {
				return
			}
			args = append(args, agv)
		case reflect.Struct, reflect.Map, reflect.Slice:
			agv := reflect.New(arg1)
			if err = c.ShouldBind(agv.Interface()); err != nil {
				return
			}
			args = append(args, agv.Elem())
		}
	}
	results := fnv.Call(args)
	switch len(results) {
	case 1:
		if results[0].Type().Implements(errorType) {
			if err = value2err(results[0]); err != nil {
				return
			}
			val = "OK"
			return
		}
		val = results[0].Interface()
	case 2:
		if err = value2err(results[1]); err != nil {
			return
		}
		val = results[0].Interface()
	default:
		val = "OK"
	}
	return
}

func value2err(val reflect.Value) error {
	if !val.Type().Implements(errorType) {
		return nil
	}
	if !val.IsValid() {
		return nil
	}
	if val.IsZero() {
		return nil
	}
	vv, ok := val.Interface().(error)
	if !ok {
		return nil
	}
	return vv
}
