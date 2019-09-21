package gontlm

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

	cert := path.Clean(path.Join(usr.HomeDir, "proxy-ca.pem"))
	if _, err := os.Stat(cert); os.IsNotExist(err) {
		log.Printf("Proxy CA does not exist: %s", cert)
		return err
	}

	goproxyCa, err := tls.LoadX509KeyPair(cert, cert)
	if err != nil {
		return err
	}

	if goproxyCa.Leaf, err = x509.ParseCertificate(goproxyCa.Certificate[0]); err != nil {
		return err
	}

	goproxy.GoproxyCa = goproxyCa
	goproxy.OkConnect = &goproxy.ConnectAction{Action: goproxy.ConnectAccept, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.MitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectMitm, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.HTTPMitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectHTTPMitm, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.RejectConnect = &goproxy.ConnectAction{Action: goproxy.ConnectReject, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	return nil
}
