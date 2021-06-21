package g2util

import (
	"path/filepath"
)

//FileAbsPath ...
func FileAbsPath(path string) string { s, _ := filepath.Abs(path); return s }
