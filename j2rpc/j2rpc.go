package j2rpc

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/atcharles/gof/v2/g2util"
	"github.com/atcharles/gof/v2/json"
)

const (
	vsn = "2.0"

	splitMethodSeparator = "."
)

var (
	contextType = reflect.TypeOf((*context.Context)(nil)).Elem()
	errorType   = reflect.TypeOf((*error)(nil)).Elem()

	_ = New()
)

// New ...
func New(opts ...*Option) RPCServer {
	s := &server{
		run:    1,
		logger: g2util.NewLevelLogger("[STDOUT]"),
	}
	if len(opts) > 0 && opts[0] != nil {
		s.opt = opts[0]
	}
	if s.opt == nil {
		s.opt = SnakeOption
	}
	return s
}

type (
	server struct {
		mutex    sync.Mutex
		opt      *Option
		run      int32
		services map[string]service
		logger   g2util.LevelLogger

		excludeMethods []string
	}
	service struct {
		name      string
		receiver  reflect.Value
		callbacks map[string]callback
	}
)

// Opt ...
func (s *server) Opt() *Option { return s.opt }

// Stop stops reading new requests, waits for stopPendingRequestTimeout to allow pending
// requests to finish, then closes all codecs which will cancel pending requests and
// subscriptions.
func (s *server) Stop() {
	if atomic.CompareAndSwapInt32(&s.run, 1, 0) {
		s.logger.Debugf("RPC server shutting down")
	}
}

func (s *server) Logger() g2util.LevelLogger { return s.logger }

func (s *server) SetLogger(logger g2util.LevelLogger) { s.logger = logger }

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Permit dumb empty requests for remote health-checks (AWS)
	if r.Method == http.MethodGet && r.ContentLength == 0 && r.URL.RawQuery == "" {
		w.WriteHeader(http.StatusOK)
		return
	}
	s.Handler(r.Context(), w, r)
}

// RegisterForApp ...
func (s *server) RegisterForApp(app interface{}) {
	if _app, ok := app.(ItfExcludeMethod); ok {
		s.excludeMethods = _app.ExcludeMethod()
	}
	namespaces := g2util.ObjectTagInstances(app, "j2rpc")
	for _, namespace := range namespaces {
		s.Register(namespace)
	}
}

// Register ...
func (s *server) Register(receiver interface{}, names ...string) {
	var _fnGetServiceName = func(rv interface{}) string {
		rvv := g2util.ValueIndirect(reflect.ValueOf(rv))
		var name string
		if len(names) > 0 && len(names[0]) > 0 {
			name = names[0]
		}
		if nsn1, ok := rv.(ItfNamespaceName); ok {
			name = nsn1.J2rpcNamespaceName()
		}
		if len(name) == 0 {
			name = s.formatName(rvv.Type().Name())
		}
		return name
	}
	serviceName := _fnGetServiceName(receiver)

	if consVal, ok := receiver.(ItfConstructor); ok {
		consVal.Constructor()
	}

	callbacks := s.suitableCallbacks(receiver)
	if len(callbacks) == 0 {
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.services == nil {
		s.services = make(map[string]service)
	}

	srv, ok := s.services[serviceName]
	if !ok {
		srv = service{
			name:      serviceName,
			receiver:  reflect.ValueOf(receiver),
			callbacks: make(map[string]callback),
		}
		s.services[serviceName] = srv
	} else {
		panic(fmt.Sprintf("namespace [%s] exists", serviceName))
	}

	for name, cb := range callbacks {
		srv.callbacks[name] = cb
	}
}

// suitableCallbacks ...
func (s *server) suitableCallbacks(receiver interface{}) (callbacks map[string]callback) {
	callbacks = make(map[string]callback)

	var skipMethods = append([]string{"Constructor", "ExcludeMethod"}, s.excludeMethods...)
	if exv, ok := receiver.(ItfExcludeMethod); ok {
		skipMethods = append(skipMethods, exv.ExcludeMethod()...)
	}
	var _fn1InSkips = func(m1 string) bool {
		for _, method := range skipMethods {
			if m1 == method {
				return true
			}
		}
		return false
	}

	var _fnAppendMethods = func(method reflect.Method) {
		if method.PkgPath != "" {
			return
		}

		if _fn1InSkips(method.Name) {
			return
		}

		c := callback{server: s, methodName: method.Name, rcv: reflect.ValueOf(receiver), fn: method.Func, errPos: -1}
		if ok := c.makeArgTypes(); !ok {
			return
		}
		callbacks[s.formatName(method.Name)] = c
	}

	rp := reflect.TypeOf(receiver)
	for i := 0; i < rp.NumMethod(); i++ {
		_fnAppendMethods(rp.Method(i))
	}
	return
}

// formatName ...
func (s *server) formatName(name string) string {
	if s.opt.SnakeNamespace {
		name = SnakeString(name)
	}
	return name
}

// Handler ...
func (s *server) Handler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	// Don't serve if server is stopped.
	if atomic.LoadInt32(&s.run) == 0 {
		return
	}
	//检测是否已经写入header
	if len(w.Header().Get("Status-Written")) != 0 {
		return
	}
	if code, err := validateRequest(r); err != nil {
		http.Error(w, err.Error(), code)
		return
	}
	// Prevents Internet Explorer from MIME-sniffing a response away
	// from the declared content-type
	w.Header().Set("x-content-type-options", "nosniff")

	msg := new(RPCMessage)
	if err := s.handle(ctx, w, r, msg); err != nil {
		msg = msg.setError(err)
	}
	msg.output().writeResponse(w)
	requestID := w.Header().Get("request-id")
	if len(requestID) > 0 {
		s.logger.Debugf("[Request-ID:%s] %s", requestID, g2util.JSONDump(msg))
	}
}

