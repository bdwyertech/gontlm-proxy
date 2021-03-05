package proxyplease

import (
	"log"
	"os"
)

var l = log.New(os.Stdout, "", log.LstdFlags)

var debugf = func(format string, a ...interface{}) {
	l.Printf("proxyplease."+format, a...)
}

// SetDebugf sets a debugf function for debug output
func SetDebugf(f func(format string, a ...interface{})) {
	debugf = f
}
