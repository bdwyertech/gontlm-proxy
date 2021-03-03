module github.com/bdwyertech/gontlm-proxy

go 1.16

replace github.com/elazarl/goproxy => github.com/bdwyertech/goproxy v0.0.0-20200419011457-7aaf118834c9

replace github.com/aus/proxyplease => github.com/bdwyertech/proxyplease v0.1.1-0.20210303211131-a5d95bc283d2

// replace github.com/aus/proxyplease => ../proxyplease

require (
	github.com/aus/proxyplease v0.0.0-00010101000000-000000000000
	github.com/elazarl/goproxy v0.0.0-00010101000000-000000000000
	github.com/kardianos/service v1.2.0
	github.com/mattn/go-colorable v0.1.8
	github.com/mattn/go-isatty v0.0.12
	github.com/sirupsen/logrus v1.8.0
	golang.org/x/sys v0.0.0-20210303074136-134d130e1a04
)
