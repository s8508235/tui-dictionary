//go:build linux || unix

package tools

import (
	"os"
	"os/exec"
)

// Exit gives back
func Exit() {
	rawModeOff := exec.Command("/bin/stty", "-raw", "echo")
	rawModeOff.Stdin = os.Stdin
	_ = rawModeOff.Run()
	//nolint
	rawModeOff.Wait() //#nosec G104
}
