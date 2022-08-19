package g2util

import (
	"os"
	"path/filepath"
	"reflect"
)

// FileAbsPath ...
func FileAbsPath(path string) string { s, _ := filepath.Abs(path); return s }

// Write2file ...
func Write2file(fileName, content string) { _ = os.WriteFile(fileName, []byte(content), 0755) }

// TranslateSlice2Interfaces ...
func TranslateSlice2Interfaces(sl interface{}) []interface{} {
	val := reflect.ValueOf(sl)
	if val.Kind() != reflect.Slice {
		panic("need slice type")
	}
	v2 := make([]interface{}, 0, val.Len())
	for i := 0; i < val.Len(); i++ {
		v2 = append(v2, val.Index(i).Interface())
	}
	return v2
}
