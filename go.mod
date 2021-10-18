module github.com/bdwyertech/gontlm-proxy

go 1.16

replace github.com/elazarl/goproxy => github.com/bdwyertech/goproxy v0.0.0-20211013015200-f4875abca87a

replace github.com/darren/gpac => github.com/bdwyertech/gpac v0.0.0-20210806174158-5a8186b8896d

// replace github.com/aus/proxyplease => ../proxyplease

// replace github.com/bdwyertech/proxyplease => ../proxyplease

// replace github.com/rapid7/go-get-proxied => ../go-get-proxied

require (
	github.com/ReneKroon/ttlcache/v2 v2.9.0
	github.com/bdwyertech/go-scutil v0.0.0-20210306002117-b25267f54e45
	github.com/bdwyertech/proxyplease v0.1.1-0.20210806173316-144d3de1dd59
	github.com/elazarl/goproxy v0.0.0-00010101000000-000000000000
	github.com/kardianos/service v1.2.0
	github.com/mattn/go-colorable v0.1.11
	github.com/mattn/go-isatty v0.0.14
	github.com/sirupsen/logrus v1.8.1
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20210927094055-39ccf1dd6fa6
)
