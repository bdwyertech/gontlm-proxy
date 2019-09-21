package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"

	ntlm_sspi "github.com/alexbrainman/sspi/ntlm"
)

// DialContext is the DialContext function that should be wrapped with a
// NTLM Authentication.
//
// Example for DialContext:
//
// dialContext := (&net.Dialer{KeepAlive: 30*time.Second, Timeout: 30*time.Second}).DialContext
type DialContext func(ctx context.Context, network, addr string) (net.Conn, error)

// WrapDialContext wraps a DialContext with an NTLM Authentication to a proxy.
func WrapDialContext(dialContext DialContext, proxyAddress string) DialContext {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		conn, err := dialContext(ctx, network, proxyAddress)
		if err != nil {
			log.Printf("ntlm> Could not call dial context with proxy: %s", err)
			return conn, err
		}

		cred, err := ntlm_sspi.AcquireCurrentUserCredentials()
		if err != nil {
			log.Fatal(err)
		}
		defer cred.Release()

		secctx, negotiate, err := ntlm_sspi.NewClientContext(cred)
		if err != nil {
			log.Fatal(err)
		}
		defer secctx.Release()

		// NTLM Step 1: Send Negotiate Message
		log.Printf("ntlm> NTLM negotiate message: '%s'", base64.StdEncoding.EncodeToString(negotiate))
		header := make(http.Header)
		header.Set("Proxy-Authorization", fmt.Sprintf("NTLM %s", base64.StdEncoding.EncodeToString(negotiate)))
		header.Set("Proxy-Connection", "Keep-Alive")
		connect := &http.Request{
			Method: "CONNECT",
			URL:    &url.URL{Opaque: addr},
			Host:   addr,
			Header: header,
		}
		if err := connect.Write(conn); err != nil {
			log.Printf("ntlm> Could not write negotiate message to proxy: %s", err)
			return conn, err
		}
		log.Printf("ntlm> Successfully sent negotiate message to proxy")
		// NTLM Step 2: Receive Challenge Message
		br := bufio.NewReader(conn)
		resp, err := http.ReadResponse(br, connect)
		if err != nil {
			log.Printf("ntlm> Could not read response from proxy: %s", err)
			return conn, err
		}
		_, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("ntlm> Could not read response body from proxy: %s", err)
			return conn, err
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusProxyAuthRequired {
			log.Printf("ntlm> Expected %d as return status, got: %d", http.StatusProxyAuthRequired, resp.StatusCode)
			return conn, errors.New(http.StatusText(resp.StatusCode))
		}
		challenge := strings.Split(resp.Header.Get("Proxy-Authenticate"), " ")
		if len(challenge) < 2 {
			log.Printf("ntlm> The proxy did not return an NTLM challenge, got: '%s'", resp.Header.Get("Proxy-Authenticate"))
			return conn, errors.New("no NTLM challenge received")
		}
		log.Printf("ntlm> NTLM challenge: '%s'", challenge[1])
		challengeMessage, err := base64.StdEncoding.DecodeString(challenge[1])
		if err != nil {
			log.Printf("ntlm> Could not base64 decode the NTLM challenge: %s", err)
			return conn, err
		}
		// NTLM Step 3: Send Authorization Message
		log.Printf("ntlm> Processing NTLM challenge")
		authenticateMessage, err := secctx.Update(challengeMessage)
		if err != nil {
			log.Printf("ntlm> Could not process the NTLM challenge: %s", err)
			return conn, err
		}
		log.Printf("ntlm> NTLM authorization: '%s'", base64.StdEncoding.EncodeToString(authenticateMessage))
		header.Set("Proxy-Authorization", fmt.Sprintf("NTLM %s", base64.StdEncoding.EncodeToString(authenticateMessage)))
		connect = &http.Request{
			Method: "CONNECT",
			URL:    &url.URL{Opaque: addr},
			Host:   addr,
			Header: header,
		}
		if err := connect.Write(conn); err != nil {
			log.Printf("ntlm> Could not write authorization to proxy: %s", err)
			return conn, err
		}
		resp, err = http.ReadResponse(br, connect)
		if err != nil {
			log.Printf("ntlm> Could not read response from proxy: %s", err)
			return conn, err
		}
		if resp.StatusCode != http.StatusOK {
			log.Printf("ntlm> Expected %d as return status, got: %d", http.StatusOK, resp.StatusCode)
			return conn, errors.New(http.StatusText(resp.StatusCode))
		}
		// Succussfully authorized with NTLM
		log.Printf("ntlm> Successfully injected NTLM to connection")
		return conn, nil
	}
}
