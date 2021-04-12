package gpac

import (
	"net"
	"time"

	"github.com/ReneKroon/ttlcache/v2"
	"github.com/dop251/goja"
)

var builtinNatives = map[string]func(*goja.Runtime) func(call goja.FunctionCall) goja.Value{
	"dnsResolve":  dnsResolve,
	"myIpAddress": myIPAddress,
}

//
// TTL Cache: Memoize DNS Lookups for 5 Minutes
//
var dnsCachier *ttlcache.Cache

func dnsCache() *ttlcache.Cache {
	if dnsCachier != nil {
		return dnsCachier
	}
	dnsCachier = ttlcache.NewCache()
	dnsCachier.SetTTL(5 * time.Minute)
	dnsCachier.SkipTTLExtensionOnHit(true)

	return dnsCachier
}

func dnsResolve(vm *goja.Runtime) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		arg := call.Argument(0)
		if arg == nil || arg.Equals(goja.Undefined()) {
			return goja.Null()
		}

		host := arg.String()

		if dctx, err := dnsCache().Get(host); err == nil {
			return vm.ToValue(dctx.(string))
		}

		ips, err := net.LookupIP(host)
		if err != nil {
			return goja.Null()
		}
		ipAddr := ips[0].String()
		dnsCache().Set(host, ipAddr)

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
