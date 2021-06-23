package g2util

import (
	"github.com/henrylee2cn/goutil"

	"github.com/atcharles/gof/v2/json"
)

//JSONDump ...
func JSONDump(val interface{}) string {
	bts, _ := json.Marshal(val)
	return goutil.BytesToString(bts)
}
