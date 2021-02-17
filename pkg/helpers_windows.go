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

	// Check if Proxy is enabled
	proxyEnable, _, err := k.GetIntegerValue("ProxyEnable")
	if err != nil {
		log.Errorln("Could not retrieve value for registry key ProxyEnable:", err)
	}

	if proxyEnable == 0 {
		// Check for PAC file
		pacFile, _, err := k.GetStringValue("AutoConfigURL")
		if err != nil {
			log.Errorln("Could not retrieve value for registry key AutoConfigURL:", err)
		}
		if pacFile == "" {
			log.Warn("No PAC file detected and Proxy is not enabled in Internet Settings")
			return
		} else {
			// Ensure we use PAC over Proxy ENV variables in ProxyPlease
			os.Unsetenv("HTTP_PROXY")
			os.Unsetenv("HTTPS_PROXY")
			log.Infoln("Using Proxy Auto-Configuration (PAC) file:", pacFile)
		}
	}

	if proxyEnable == 1 {
		proxyServer, _, err = k.GetStringValue("ProxyServer")
		if err != nil {
			log.Error(err)
		}
		if proxyServer != "" {
			proxyServer = "http://" + proxyServer
		}
	}

	return
}
