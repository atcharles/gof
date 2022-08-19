package j2rpc

import (
	"regexp"
	"strings"
)

// middleInfo ...中间件信息
type middleInfo struct {
	//作用于方法名, 支持正则表达式
	method []string
	//排除的方法
	excludes []string
	//处理函数:参数顺序: ctx,method,writer,request
	function interface{}
	//所有均经过
	all bool
}

// getMatchFunction ...
func (m *middleInfo) getMatchFunction(method string) interface{} {
	if m.all {
		return m.function
	}
	_f1 := func(_m string) bool {
		for _, s := range m.excludes {
			if methodPathEqual(s, method) {
				return false
			}
		}
		for _, s := range m.method {
			if methodPathEqual(s, method) {
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

func snakePath(str string) string {
	const splitMethodSeparator = "."
	sl1 := strings.Split(str, splitMethodSeparator)
	for i, s := range sl1 {
		sl1[i] = SnakeString(s)
	}
	return strings.Join(sl1, splitMethodSeparator)
}

func methodPathEqual(s, method string) bool {
	if s == method {
		return true
	}
	if snakePath(s) == method {
		return true
	}
	if regexp.MustCompile(s).MatchString(method) {
		return true
	}
	return false
}
