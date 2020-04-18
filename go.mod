module github.com/bdwyertech/gontlm-proxy

go 1.14

replace github.com/elazarl/goproxy => github.com/bdwyertech/goproxy v0.0.0-20200418161122-ca28c8abeeda

replace github.com/aus/proxyplease => github.com/bdwyertech/proxyplease v0.0.0-20200418183505-056440848627

// replace github.com/aus/proxyplease => ../proxyplease

require (
	github.com/aus/proxyplease v0.0.0-20200207024103-49defe237a30
	github.com/elazarl/goproxy v0.0.0-20200310082302-296d8939dc5a
	github.com/kardianos/service v1.0.1-0.20191211031725-3c356ae54c8a
	github.com/mattn/go-colorable v0.1.6
	github.com/mattn/go-isatty v0.0.12
	github.com/sirupsen/logrus v1.5.0
	golang.org/x/sys v0.0.0-20200223170610-d5e6a3e2c0ae
)
