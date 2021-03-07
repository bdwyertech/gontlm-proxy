package ntlm_proxy

import (
	"context"
	"flag"
	"net"
	"net/http"
	"net/url"
	"os"
	// "regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/bdwyertech/proxyplease"
	"github.com/elazarl/goproxy"
	"github.com/kofalt/go-memoize"
	// "github.com/bhendo/concord"
	// "github.com/bhendo/concord/handshakers"
)

var ProxyBind string
var ProxyServer string
var ProxyVerbose bool

func init() {
	flag.StringVar(&ProxyBind, "bind", getEnv("GONTLM_BIND", "http://0.0.0.0:3128"), "IP & Port to bind to")
	flag.StringVar(&ProxyServer, "proxy", getEnv("GONTLM_PROXY", ""), "Forwarding proxy server")
	flag.BoolVar(&ProxyVerbose, "verbose", false, "Enable verbose logging")
}

var ProxyUser = os.Getenv("GONTLM_USER")
var ProxyPass = os.Getenv("GONTLM_PASS")
var ProxyDomain = os.Getenv("GONTLM_DOMAIN")
var ProxyOverrides *map[string]*url.URL

func Run() {
	proxy := goproxy.NewProxyHttpServer()
	//
	// Log Configuration
	//
	if _, verbose := os.LookupEnv("GONTLM_PROXY_VERBOSE"); log.IsLevelEnabled(log.DebugLevel) || ProxyVerbose || verbose {
		if !log.IsLevelEnabled(log.DebugLevel) {
			log.SetLevel(log.DebugLevel)
		}
		proxy.Verbose = true
	}
	// Override ProxyPlease Logger
	proxyplease.SetDebugf(func(section string, msgs ...interface{}) {
		log.Debugf("proxyplease."+section, msgs...)
	})

	if ProxyServer == "" {
		ProxyServer = getProxyServer()
	}

	bind, err := url.Parse(ProxyBind)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("Listening on: %s", bind.Host)

	var proxyUrl *url.URL
	if ProxyServer != "" {
		proxyUrl, err = url.Parse(ProxyServer)
		if err != nil {
			log.Fatal(err)
		}
		if isLocalHost(proxyUrl.Hostname()) {
			if bind.Port() == proxyUrl.Port() {
				log.WithFields(log.Fields{
					"Bind":  bind.Host,
					"Proxy": proxyUrl.Host,
				}).Fatal("Loop condition detected!")
			}
		}
		log.Infof("Forwarding Proxy is: %s", proxyUrl.Redacted())
	}

	// Memoize DialContexts for 30 minutes
	dialerCache := memoize.NewMemoizer(30*time.Minute, 30*time.Minute)

	proxyDialer := func(addr string, pxyUrl *url.URL) proxyplease.DialContext {
		cacheKey := addr
		if pxyUrl != nil && pxyUrl.Host != "" {
			cacheKey = strings.Split(pxyUrl.Host, ":")[0]
		}
		s, _, _ := dialerCache.Memoize(cacheKey, func() (pxyCtx interface{}, err error) {
			if ProxyOverrides != nil {
				if pxy, ok := (*ProxyOverrides)[addr]; ok {
					if pxy == nil {
						d := net.Dialer{}
						return d.DialContext, nil
					}
					pxyUrl = pxy
				}
			}

			pxyCtx = proxyplease.NewDialContext(proxyplease.Proxy{
				URL:       pxyUrl,
				Username:  ProxyUser,
				Password:  ProxyPass,
				Domain:    ProxyDomain,
				TargetURL: &url.URL{Host: addr},
			})
			return
		})
		return s.(proxyplease.DialContext)
	}

	//
	// Proxy DialContexts
	//
	proxy.Tr.Proxy = nil

	// HTTP
	proxy.Tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return proxyDialer(addr, proxyUrl)(ctx, network, addr)
	}

	// HTTPS
	proxy.ConnectDial = func(network, addr string) (net.Conn, error) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		return proxyDialer(addr, proxyUrl)(ctx, network, addr)
	}

	//
	// HTTP Handler
	//
	//	var HttpConnect goproxy.FuncHttpsHandler = func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
	//		HTTPConnect := &goproxy.ConnectAction{
	//			Action:    goproxy.ConnectAccept,
	//			TLSConfig: goproxy.TLSConfigFromCA(&goproxy.GoproxyCa),
	//		}
	//
	//		return HTTPConnect, host
	//	}
	//	proxy.OnRequest(goproxy.ReqHostMatches(regexp.MustCompile(".*:80$|.*:8080$"))).HandleConnect(HttpConnect)

	//
	// Connect Handler
	//
	var AlwaysMitm goproxy.FuncHttpsHandler = func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		// HTTPSConnect := &goproxy.ConnectAction{
		// 	// ConnectMitm enables SSL Interception, required for request filtering over HTTPS.
		// 	// Action:    goproxy.ConnectMitm,
		// 	// ConnectAccept preserves upstream SSL Certificates, etc. TCP tunneling basically.
		// 	Action:    goproxy.ConnectAccept,
		// 	TLSConfig: goproxy.TLSConfigFromCA(&goproxy.GoproxyCa),
		// }

		// return HTTPSConnect, host
		return goproxy.OkConnect, host
	}
	proxy.OnRequest().HandleConnect(AlwaysMitm)

	//
	// Request Handling
	//
	// MITM Action is required for HTTPS Requests (e.g. goproxy.ConnectMitm instead of goproxy.ConnectAccept)
	//
	// proxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	// 	log.Fatal(req.URL.String())
	// 	return req, nil
	// })

	srv := &http.Server{
		Addr:        bind.Host,
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
