package ntlm_proxy

import (
	"crypto/tls"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"

	"github.com/bhendo/concord"
	"github.com/bhendo/concord/handshakers"
	"github.com/elazarl/goproxy"
)

func Run() {
	proxyServer := getEnv("GONTLM_PROXY", getProxyServer())
	bind := getEnv("GONTLM_BIND", ":3128")
	log.Printf("INFO: Forwarding Proxy is: %s\n", proxyServer)
	log.Printf("INFO: Listening on: %s\n", bind)

	proxyUrl, err := url.Parse("http://" + proxyServer)
	if err != nil {
		log.Fatal(err)
	}
	setGoProxyCA()
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = false
	if _, enabled := os.LookupEnv("GONTLM_PROXY_VERBOSE"); enabled {
		proxy.Verbose = true
	}

	var AlwaysMitmAuth goproxy.FuncHttpsHandler = func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		return goproxy.MitmConnect, host
	}

	// Handle HTTPS Connect Requests
	proxy.OnRequest(goproxy.ReqHostMatches(regexp.MustCompile("^.*"))).HandleConnect(AlwaysMitmAuth)

	// TLS Client Configuration
	tlsClientConfig := &tls.Config{}
	if _, disabled := os.LookupEnv("GONTLM_PROXY_NO_SSL_VERIFY"); disabled {
		tlsClientConfig.InsecureSkipVerify = true
	}

	// NTLM Transport
	tr := concord.Transport{
		Proxy:           http.ProxyURL(proxyUrl),
		ProxyAuthorizer: &handshakers.NTLMProxyAuthorizer{},
		TLSClientConfig: tlsClientConfig,
	}

	// Handle HTTP Connect Requests
	proxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		ctx.RoundTripper = goproxy.RoundTripperFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (resp *http.Response, err error) {
			resp, err = tr.RoundTrip(req)
			return
		})

		return req, nil
	})

	log.Fatal(http.ListenAndServe(bind, proxy))

}
