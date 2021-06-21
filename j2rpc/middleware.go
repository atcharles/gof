package j2rpc

import (
	"regexp"
)

//middleInfo ...中间件信息
type middleInfo struct {
	//作用于方法名, 支持正则表达式
	method []string
	//处理函数
	function interface{}
}

//getMatchFunction ...
func (m *middleInfo) getMatchFunction(method string) interface{} {
	_f1 := func(regexpStr, methodStr string) (match bool) {
		rxp, e := regexp.Compile(regexpStr)
		if e != nil {
			return
		}
		return rxp.MatchString(methodStr)
	}

	for _, s := range m.method {
		if s == method || _f1(s, method) {
			return m.function
		}
	}
	return nil
}
