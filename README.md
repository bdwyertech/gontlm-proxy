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
When GoNTLM-Proxy first starts, it will create a self-signed certificate, unique to your system.  It is created in your home folder at `~/.gontlm-ca.pem` and `~/.gontlm-ca.key` respectively.  If you want to avoid validation errors, you can add the certificate to your systems trust store.

It reads the configured proxy from the Windows Registry, or can be set via the `GONTLM_PROXY` environment variable.

By default, GoNTLM-Proxy listens locally on port 3128, however this can be set via the `GONTLM_BIND` environment variable.

## Background Task
Running this as a background task is likely preferred over running it as a service.  Unfortunately, Windows does not let you run services as users without specifying credentials unless you turn off some Security Policy and I do not recommend this.  The whole purpose of this project is to remove the need for hardcoded credentials after all.

Chances are, you want to use this with a CLI tool, so I have found it best to run this as a background job with PowerShell.  The beauty of this is that when you close your terminal, it also kills the process.

```powershell
function GoNTLM-Enable {
	Remove-Job -Name GoCNTLM-Proxy -Force -ErrorAction SilentlyContinue
	Start-Job -Name GoNTLM-Proxy -ScriptBlock { C:\Path\to\gontlm-proxy.exe }
	$env:http_proxy='http://127.0.0.1:3128'
}
```

## Service
If you run this as a service, it will run as NT AUTHORITY/SYSTEM.  If you wish to run it as another user, you can edit the service after installation.

Please note that when running the service as `SYSTEM`, dynamically-generated CA certificate and key will be located in the `SYSTEM` user's home folder at `C:\WINDOWS\system32\config\systemprofile\.gontlm-ca.pem`.  You can always replace these with your own if you already have cert/key.

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
