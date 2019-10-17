// +build windows
package main

import (
	"golang.org/x/sys/windows/registry"
	"log"
)

func getProxyServer() (proxyServer string) {
	// Pull Proxy from the Registry
	k, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.QUERY_VALUE)
	if err != nil {
		log.Fatal(err)
	}
	defer k.Close()

	// proxyServer := os.Args[1]
	proxyServer, _, err = k.GetStringValue("ProxyServer")
	if err != nil {
		log.Fatal(err)
	}
	return
}
