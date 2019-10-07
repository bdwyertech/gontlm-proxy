package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"log"
	"math/big"
	"net"
	"os"
	"os/user"
	"time"
)

func CertTemplate() (*x509.Certificate, error) {
	// Generate a Random Serial
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, errors.New("Failed to generate serial number: " + err.Error())
	}

	// Grab User Information
	usr, err := user.Current()
	if err != nil {
		return nil, errors.New("Failed to retrieve current user: " + err.Error())
	}

	tmpl := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   "gontlm-proxy-" + usr.Username,
			Organization: []string{"GoNTLM-Proxy Self-Signed CA"},
		},
		SignatureAlgorithm:    x509.SHA256WithRSA,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(4, 0, 0), // 4 Years
		BasicConstraintsValid: true,
	}
	return &tmpl, nil
}

func CreateCert(template, parent *x509.Certificate, pub interface{}, parentPriv interface{}) (
	cert *x509.Certificate, err error) {

	certDER, err := x509.CreateCertificate(rand.Reader, template, parent, pub, parentPriv)
	if err != nil {
		return
	}
	// Parse the Certificate
	cert, err = x509.ParseCertificate(certDER)
	if err != nil {
		return
	}
	return
}

func createCertificate(certFile string, keyFile string) (err error) {
	// Create a Private key
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("Error generating random private key: %v", err)
	}

	rootCertTmpl, err := CertTemplate()
	if err != nil {
		log.Fatalf("Error creating CertTemplate: %v", err)
	}
	// Custom-Tailor the Certificate Template
	rootCertTmpl.IsCA = true
	rootCertTmpl.KeyUsage = x509.KeyUsageCertSign | x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature
	rootCertTmpl.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}
	rootCertTmpl.IPAddresses = []net.IP{net.ParseIP("127.0.0.1")}

	// Create the Certificate
	caCert, err := CreateCert(rootCertTmpl, rootCertTmpl, &privKey.PublicKey, privKey)
	if err != nil {
		log.Fatalf("Error creating cert: %v", err)
	}

	keyOut, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Failed to open %s for writing:", keyFile, err)
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		log.Fatalf("Unable to marshal private key: %v", err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		log.Fatalf("Failed to write key data to %s: %s", keyFile, err)
	}
	certOut, err := os.OpenFile(certFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalf("Failed to open %s for writing:", certFile, err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: caCert.Raw}); err != nil {
		log.Fatalf("Failed to write certificate data to %s: %s", certFile, err)
	}
	if err := certOut.Close(); err != nil {
		log.Fatalf("Error closing %s: %s", certFile, err)
	}
	return
}
