// +build windows
package ntlm_proxy

import (
	"os"

	log "github.com/sirupsen/logrus"

	"golang.org/x/sys/windows/registry"
)

func getProxyServer() (proxyServer string) {
	// Check Environment
	if proxyFromEnv, ok := os.LookupEnv("GONTLM_PROXY"); ok {
		proxyServer = proxyFromEnv
		return
	}
	// Pull Proxy from the Registry
	k, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.QUERY_VALUE)
	if err != nil {
		log.Fatal(err)
	}
	defer k.Close()

	proxyServer, _, err = k.GetStringValue("ProxyServer")
	if err != nil {
		log.Error(err)
	}
	if proxyServer != "" {
		proxyServer = "http://" + proxyServer
	}
	return
}
