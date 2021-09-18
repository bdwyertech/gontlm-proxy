//go:build darwin
// +build darwin

package ntlm_proxy

import (
	"os"

	scutil "github.com/bdwyertech/go-scutil/proxy"
	log "github.com/sirupsen/logrus"
)

func getProxyServer() (proxyServer string) {
	// Check Environment
	if proxyFromEnv, ok := os.LookupEnv("GONTLM_PROXY"); ok {
		proxyServer = proxyFromEnv
		return
	}

	// Pull Proxy from SCUtil
	scutilCfg, err := scutil.Get()
	if err != nil {
		log.Fatal(err)
	}

	if scutilCfg.ProxyAutoConfigEnable == "1" && scutilCfg.ProxyAutoConfigURLString != "" {
		// Ensure we use PAC over Proxy ENV variables in ProxyPlease
		os.Unsetenv("HTTP_PROXY")
		os.Unsetenv("http_proxy")
		os.Unsetenv("HTTPS_PROXY")
		os.Unsetenv("https_proxy")
		log.Infoln("Using Proxy Auto-Configuration (PAC) file:", scutilCfg.ProxyAutoConfigURLString)
		return
	}

	return
}
