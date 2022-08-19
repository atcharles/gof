package g2util

import (
	"encoding/json"
	"reflect"
)

// MergeBean ...
// @Description: merge struct's point, merge src to dst
// @param dst
// @param src
func MergeBean(dst interface{}, src Map) (err error) {
	dv1 := reflect.ValueOf(dst)
	if dv1.Kind() != reflect.Ptr {
		panic("dst needs pointer kind")
	}
	dstBs, err := json.Marshal(dst)
	if err != nil {
		return
	}
	tmp1 := make(map[string]interface{})
	if err = json.Unmarshal(dstBs, &tmp1); err != nil {
		return
	}
	for k, v := range src {
		tmp1[k] = v
	}
	dstBs, err = json.Marshal(tmp1)
	if err != nil {
		return
	}
	err = json.Unmarshal(dstBs, dst)
	return
}
