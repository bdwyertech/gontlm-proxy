# gontlm
:wrench:  NTLM Proxy Forwarder in Golang.

[![GoDoc](https://godoc.org/github.com/bdwyertech/go-gontlm?status.svg)](https://godoc.org/github.com/bdwyertech/go-gontlm)
[![Build Status](https://travis-ci.org/bdwyertech/go-gontlm.svg?branch=master)](https://travis-ci.org/bdwyertech/go-gontlm)
[![Coverage Status](https://coveralls.io/repos/bdwyertech/go-gontlm/badge.svg?branch=master&service=github)](https://coveralls.io/github/bdwyertech/go-gontlm?branch=master)
[![](https://badge.imagelayers.io/bdwyertech/go-gontlm:latest.svg)](https://imagelayers.io/?images=bdwyertech/go-gontlm:latest)

## Overview
This project was inspired by CNTLM & PX.  Operating behind a corporate proxy can make using tooling difficult.  It can also force you into putting your credentials into ENV variables, definitely not good!  The goal here is to leverage the Windows SSPI subsystem to authenticate to your proxy automatically.

## Usage
TODO: Write some usage instructions

## Install

```console
$ go get github.com/bdwyertech/go-gontlm
```

## Development
```console
$ go run .
```

## License

MIT
