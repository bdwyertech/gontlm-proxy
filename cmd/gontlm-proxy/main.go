package main

import (
	"flag"
	"fmt"

	proxy "github.com/bdwyertech/gontlm-proxy/pkg/gontlm-proxy"
)

var verFlag = flag.Bool("version", false, "Display version")

var GitCommit string
var ReleaseVer string

func showVersion() {
	if GitCommit == "" {
		GitCommit = "DEVELOPMENT"
	}
	if ReleaseVer == "" {
		ReleaseVer = "DEVELOPMENT"
	}
	fmt.Println("version:", ReleaseVer)
	fmt.Println("commit:", GitCommit)
}

func main() {
	flag.Parse()
	if *verFlag {
		showVersion()
		return
	}
	proxy.Run()
}
