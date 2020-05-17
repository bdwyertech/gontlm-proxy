package cmd

import (
	"flag"
	"fmt"
	"runtime"
	"time"

	proxy "github.com/bdwyertech/gontlm-proxy/pkg"
)

var verFlag = flag.Bool("version", false, "Display version")

var GitCommit string
var ReleaseVer string
var ReleaseDate string

func showVersion() {
	if GitCommit == "" {
		GitCommit = "DEVELOPMENT"
	}
	if ReleaseVer == "" {
		ReleaseVer = "DEVELOPMENT"
	}
	if ReleaseDate == "" {
		ReleaseDate = time.Now().String()
	}
	fmt.Println("version:", ReleaseVer)
	fmt.Println("commit:", GitCommit)
	fmt.Println("date:", ReleaseDate)
	fmt.Println("runtime:", runtime.Version())
}

func Execute() {
	flag.Parse()
	if *verFlag {
		showVersion()
		return
	}

	if runtime.GOOS == "windows" {
		proxy.RunWindows()
	} else {
		proxy.Run()
	}
}
