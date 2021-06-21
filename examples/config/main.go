package main

import (
	"os"
	"path/filepath"

	"github.com/davecgh/go-spew/spew"
)

func main() {
	dir()
}

func dir() {
	spew.Dump(filepath.Abs(""))
	spew.Dump(os.Getwd())
}
