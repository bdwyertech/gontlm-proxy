module github.com/bdwyertech/gontlm-proxy

go 1.14

replace github.com/elazarl/goproxy => github.com/bdwyertech/goproxy v0.0.0-20200419011457-7aaf118834c9

replace github.com/aus/proxyplease => github.com/bdwyertech/proxyplease v0.0.0-20200419011350-b922e3822bff

// replace github.com/aus/proxyplease => ../proxyplease

require (
	github.com/aus/proxyplease v0.0.0-20200207024103-49defe237a30
	github.com/elazarl/goproxy v0.0.0-20200310082302-296d8939dc5a
	github.com/kardianos/service v1.1.0
	github.com/mattn/go-colorable v0.1.7
	github.com/mattn/go-isatty v0.0.12
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.1 // indirect
	golang.org/x/sys v0.0.0-20200223170610-d5e6a3e2c0ae
)
