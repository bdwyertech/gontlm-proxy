package ntlm_proxy

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	// "regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/bdwyertech/proxyplease"
	"github.com/elazarl/goproxy"
	"github.com/jellydator/ttlcache/v2"
	"golang.org/x/sync/singleflight"
	// "github.com/bhendo/concord"
	// "github.com/bhendo/concord/handshakers"
)

var ProxyBind string
var ProxyServer string
var ProxyVerbose bool
var ProxyContext context.Context

func init() {
	flag.StringVar(&ProxyBind, "bind", getEnv("GONTLM_BIND", "http://0.0.0.0:3128"), "IP & Port to bind to")
	flag.StringVar(&ProxyServer, "proxy", getEnv("GONTLM_PROXY", ""), "Forwarding proxy server")
	flag.BoolVar(&ProxyVerbose, "verbose", false, "Enable verbose logging")
}

var ProxyUser = os.Getenv("GONTLM_USER")
var ProxyPass = os.Getenv("GONTLM_PASS")
var ProxyDomain = os.Getenv("GONTLM_DOMAIN")
var ProxyOverrides map[string]*url.URL
var ProxyDialerCacheTimeout = 60 * time.Minute

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

	//
	// LRU Cache: Memoize DialContexts for 60 minutes
	//
	dialerCache := ttlcache.NewCache()
	dialerCache.SetTTL(ProxyDialerCacheTimeout)
	dialerCacheGroup := singleflight.Group{}

	directDialer := new(net.Dialer).DialContext

	proxyDialer := func(scheme, addr string, pxyUrl *url.URL) proxyplease.DialContext {
		cacheKey := addr
		if pxyUrl != nil && pxyUrl.Host != "" && ProxyOverrides == nil {
			cacheKey = pxyUrl.Host
		}

		if dctx, err := dialerCache.Get(cacheKey); err == nil {
			return dctx.(proxyplease.DialContext)
		}

		dctx, err, _ := dialerCacheGroup.Do(cacheKey, func() (pxyCtx interface{}, err error) {
			if ProxyOverrides != nil {
				var detected bool
				hosts := []string{addr, strings.Split(addr, ":")[0]}
				//
				// Exact Match
				//
				for _, host := range hosts {
					if pxy, ok := ProxyOverrides[strings.ToLower(host)]; ok {
						// If empty (nil) assume direct connection
						if pxy == nil {
							return directDialer, nil
						}

						detected = true
						pxyUrl = pxy
						break
					}
				}

				//
				// Suffix Match
				//
				if !detected {
					for _, host := range hosts {
						for dns, pxy := range ProxyOverrides {
							if strings.HasSuffix(strings.ToLower(host), dns) {
								// If empty (nil) assume direct connection
								if pxy == nil {
									return directDialer, nil
								}
								detected = true
								pxyUrl = pxy
								break
							}
						}
						if detected {
							break
						}
					}
				}

				//
				// Check if we need to tunnel
				//
				if detected {
					if tunnelPxy, ok := ProxyOverrides[pxyUrl.Host]; ok {
						var tunnelctx proxyplease.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
							conn, err := proxyplease.NewDialContext(proxyplease.Proxy{
								URL:      tunnelPxy,
								Username: ProxyUser,
								Password: ProxyPass,
								Domain:   ProxyDomain,
							})(ctx, network, pxyUrl.Host)

							if err != nil {
								return conn, err
							}

							req := &http.Request{
								Method: http.MethodConnect,
								URL:    &url.URL{Opaque: addr, Scheme: scheme},
								Host:   addr,
								Header: http.Header{
									"Proxy-Connection": []string{"Keep-Alive"},
								},
							}

							// req.Write(os.Stdout)
							br := bufio.NewReader(conn)
							if err = req.Write(conn); err != nil {
								return conn, err
							}

							resp, err := http.ReadResponse(br, req)
							resp.Body.Close()
							if err != nil {
								return conn, err
							}

							if resp.StatusCode == http.StatusOK {
								return conn, nil
							}
							return conn, errors.New(resp.Status)
						}
						dialerCache.Set(cacheKey, tunnelctx)
						return tunnelctx, nil
					}
				}
			}

			pxyCtx = proxyplease.NewDialContext(proxyplease.Proxy{
				URL:       pxyUrl,
				Username:  ProxyUser,
				Password:  ProxyPass,
				Domain:    ProxyDomain,
				TargetURL: &url.URL{Host: addr, Scheme: scheme},
			})

			err = dialerCache.Set(cacheKey, pxyCtx)

			return
		})
		if err != nil {
			log.Fatal(err)
		}

		return dctx.(proxyplease.DialContext)
	}

	//
	// Proxy DialContexts
	//
	proxy.Tr.Proxy = nil
	proxy.Tr.MaxIdleConnsPerHost = 10

	// HTTP
	proxy.Tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return proxyDialer("http", addr, proxyUrl)(ctx, network, addr)
	}

	// HTTPS
	proxy.ConnectDialWithReq = func(req *http.Request, network, addr string) (net.Conn, error) {
		return proxyDialer("https", addr, proxyUrl)(req.Context(), network, addr)
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
	var ConnectHandler goproxy.FuncHttpsHandler = func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
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
	proxy.OnRequest().HandleConnect(ConnectHandler)

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
		Handler: proxy,
		IdleTimeout: func() time.Duration {
			if timeout, err := time.ParseDuration(os.Getenv("GONTLM_PROXY_IDLE_TIMEOUT")); err == nil {
				return timeout
			} else {
				return 5 * time.Second
			}
		}(),
	}
	listener, err := net.Listen("tcp4", bind.Host)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.Serve(listener); err != nil {
			if ProxyContext == nil || !errors.Is(ProxyContext.Err(), context.Canceled) {
				log.Error(err)
			}
		}
	}()

	if ProxyContext != nil {
		<-ProxyContext.Done()
	} else {
		c := make(chan os.Signal, 1)
		// Accept graceful shutdowns when quit via SIGINT (Ctrl+C)
		// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
		signal.Notify(c, os.Interrupt)
		// Block until we receive our signal.
		<-c
	}

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
}

// Check if it is a WebSocketUpgrade
// func IsWebSocketUpgrade() goproxy.ReqConditionFunc {
// 	return func(req *http.Request, ctx *goproxy.ProxyCtx) bool {
// 		return websocket.IsWebSocketUpgrade(req)
// 	}
// }
