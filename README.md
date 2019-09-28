# gontlm-proxy
:wrench:  NTLM Proxy Forwarder in Golang.

[![GoDoc](https://godoc.org/github.com/bdwyertech/gontlm-proxy?status.svg)](https://godoc.org/github.com/bdwyertech/gontlm-proxy)
[![Build Status](https://travis-ci.org/bdwyertech/gontlm-proxy.svg?branch=master)](https://travis-ci.org/bdwyertech/gontlm-proxy)
[![Coverage Status](https://coveralls.io/repos/bdwyertech/gontlm-proxy/badge.svg?branch=master&service=github)](https://coveralls.io/github/bdwyertech/gontlm-proxy?branch=master)
[![](https://badge.imagelayers.io/bdwyertech/gontlm-proxy:latest.svg)](https://imagelayers.io/?images=bdwyertech/gontlm-proxy:latest)

## Overview
This project was inspired by CNTLM & PX.  Operating behind a corporate proxy can make using tooling difficult.  It can also force you into putting your credentials into ENV variables, definitely not good!  The goal here is to leverage the Windows SSPI subsystem to authenticate to your proxy automatically.

## Usage
TODO: Write some usage instructions

## Install

```console
$ go get github.com/bdwyertech/gontlm-proxy
```

## Development
```console
$ go run .
```

## License

MIT
