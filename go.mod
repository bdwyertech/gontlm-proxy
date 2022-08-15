module github.com/bdwyertech/gontlm-proxy

go 1.16

replace github.com/elazarl/goproxy => github.com/bdwyertech/goproxy v0.0.0-20220523170445-07bdbc2a32f7

replace github.com/darren/gpac => github.com/bdwyertech/gpac v0.0.0-20220523171425-bda1923965db

// replace github.com/aus/proxyplease => ../proxyplease

// replace github.com/bdwyertech/proxyplease => ../proxyplease

// replace github.com/rapid7/go-get-proxied => ../go-get-proxied

require (
	github.com/bdwyertech/go-scutil v0.0.0-20210306002117-b25267f54e45
	github.com/bdwyertech/proxyplease v0.1.1-0.20211019140244-55998f26eb51
	github.com/elazarl/goproxy v0.0.0-00010101000000-000000000000
	github.com/jellydator/ttlcache/v2 v2.11.1
	github.com/kardianos/service v1.2.1
	github.com/mattn/go-colorable v0.1.12
	github.com/mattn/go-isatty v0.0.16
	github.com/sirupsen/logrus v1.8.1
	golang.org/x/sync v0.0.0-20220513210516-0976fa681c29
	golang.org/x/sys v0.0.0-20220811171246-fbc7d0a398ab
)
