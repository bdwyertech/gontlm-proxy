package ntlm_proxy

import (
	"context"
	"flag"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/aus/proxyplease"
	"github.com/elazarl/goproxy"
	// "github.com/bhendo/concord"
	// "github.com/bhendo/concord/handshakers"
)

var proxyBind string
var proxyServer string
var proxyVerbose bool

func init() {
	flag.StringVar(&proxyBind, "bind", getEnv("GONTLM_BIND", "0.0.0.0:3128"), "IP & Port to bind to")
	flag.StringVar(&proxyServer, "proxy", getEnv("GONTLM_PROXY", getProxyServer()), "Forwarding proxy server")
	flag.BoolVar(&proxyVerbose, "verbose", false, "Enable verbose logging")
}

func Run() {
	log.Infof("Forwarding Proxy is: %s", proxyServer)
	log.Infof("Listening on: %s", proxyBind)

	proxyUrl, err := url.Parse("http://" + proxyServer)
	if err != nil {
		log.Fatal(err)
	}

	goproxy.GoproxyCa = SetupGoProxyCA()
	proxy := goproxy.NewProxyHttpServer()

	//
	// Log Configuration
	//
	if _, verbose := os.LookupEnv("GONTLM_PROXY_VERBOSE"); log.IsLevelEnabled(log.DebugLevel) || proxyVerbose || verbose {
		if !log.IsLevelEnabled(log.DebugLevel) {
			log.SetLevel(log.DebugLevel)
		}
		proxy.Verbose = true
	}
	// Override ProxyPlease Logger
	proxyplease.SetDebugf(func(section string, msgs ...interface{}) {
		log.Debugf("proxyplease."+section, msgs...)
	})

	//
	// Dial Context
	//
	dialContext := proxyplease.NewDialContext(proxyplease.Proxy{URL: proxyUrl})
	proxy.ConnectDial = func(network, addr string) (net.Conn, error) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		return dialContext(ctx, network, addr)
	}
	proxy.Tr.Proxy = nil
	proxy.Tr.DialContext = dialContext

	// Connect Handler
	var AlwaysMitm goproxy.FuncHttpsHandler = func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		MitmConnect := &goproxy.ConnectAction{
			Action:    goproxy.ConnectMitm,
			TLSConfig: goproxy.TLSConfigFromCA(&goproxy.GoproxyCa),
		}

		return MitmConnect, host
	}

	proxy.OnRequest().HandleConnect(AlwaysMitm)
	// NTLM Transport
	// tr := concord.Transport{
	// 	Proxy:               http.ProxyURL(proxyUrl),
	// 	ProxyAuthorizer:     &handshakers.NTLMProxyAuthorizer{},
	// 	TLSClientConfig:     tlsClientConfig,
	// 	TLSHandshakeTimeout: time.Second * 15,
	// }

	// Handle Requests
	proxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		// ctx.RoundTripper = goproxy.RoundTripperFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (resp *http.Response, err error) {
		// 	resp, err = tr.RoundTrip(req)
		// 	return
		// })

		return req, nil
	})

	srv := &http.Server{
		Addr:        proxyBind,
		Handler:     proxy,
		IdleTimeout: time.Second * 60,
	}

	log.Fatal(srv.ListenAndServe())

}

// Check if it is a WebSocketUpgrade
// func IsWebSocketUpgrade() goproxy.ReqConditionFunc {
// 	return func(req *http.Request, ctx *goproxy.ProxyCtx) bool {
// 		return websocket.IsWebSocketUpgrade(req)
// 	}
// }
