module github.com/bdwyertech/gontlm-proxy

go 1.25.0

replace github.com/elazarl/goproxy => github.com/bdwyertech/goproxy v0.0.0-20230608195657-05e9c5da7707

replace github.com/darren/gpac => github.com/bdwyertech/gpac v0.0.0-20260714135407-41f3d46300f7

// replace github.com/aus/proxyplease => ../proxyplease

// replace github.com/bdwyertech/proxyplease => ../proxyplease

// replace github.com/rapid7/go-get-proxied => ../go-get-proxied

require (
	github.com/bdwyertech/go-scutil v0.0.0-20230606025039-57a4d936729f
	github.com/bdwyertech/proxyplease v0.1.1-0.20251114174812-28ae43a69614
	github.com/darren/gpac v0.0.0-20210609082804-b56d6523a3af
	github.com/elazarl/goproxy v0.0.0-20221015165544-a0805db90819
	github.com/jellydator/ttlcache/v3 v3.4.1
	github.com/kardianos/service v1.2.4
	github.com/mattn/go-colorable v0.1.14
	github.com/mattn/go-isatty v0.0.20
	github.com/sirupsen/logrus v1.9.3
	golang.org/x/sync v0.22.0
	golang.org/x/sys v0.46.0
)

require (
	github.com/alexbrainman/sspi v0.0.0-20250919150558-7d374ff0d59e // indirect
	github.com/bdwyertech/go-get-proxied v0.0.0-20221029171534-ea033ac5f9fa // indirect
	github.com/dlclark/regexp2/v2 v2.5.0 // indirect
	github.com/dop251/goja v0.0.0-20260701091749-b07b74453ea9 // indirect
	github.com/go-sourcemap/sourcemap v2.1.4+incompatible // indirect
	github.com/google/pprof v0.0.0-20260709232956-b9395ee17fa0 // indirect
	github.com/launchdarkly/go-ntlmssp v1.0.2 // indirect
	golang.org/x/crypto v0.53.0 // indirect
	golang.org/x/net v0.56.0 // indirect
	golang.org/x/text v0.40.0 // indirect
	h12.io/socks v1.0.3 // indirect
)
