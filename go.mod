module github.com/bdwyertech/gontlm-proxy

go 1.16

replace github.com/elazarl/goproxy => github.com/bdwyertech/goproxy v0.0.0-20200419011457-7aaf118834c9

replace github.com/darren/gpac => github.com/bdwyertech/gpac v0.0.0-20210412191422-5e2441af9ea4

// replace github.com/aus/proxyplease => ../proxyplease

// replace github.com/rapid7/go-get-proxied => ../go-get-proxied

require (
	github.com/ReneKroon/ttlcache/v2 v2.6.0
	github.com/bdwyertech/go-scutil v0.0.0-20210306002117-b25267f54e45
	github.com/bdwyertech/proxyplease v0.1.1-0.20210411181055-87b0b23e723b
	github.com/elazarl/goproxy v0.0.0-00010101000000-000000000000
	github.com/kardianos/service v1.2.0
	github.com/mattn/go-colorable v0.1.8
	github.com/mattn/go-isatty v0.0.13
	github.com/sirupsen/logrus v1.8.1
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20210403161142-5e06dd20ab57
)
