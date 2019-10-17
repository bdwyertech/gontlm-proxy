package main

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/elazarl/goproxy"
	"log"
	"os"
	"os/user"
	"path"
)

func setGoProxyCA() error {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	cert := getEnv("GONTLM_CA", path.Clean(path.Join(usr.HomeDir, ".gontlm-ca.pem")))
	key := cert[0:len(cert)-len(path.Ext(cert))] + ".key"
	if _, err := os.Stat(cert); os.IsNotExist(err) {
		log.Printf("GoNTLM-Proxy CA does not exist.. Creating: %s", cert)
		createCertificate(cert, key)
	}

	goproxyCa, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return err
	}

	if goproxyCa.Leaf, err = x509.ParseCertificate(goproxyCa.Certificate[0]); err != nil {
		return err
	}

	// Validate CA Certificate
	_, err = goproxyCa.Leaf.Verify(x509.VerifyOptions{})
	if err != nil {
		switch err.(type) {
		case x509.UnknownAuthorityError:
			log.Println("WARN: GoNTLM-Proxy certificate is not trusted... You should add it to your trusted CA store!")
			log.Printf("WARN: GoNTLM-Proxy CA Cert Location: %s", cert)
			log.Printf("WARN: %s", err)
		default:
			log.Fatal(err)
		}
	}

	goproxy.GoproxyCa = goproxyCa
	goproxy.OkConnect = &goproxy.ConnectAction{Action: goproxy.ConnectAccept, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.MitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectMitm, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.HTTPMitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectHTTPMitm, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.RejectConnect = &goproxy.ConnectAction{Action: goproxy.ConnectReject, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	return nil
}
