// +build !windows

package ntlm_proxy

import (
	"log"
)

func RunWindows() {
	log.Fatal("RunWindows should never be called on a platform other than Windows!")
}
