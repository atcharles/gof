package g2util

import (
	"strconv"

	"github.com/atcharles/gof/v2/json"

	"github.com/unknwon/com"
)

//Map ...
type Map map[string]interface{}

//MergeTo ...
func (m Map) MergeTo(mp Map) {
	for k, v := range m {
		mp[k] = v
	}
}

//Bean2Map ...
func Bean2Map(bean interface{}) (Map, error) {
	b1, err := json.Marshal(bean)
	if err != nil {
		return nil, err
	}
	mp := make(Map)
	if err = json.Unmarshal(b1, &mp); err != nil {
		return nil, err
	}
	return mp, nil
}

//ToBean ...
func (m Map) ToBean(bean interface{}) (err error) {
	bs1, err := json.Marshal(m)
	if err != nil {
		return
	}
	return json.Unmarshal(bs1, bean)
}

//UnmarshalBinary ...
func (m *Map) UnmarshalBinary(data []byte) error { return json.Unmarshal(data, m) }

//MarshalBinary ...
func (m Map) MarshalBinary() (data []byte, err error) { return json.Marshal(m) }

//GetString ...
func (m Map) GetString(key string) string {
	val, ok := m[key]
	if !ok {
		return ""
	}
	return com.ToStr(val)
}

//GetInt64 ...
func (m Map) GetInt64(key string) int64 {
	val, ok := m[key]
	if !ok {
		return 0
	}
	return com.StrTo(com.ToStr(val)).MustInt64()
}

//GetInt ...
func (m Map) GetInt(key string) int {
	val, ok := m[key]
	if !ok {
		return 0
	}
	return com.StrTo(com.ToStr(val)).MustInt()
}

//GetBool ...
func (m Map) GetBool(key string) bool {
	val, ok := m[key]
	if !ok {
		return false
	}
	a, _ := strconv.ParseBool(com.ToStr(val))
	return a
}
