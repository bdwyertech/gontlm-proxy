module github.com/bdwyertech/gontlm-proxy

go 1.16

replace github.com/elazarl/goproxy => github.com/bdwyertech/goproxy v0.0.0-20200419011457-7aaf118834c9

// replace github.com/aus/proxyplease => ../proxyplease

// replace github.com/rapid7/go-get-proxied => ../go-get-proxied

require (
	github.com/bdwyertech/proxyplease v0.1.1-0.20210305210130-a100ce82b827
	github.com/elazarl/goproxy v0.0.0-00010101000000-000000000000
	github.com/kardianos/service v1.2.0
	github.com/mattn/go-colorable v0.1.8
	github.com/mattn/go-isatty v0.0.12
	github.com/sirupsen/logrus v1.8.0
	golang.org/x/sys v0.0.0-20210304203840-7b4935edff86
)
