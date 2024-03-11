module github.com/bdwyertech/gontlm-proxy

go 1.17

replace github.com/elazarl/goproxy => github.com/bdwyertech/goproxy v0.0.0-20230608195657-05e9c5da7707

replace github.com/darren/gpac => github.com/bdwyertech/gpac v0.0.0-20231023175238-585710495f17

// replace github.com/aus/proxyplease => ../proxyplease

// replace github.com/bdwyertech/proxyplease => ../proxyplease

// replace github.com/rapid7/go-get-proxied => ../go-get-proxied

require (
	github.com/bdwyertech/go-scutil v0.0.0-20230606025039-57a4d936729f
	github.com/bdwyertech/proxyplease v0.1.1-0.20221126170535-8a386bcb7c7a
	github.com/elazarl/goproxy v0.0.0-20221015165544-a0805db90819
	github.com/jellydator/ttlcache/v2 v2.11.1
	github.com/kardianos/service v1.2.2
	github.com/mattn/go-colorable v0.1.13
	github.com/mattn/go-isatty v0.0.20
	github.com/sirupsen/logrus v1.9.3
	golang.org/x/sync v0.6.0
	golang.org/x/sys v0.18.0
)

require (
	github.com/alexbrainman/sspi v0.0.0-20210105120005-909beea2cc74 // indirect
	github.com/bdwyertech/go-get-proxied v0.0.0-20221029171534-ea033ac5f9fa // indirect
	github.com/darren/gpac v0.0.0-20210609082804-b56d6523a3af // indirect
	github.com/dlclark/regexp2 v1.10.0 // indirect
	github.com/dop251/goja v0.0.0-20230304130813-e2f543bf4b4c // indirect
	github.com/go-sourcemap/sourcemap v2.1.3+incompatible // indirect
	github.com/google/pprof v0.0.0-20230228050547-1710fef4ab10 // indirect
	github.com/launchdarkly/go-ntlmssp v1.0.1 // indirect
	golang.org/x/crypto v0.18.0 // indirect
	golang.org/x/net v0.20.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	h12.io/socks v1.0.3 // indirect
)
