//go:build !windows
// +build !windows

package ntlm_proxy

import (
	log "github.com/sirupsen/logrus"
)

func RunWindows() {
	log.Fatal("RunWindows should never be called on a platform other than Windows!")
}
