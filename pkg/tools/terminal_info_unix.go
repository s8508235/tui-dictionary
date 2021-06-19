// +build linux

package tools

import (
	"os/exec"
	"strconv"
	"strings"
)

func Lines() (int, error) {
	output, err := exec.Command("tput", "lines").Output()
	if err != nil {
		return -1, err
	}
	lines, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return -1, err
	}
	return lines, nil
}

func Cols() (int, error) {
	output, err := exec.Command("tput", "cols").Output()
	if err != nil {
		return -1, err
	}
	lines, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return -1, err
	}
	return lines, nil
}
