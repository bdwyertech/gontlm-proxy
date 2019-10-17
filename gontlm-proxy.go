package main

import (
	"github.com/bhendo/concord"
	"github.com/bhendo/concord/handshakers"
	"github.com/elazarl/goproxy"
	"log"
	"net/http"
	"net/url"
	"regexp"
)

func main() {
	proxyServer := getProxyServer()
	log.Printf("Forwarding Proxy is: %q\n", proxyServer)
	proxyUrl, err := url.Parse("http://" + proxyServer)
	if err != nil {
		log.Fatal(err)
	}
	setGoProxyCA()
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = false

	var AlwaysMitmAuth goproxy.FuncHttpsHandler = func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		return goproxy.MitmConnect, host
	}

	// Handle HTTPS Connect Requests
	proxy.OnRequest(goproxy.ReqHostMatches(regexp.MustCompile("^.*"))).HandleConnect(AlwaysMitmAuth)

	// Handle HTTP Connect Requests
	proxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		t := concord.Transport{
			Proxy:           http.ProxyURL(proxyUrl),
			ProxyAuthorizer: &handshakers.NTLMProxyAuthorizer{},
		}

		ctx.RoundTripper = goproxy.RoundTripperFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (resp *http.Response, err error) {
			resp, err = t.RoundTrip(req)
			return
		})

		return req, nil
	})

	log.Fatal(http.ListenAndServe(":53128", proxy))

}
