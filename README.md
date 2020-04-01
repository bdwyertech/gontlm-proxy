# gontlm-proxy
:wrench:  NTLM Proxy Forwarder in Golang.

[![GoDoc](https://godoc.org/github.com/bdwyertech/gontlm-proxy?status.svg)](https://godoc.org/github.com/bdwyertech/gontlm-proxy)
[![Build Status](https://github.com/bdwyertech/gontlm-proxy/workflows/Go/badge.svg?branch=master)](https://github.com/bdwyertech/gontlm-proxy/actions?query=workflow%3AGo+branch%3Amaster)
[![Coverage Status](https://coveralls.io/repos/bdwyertech/gontlm-proxy/badge.svg?branch=dev&service=github)](https://coveralls.io/github/bdwyertech/gontlm-proxy?branch=dev)
[![](https://badge.imagelayers.io/bdwyertech/gontlm-proxy:latest.svg)](https://imagelayers.io/?images=bdwyertech/gontlm-proxy:latest)

## Overview
This project was inspired by CNTLM & PX.  Operating behind a corporate proxy can make using tooling difficult.  It can also force you into putting your credentials into ENV variables, definitely not good!  The goal here is to leverage the Windows SSPI subsystem to authenticate to your proxy automatically.

## Usage
When GoNTLM-Proxy first starts, it will create a self-signed certificate, unique to your system.  It is created in your home folder at `~/.gontlm-ca.pem` and `~/.gontlm-ca.key` respectively.  If you want to avoid validation errors, you can add the certificate to your systems trust store.

It reads the configured proxy from the Windows Registry, or can be set via the `GONTLM_PROXY` environment variable.

By default, GoNTLM-Proxy listens locally on port 3128, however this can be set via the `GONTLM_BIND` environment variable.

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
