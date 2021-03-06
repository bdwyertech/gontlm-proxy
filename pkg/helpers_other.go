// +build !darwin,!windows

package ntlm_proxy

func getProxyServer() (proxyServer string) {
	proxyServer = getEnv("GONTLM_PROXY", "")
	return
}
