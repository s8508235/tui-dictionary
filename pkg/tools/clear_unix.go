// +build aix darwin dragonfly freebsd js,wasm linux nacl netbsd openbsd solaris

package tools

import (
	"os"
	"os/exec"
)

// https://www.digitalocean.com/community/tutorials/building-go-applications-for-different-operating-systems-and-architectures

func Clear() error {
	cmd := exec.Command("clear") //Linux example, its tested
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
