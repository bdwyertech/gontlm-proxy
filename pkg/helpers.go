package ntlm_proxy

import (
	"net"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func isLocalHost(host string) bool {
	if strings.ToLower(host) == "localhost" {
		return true
	}
	addr := net.ParseIP(host)
	if addr.IsLoopback() || addr.IsLinkLocalUnicast() {
		return true
	}
	ips := []net.IP{addr}
	if addr == nil {
		ipstrings, err := net.LookupHost(host)
		if err != nil {
			log.Fatal(err)
		}
		for _, ipString := range ipstrings {
			ips = append(ips, net.ParseIP(ipString))
		}
	}
	for _, ip := range ips {
		if ip.IsLoopback() || ip.IsLinkLocalUnicast() {
			return true
		}
		for _, localIP := range localIPs() {
			if localIP.Equal(ip) {
				return true
			}
		}
	}
	return false
}

func localIPs() (ips []net.IP) {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Fatal(err)
	}
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			log.Fatal(err)
		}
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				ips = append(ips, v.IP)
			case *net.IPAddr:
				ips = append(ips, v.IP)
			}
		}
	}
	return
}
