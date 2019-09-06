package hoz

import (
	"log"
	"os"
)

var LOG *log.Logger

func init() {
	LOG = log.New(os.Stdout, "hoz>>", log.Ldate|log.Ltime|log.Lshortfile)
}
