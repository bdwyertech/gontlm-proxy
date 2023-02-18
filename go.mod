module github.com/bdwyertech/gontlm-proxy

go 1.17

replace github.com/elazarl/goproxy => github.com/bdwyertech/goproxy v0.0.0-20221023183101-a7527e535a15

replace github.com/darren/gpac => github.com/bdwyertech/gpac v0.0.0-20221123005236-07b6d0100900

// replace github.com/aus/proxyplease => ../proxyplease

// replace github.com/bdwyertech/proxyplease => ../proxyplease

// replace github.com/rapid7/go-get-proxied => ../go-get-proxied

require (
	github.com/bdwyertech/go-scutil v0.0.0-20210306002117-b25267f54e45
	github.com/bdwyertech/proxyplease v0.1.1-0.20221124181130-2c956e798f18
	github.com/elazarl/goproxy v0.0.0-20221015165544-a0805db90819
	github.com/jellydator/ttlcache/v2 v2.11.1
	github.com/kardianos/service v1.2.2
	github.com/mattn/go-colorable v0.1.13
	github.com/mattn/go-isatty v0.0.16
	github.com/sirupsen/logrus v1.9.0
	golang.org/x/sync v0.1.0
	golang.org/x/sys v0.5.0
)

require (
	github.com/alexbrainman/sspi v0.0.0-20210105120005-909beea2cc74 // indirect
	github.com/bdwyertech/go-get-proxied v0.0.0-20221029171534-ea033ac5f9fa // indirect
	github.com/darren/gpac v0.0.0-20210609082804-b56d6523a3af // indirect
	github.com/dlclark/regexp2 v1.7.0 // indirect
	github.com/dop251/goja v0.0.0-20221118162653-d4bf6fde1b86 // indirect
	github.com/go-sourcemap/sourcemap v2.1.3+incompatible // indirect
	github.com/launchdarkly/go-ntlmssp v1.0.1 // indirect
	golang.org/x/crypto v0.3.0 // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/text v0.7.0 // indirect
	h12.io/socks v1.0.3 // indirect
)
