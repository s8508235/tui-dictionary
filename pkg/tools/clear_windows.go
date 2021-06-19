// +build windows

package tools

import (
	"os"
	"os/exec"
)

func Clear() error {
	cmd := exec.Command("cmd", "/c", "cls") //Windows example, its tested
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
