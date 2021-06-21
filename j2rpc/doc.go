package j2rpc

/**
example1:

type aa struct{}

//API1 ...
func (a *aa) API1() (data interface{}) { return "hello world" }

//API2 ...
func (a *aa) API2(c context.Context) (data interface{}) { return reflect.TypeOf(c).String() }

//API3 ...
func (a *aa) API3(val string) (data interface{}) { return val }

func main() {
	opt := j2rpc.SnakeOption
	opt.AddBeforeMiddleware(
		[]string{`aa.api1`, `^aa\.\S+[1]$`},
		func(c context.Context, method string, w http.ResponseWriter, r *http.Request) (err error) {
			spew.Dump(reflect.TypeOf(c).Elem().Name())
			j2rpc.AbortWriteHeader(w, 401)
			return j2rpc.NewError(j2rpc.ErrServer, method)
		},
	)

	rpc1server := j2rpc.New(opt)
	rpc1server.Register(new(aa))

	s := http.NewServeMux()
	s.Handle("/jsonrpc", rpc1server)
	go func() { log.Fatalln(http.ListenAndServe(":301", s)) }()

	s1 := gin.Default()
	s1.Any("/jsonrpc", func(c *gin.Context) { rpc1server.Handler(c, c.Writer, c.Request) })
	log.Fatalln(s1.Run(":300"))
}

*/
