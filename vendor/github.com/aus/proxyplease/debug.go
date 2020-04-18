package proxyplease

import (
	log "github.com/sirupsen/logrus"
)

var debugf = func(format string, a ...interface{}) {
	log.Debugf("proxyplease."+format, a...)
}

// SetDebugf sets a debugf function for debug output
func SetDebugf(f func(format string, a ...interface{})) {
	debugf = f
}
