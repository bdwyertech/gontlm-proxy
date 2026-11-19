module github.com/bdwyertech/gontlm-proxy

go 1.26.0

replace github.com/elazarl/goproxy => github.com/bdwyertech/goproxy v0.0.0-20261119205855-bc7468447e3f

replace github.com/darren/gpac => github.com/bdwyertech/gpac v0.0.0-20261119173313-9230e40288bb

// replace github.com/aus/proxyplease => ../proxyplease

// replace github.com/bdwyertech/proxyplease => ../proxyplease

// replace github.com/rapid7/go-get-proxied => ../go-get-proxied

require (
	github.com/alexbrainman/sspi v0.0.0-20250919150558-7d374ff0d59e
	github.com/bdwyertech/go-scutil v0.0.0-20230606025039-57a4d936729f
	github.com/bdwyertech/proxyplease v0.1.1-0.20260714155200-12477b93b6d8
	github.com/darren/gpac v0.0.0-20210609082804-b56d6523a3af
	github.com/elazarl/goproxy v1.9.4
	github.com/jellydator/ttlcache/v3 v3.4.1
	github.com/kardianos/service v1.3.0
	github.com/launchdarkly/go-ntlmssp v1.0.5
	github.com/mattn/go-colorable v0.1.16
	github.com/mattn/go-isatty v0.0.24
	github.com/sirupsen/logrus v1.10.2
	golang.org/x/sync v0.25.0
	golang.org/x/sys v0.50.0
)

require (
	github.com/bdwyertech/go-get-proxied v0.0.0-20221029171534-ea033ac5f9fa // indirect
	github.com/dlclark/regexp2/v2 v2.13.1 // indirect
	github.com/dop251/goja v0.0.0-20261118204651-8dfece3d1e1d // indirect
	github.com/go-sourcemap/sourcemap v2.1.4+incompatible // indirect
	github.com/google/pprof v0.0.0-20261118010117-084b33a85f63 // indirect
	golang.org/x/net v0.63.0 // indirect
	golang.org/x/text v0.44.0 // indirect
	h12.io/socks v1.0.3 // indirect
)
