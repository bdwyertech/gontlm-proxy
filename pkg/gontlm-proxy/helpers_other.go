// +build !windows

package ntlm_proxy

import (
	"os"
)

func getProxyServer() (proxyServer string) {
	proxyServer = os.Args[1]
	return
}