// handle ...
func (s *server) handle(ctx context.Context, w http.ResponseWriter, r *http.Request, msg *RPCMessage) (err error) {
	err = json.NewDecoder(r.Body).Decode(msg)
	_ = r.Body.Close()
	if err != nil {
		err = NewError(ErrParse, err.Error())
		return
	}

	if !msg.hasValidID() {
		err = NewError(ErrInvalidRequest, "id is invalid")
		return
	}
	elem, err := msg.methods()
	if err != nil {
		return
	}
	for i, e2 := range elem {
		//elem[i] = s.formatName(CamelString(e2))
		elem[i] = s.formatName(e2)
	}
	msg.Method = strings.Join(elem, splitMethodSeparator)

	cbk, err := s.getCallBack(elem)
	if err != nil {
		return
	}

	if err = s.opt.beforeMiddlewareAction(ctx, msg.Method, w, r); err != nil {
		return
	}

	//Catch panic while running the callback.
	defer func() {
		if p := recover(); p != nil {
			err = s.stack(p, cbk.methodName)
			return
		}
	}()

	callArgs, err := parsePositionalArguments(msg.Params, cbk.argTypes)
	if err != nil {
		err = NewError(ErrBadParams, err.Error())
		return
	}

	res, err := cbk.call(ctx, callArgs)
	if err != nil {
		return
	}
	val := reflect.ValueOf(res)
	if !val.IsValid() {
		return
	}
	if val.IsZero() {
		return
	}
	answer, err := json.Marshal(res)
	if err != nil {
		return
	}
	msg.Result = answer
	return
}

// stack ...
func (s *server) stack(recover interface{}, methodName string) error {
	msg := fmt.Sprintf("RPC method %s handler crashed: %v", methodName, recover)
	const size = 64 << 10
	buf := make([]byte, size)
	buf = buf[:runtime.Stack(buf, false)]
	s.logger.Errorf("%s\n%s", msg, buf)
	return NewError(ErrInternal, msg)
}

// getCallBack ...
func (s *server) getCallBack(elem []string) (cbk callback, err error) {
	svs, ok := s.services[elem[0]]
	if !ok {
		err = NewError(ErrNoMethod, "no namespace")
		return
	}

	cbk, ok = svs.callbacks[elem[1]]
	if !ok {
		err = NewError(ErrNoMethod, "no method")
		return
	}
	return
}
