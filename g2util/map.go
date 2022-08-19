package g2util

import (
	"strconv"

	"github.com/atcharles/gof/v2/json"

	"github.com/unknwon/com"
)

// MergeBeans ...
// 合并结构数据
func MergeBeans(dst interface{}, src interface{}) (err error) {
	mpDst, err := Bean2Map(src)
	if err != nil {
		return
	}
	err = mpDst.Merge2Bean(dst)
	return
}

// Bean2Map ...
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

//CopyBean ...
/**
 * @Description: copy原始数据到新数据
 * @param bean
 * @return newBean
 * @return err
 */
func CopyBean(bean interface{}) (newBean interface{}, err error) {
	mp1, err := Bean2Map(bean)
	if err != nil {
		return
	}
	newBean = NewValue(bean)
	err = mp1.ToBean(newBean)
	return
}

// MapString ...
type MapString map[string]string

// TransMapStringToMap ...
func (m MapString) TransMapStringToMap() Map {
	m1 := make(Map)
	for k, v := range m {
		m1[k] = v
	}
	return m1
}

// Map ...
type Map map[string]interface{}

// Merge2Bean ...
func (m Map) Merge2Bean(bean interface{}) (err error) {
	mp, err := Bean2Map(bean)
	if err != nil {
		return
	}
	m.MergeTo(mp)
	err = mp.ToBean(bean)
	return
}

// MergeTo ...
func (m Map) MergeTo(mp Map) {
	for k, v := range m {
		mp[k] = v
	}
}

// ToBean ...
func (m Map) ToBean(bean interface{}) (err error) {
	bs1, err := json.Marshal(m)
	if err != nil {
		return
	}
	return json.Unmarshal(bs1, bean)
}

// UnmarshalBinary ...
func (m *Map) UnmarshalBinary(data []byte) error { return json.Unmarshal(data, m) }

// MarshalBinary ...
func (m Map) MarshalBinary() (data []byte, err error) { return json.Marshal(m) }

// GetString ...
func (m Map) GetString(key string) string {
	val, ok := m[key]
	if !ok {
		return ""
	}
	return com.ToStr(val)
}

// GetInt64 ...
func (m Map) GetInt64(key string) int64 {
	val, ok := m[key]
	if !ok {
		return 0
	}
	return com.StrTo(com.ToStr(val)).MustInt64()
}

// GetInt ...
func (m Map) GetInt(key string) int {
	val, ok := m[key]
	if !ok {
		return 0
	}
	return com.StrTo(com.ToStr(val)).MustInt()
}

// GetBool ...
func (m Map) GetBool(key string) bool {
	val, ok := m[key]
	if !ok {
		return false
	}
	a, _ := strconv.ParseBool(com.ToStr(val))
	return a
}

// Keys ...
func (m Map) Keys() []string {
	list := make([]string, 0)
	for k := range m {
		list = append(list, k)
	}
	return list
}
