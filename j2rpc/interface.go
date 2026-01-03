package j2rpc

import (
	"context"
	"net/http"
	"reflect"

	"github.com/atcharles/gof/v2/g2util"
)

type (
	//ItfConstructor ...
	ItfConstructor interface{ Constructor() }
	//ItfExcludeMethod ...
	ItfExcludeMethod interface{ ExcludeMethod() []string }
	//RPCServer ...
	RPCServer interface {
		Opt() *Option
		Logger() g2util.LevelLogger
		SetLogger(logger g2util.LevelLogger)
		ServeHTTP(w http.ResponseWriter, r *http.Request)
		RegisterForApp(app interface{})
		Register(receiver interface{}, names ...string)
		Handler(ctx context.Context, w http.ResponseWriter, r *http.Request)
		Stop()
	}
	//ItfNamespaceName ...
	ItfNamespaceName interface {
		J2rpcNamespaceName() string
	}
)

// PopulateConstructor ...
func PopulateConstructor(value interface{}) {
	vp := reflect.ValueOf(value)
	if vp.Kind() != reflect.Ptr {
		panic("need pointer")
	}

	for vp.Kind() == reflect.Ptr {
		vp = vp.Elem()
	}

	for i := 0; i < vp.NumField(); i++ {
		fv1 := vp.Field(i)
		if fv1.Kind() != reflect.Ptr {
			continue
		}
		if fv1v, ok := fv1.Interface().(ItfConstructor); ok {
			fv1v.Constructor()
		}
	}
}
