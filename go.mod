module github.com/bdwyertech/gontlm-proxy

go 1.16

replace github.com/elazarl/goproxy => github.com/bdwyertech/goproxy v0.0.0-20200419011457-7aaf118834c9

// replace github.com/aus/proxyplease => ../proxyplease

// replace github.com/rapid7/go-get-proxied => ../go-get-proxied

require (
	github.com/bdwyertech/proxyplease v0.1.1-0.20210306012352-4ea051c58b75
	github.com/elazarl/goproxy v0.0.0-00010101000000-000000000000
	github.com/kardianos/service v1.2.0
	github.com/kofalt/go-memoize v0.0.0-20200917044458-9b55a8d73e1c
	github.com/mattn/go-colorable v0.1.8
	github.com/mattn/go-isatty v0.0.12
	github.com/sirupsen/logrus v1.8.0
	golang.org/x/sys v0.0.0-20210305230114-8fe3ee5dd75b
)
