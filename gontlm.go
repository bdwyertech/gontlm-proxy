package gontlm

import (
	"crypto/tls"
	"github.com/elazarl/goproxy"
	"golang.org/x/sys/windows/registry"
	"log"
	"net"
	"net/http"
	"regexp"
	"time"
)

func main() {
	// Pull Proxy from the Registry
	k, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.QUERY_VALUE)
	if err != nil {
		log.Fatal(err)
	}
	defer k.Close()

	// proxyServer := os.Args[1]
	proxyServer, _, err := k.GetStringValue("ProxyServer")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Forwarding Proxy is: %q\n", proxyServer)

	setGoProxyCA()
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true

	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	var AlwaysMitmAuth goproxy.FuncHttpsHandler = func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		ntlmDialContext := WrapDialContext(dialer.DialContext, proxyServer)

		proxy.Tr = &http.Transport{
			Proxy:       nil,
			Dial:        dialer.Dial,
			DialContext: ntlmDialContext,
			// TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		return goproxy.MitmConnect, host
	}

	// Handle HTTP authenticate responses
	//	proxy.OnResponse(HasNegotiateChallenge()).DoFunc(func(r *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	//		ctx.Logf("Received 407 and Proxy-Authenticate from server, proceeding to reply")
	//
	//		headerstr := getAuthorizationHeader(os.Args[1])
	//
	//		// Modify the original request, and rerun the request
	//		ctx.Req.Header["Proxy-Authorization"] = []string{headerstr}
	//		client := http.Client{
	//			Transport: proxy.Tr,
	//		}
	//
	//		newr, err := client.Do(ctx.Req)
	//
	//		if err != nil {
	//			ctx.Warnf("New request failed: %v", err)
	//		}
	//
	//		ctx.Logf("Got response, forwarding it back to client")
	//
	//		// Return the new response in place of the original
	//		return newr
	//	})
	//

	// Handle HTTPS Connect Requests
	proxy.OnRequest(goproxy.ReqHostMatches(regexp.MustCompile("^.*"))).HandleConnect(AlwaysMitmAuth)

	// Handle HTTP Connect Requests
	proxy.OnRequest(goproxy.Not(goproxy.ReqHostMatches(regexp.MustCompile("^.*:443$")))).
		DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			ntlmDialContext := WrapDialContext(dialer.DialContext, proxyServer)

			proxy.Tr = &http.Transport{
				Proxy:           nil,
				Dial:            dialer.Dial,
				DialContext:     ntlmDialContext,
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}

			return req, nil
		})

	log.Fatal(http.ListenAndServe(":53128", proxy))

}
