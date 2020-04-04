package ntlm_proxy

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/bhendo/concord"
	"github.com/bhendo/concord/handshakers"
	"github.com/elazarl/goproxy"
)

var proxyServer string

func init() {
	flag.StringVar(&proxyServer, "proxy", getEnv("GONTLM_PROXY", getProxyServer()), "Forwarding proxy server")
}

func Run() {
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
	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)

	// TLS Client Configuration
	tlsClientConfig := &tls.Config{}
	if _, disabled := os.LookupEnv("GONTLM_PROXY_NO_SSL_VERIFY"); disabled {
		tlsClientConfig.InsecureSkipVerify = true
	}

	// NTLM Transport
	tr := concord.Transport{
		Proxy:               http.ProxyURL(proxyUrl),
		ProxyAuthorizer:     &handshakers.NTLMProxyAuthorizer{},
		TLSClientConfig:     tlsClientConfig,
		TLSHandshakeTimeout: time.Second * 15,
	}

	// Handle HTTP Connect Requests
	proxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		ctx.RoundTripper = goproxy.RoundTripperFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (resp *http.Response, err error) {
			resp, err = tr.RoundTrip(req)
			return
		})

		return req, nil
	})

	srv := &http.Server{
		Addr:         bind,
		Handler:      proxy,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
	}

	log.Fatal(srv.ListenAndServe())

}
