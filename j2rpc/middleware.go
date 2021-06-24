package j2rpc

import (
	"regexp"
)

//middleInfo ...中间件信息
type middleInfo struct {
	//作用于方法名, 支持正则表达式
	method []string
	//排除的方法
	excludes []string
	//处理函数
	function interface{}
}

//getMatchFunction ...
func (m *middleInfo) getMatchFunction(method string) interface{} {
	_f1 := func(_m string) bool {
		for _, exclude := range m.excludes {
			if _m == exclude {
				return false
			}
			if regexp.MustCompile(exclude).MatchString(_m) {
				return false
			}
		}
		for _, s := range m.method {
			if s == _m {
				return true
			}
			if regexp.MustCompile(s).MatchString(_m) {
				return true
			}
		}
		return false
	}
	if _f1(method) {
		return m.function
	}
	return nil
}
