package g2util

import (
	"log"

	"github.com/facebookgo/inject"
)

// InjectPopulate ...
func InjectPopulate(values ...interface{}) {
	err := inject.Populate(values...)
	if err != nil {
		log.Fatalln(err)
	}
}
