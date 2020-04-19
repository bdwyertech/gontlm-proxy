package ntlm_proxy

import (
	"crypto/tls"
	"crypto/x509"
	"os"
	"os/user"
	"path"

	log "github.com/sirupsen/logrus"
)

func SetupGoProxyCA() (goproxyCa tls.Certificate) {
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

	goproxyCa, err = tls.LoadX509KeyPair(cert, key)
	if err != nil {
		log.Fatal(err)
	}

	if goproxyCa.Leaf, err = x509.ParseCertificate(goproxyCa.Certificate[0]); err != nil {
		log.Fatal(err)
	}

	// Validate CA Certificate
	_, err = goproxyCa.Leaf.Verify(x509.VerifyOptions{})
	if err != nil {
		switch err.(type) {
		case x509.UnknownAuthorityError:
			log.Println("WARN: GoNTLM-Proxy certificate is not trusted... If you want to avoid validation errors, you can add the certificate to your systems trust store.")
			log.Printf("WARN: GoNTLM-Proxy CA Cert Location: %s", cert)
			log.Printf("WARN: %s", err)
		default:
			log.Fatal(err)
		}
	}

	return
}
