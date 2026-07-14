//go:build !darwin && !windows
// +build !darwin,!windows

package ntlm_proxy

var PacFileURL string

func getProxyServer() (proxyServer string) {
	proxyServer = getEnv("GONTLM_PROXY", "")
	return
}
