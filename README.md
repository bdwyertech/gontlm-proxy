# gontlm-proxy
:wrench:  NTLM Proxy Forwarder in Golang.

[![GoDoc](https://godoc.org/github.com/bdwyertech/gontlm-proxy?status.svg)](https://godoc.org/github.com/bdwyertech/gontlm-proxy)
[![Build Status](https://github.com/bdwyertech/gontlm-proxy/workflows/Go/badge.svg?branch=master)](https://github.com/bdwyertech/gontlm-proxy/actions?query=workflow%3AGo+branch%3Amaster)
[![Coverage Status](https://coveralls.io/repos/github/bdwyertech/gontlm-proxy/badge.svg?branch=master)](https://coveralls.io/github/bdwyertech/gontlm-proxy?branch=master)
[![Gitter](https://img.shields.io/badge/Gitter-bdwyertech%2Fgontlm--proxy-brightgreen.svg)][gitter]

[gitter]: https://gitter.im/bdwyertech/gontlm-proxy

## Overview
This project was inspired by CNTLM & PX.  Operating behind a corporate proxy can make using tooling difficult.  It can also force you into putting your credentials into ENV variables, definitely not good!  The goal here is to leverage the Windows SSPI subsystem to authenticate to your proxy automatically.

## Usage
When GoNTLM-Proxy first starts, it reads the configured proxy from the Windows Registry `SOFTWARE\Microsoft\Windows\CurrentVersion\Internet Settings`, or can be set via the `GONTLM_PROXY` environment variable.

By default, GoNTLM-Proxy listens locally on port 3128, however this can be set via the `GONTLM_BIND` environment variable.

## Available environment variables

| Variable | Default | Description |
| --- | --- | --- |
| GONTLM_PROXY | On Win: from registry.<br>On MacOS: from `scutil`.<br>On others: "" | The upstream proxy URL |
| GONTLM_BIND | "http://0.0.0.0:3128" | This defines on which IP and port the proxy will be listen |
| GONTLM_USER | "" | The Username which will be used for the upstream proxy for authentication |
| GONTLM_PASS | "" | The Password which will be used for the upstream proxy for authentication |
| GONTLM_DOMAIN | "" | The Domain which will be used for the upstream proxy for authentication |
| GONTLM_CA | `USERS_HOMEDIR`/.gontlm-ca.pem | The Certificate Authority which will be used for TLS communication |
| GONTLM_PROXY_VERBOSE | false | This set the loglevel for the logging library |
| GONTLM_PROXY_IDLE_TIMEOUT | unset | This set the [IdleTimeout](https://pkg.go.dev/net/http#Server) for the proxy. The format is documented in [ParseDuration](https://pkg.go.dev/time#ParseDuration) |

## Connection Pooling and Timeout Defaults

By default, gontlm-proxy is tuned for compatibility with aggressive enterprise proxies, such as Bluecoat (Symantec ProxySG), which are known to close idle connections quickly. These defaults help avoid connection reuse errors and connection resets in such environments.

**Default values:**
- `MaxIdleConns`: 50
- `MaxIdleConnsPerHost`: 10
- `IdleConnTimeout`: 10s

You can override these defaults using the following environment variables:
- `GONTLM_MAX_IDLE_CONNS`
- `GONTLM_MAX_IDLE_CONNS_PER_HOST`
- `GONTLM_IDLE_CONN_TIMEOUT` (e.g., `10s`, `30s`)

**Why these defaults?**
- Bluecoat/Symantec ProxySG proxies often close idle TCP connections after 10–15 seconds. See:
  - [Go issue: Bluecoat closes idle connections](https://github.com/golang/go/issues/16465)
  - [Stack Overflow: Go HTTP client and Bluecoat](https://stackoverflow.com/questions/35522732/golang-http-client-bluecoat-proxy-connection-reset)
- These settings are safe for most environments, but users with less aggressive proxies can increase these values for better performance.

## Background Task
Running this as a background task is likely preferred over running it as a service.  Unfortunately, Windows does not let you run services as users without specifying credentials unless you turn off some Security Policy and I do not recommend this.  The whole purpose of this project is to remove the need for hardcoded credentials after all.

Chances are, you want to use this with a CLI tool, so I have found it best to run this as a background job with PowerShell.  The beauty of this is that when you close your terminal, it also kills the process.

```powershell
function GoNTLM-Enable {
	Remove-Job -Name GoNTLM-Proxy -Force -ErrorAction SilentlyContinue
	Start-Job -Name GoNTLM-Proxy -ScriptBlock { C:\Path\to\gontlm-proxy.exe }
	$env:http_proxy='http://127.0.0.1:3128'
}
```

## Service
If you run this as a service, it will run as NT AUTHORITY/SYSTEM.  If you wish to run it as another user, you can edit the service after installation.

## Install
Release binaries are available under the GitHub Releases page.  Alternatively, you can do this the Go way.
```console
$ go get github.com/bdwyertech/gontlm-proxy
```

## Development
```console
$ go run .\cmd\gontlm-proxy\
```

## License

MIT
