// +build !windows

package main

import (
	"os"
)

func getProxyServer() (proxyServer string) {
	proxyServer = os.Args[1]
	return
}
