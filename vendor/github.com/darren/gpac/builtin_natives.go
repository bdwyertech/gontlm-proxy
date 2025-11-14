package gpac

import (
	"net"
	"time"

	"github.com/dop251/goja"
	"github.com/jellydator/ttlcache/v3"
)

var builtinNatives = map[string]func(*goja.Runtime) func(call goja.FunctionCall) goja.Value{
	"dnsResolve":  dnsResolve,
	"myIpAddress": myIPAddress,
}

// TTL Cache: Memoize DNS Lookups for 5 Minutes
var dnsCachier = ttlcache.New(
	ttlcache.WithTTL[string, string](5*time.Minute),
	ttlcache.WithDisableTouchOnHit[string, string](),
)

func dnsResolve(vm *goja.Runtime) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		arg := call.Argument(0)
		if arg == nil || arg.Equals(goja.Undefined()) {
			return goja.Null()
		}

		host := arg.String()

		if dctx := dnsCachier.Get(host); dctx != nil {
			return vm.ToValue(dctx.Value())
		}

		ips, err := net.LookupIP(host)
		if err != nil {
			return goja.Null()
		}
		ipAddr := ips[0].String()
		dnsCachier.Set(host, ipAddr, ttlcache.DefaultTTL)

		return vm.ToValue(ipAddr)
	}
}

func myIPAddress(vm *goja.Runtime) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		ifs, err := net.Interfaces()
		if err != nil {
			return goja.Null()
		}

		for _, ifn := range ifs {
			if ifn.Flags&net.FlagUp != net.FlagUp {
				continue
			}

			addrs, err := ifn.Addrs()
			if err != nil {
				continue
			}

			for _, addr := range addrs {
				ip, ok := addr.(*net.IPNet)
				if ok && ip.IP.IsGlobalUnicast() {
					ipstr := ip.IP.String()
					return vm.ToValue(ipstr)
				}
			}
		}
		return goja.Null()
	}
}
