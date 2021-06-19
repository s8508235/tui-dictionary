// +build windows

package tools

import (
	"errors"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const cmd = `MODE CON /Status`

var (
	re            = regexp.MustCompile(`[\s]+Lines\:\s*(\d+)\s+Columns:\s+(\d+)`)
	errModeFormat = errors.New("wrong format while parsing MODE command of windows")
)

func Lines() (int, error) {
	output, err := exec.Command("cmd", "/c", cmd).Output()
	if err != nil {
		return -1, err
	}
	matches := re.FindStringSubmatch(string(output))
	if len(matches) != 3 {
		return -1, errModeFormat
	}
	lines, err := strconv.Atoi(matches[1])
	if err != nil {
		return -1, err
	}
	return lines, nil
}

func Cols() (int, error) {
	output, err := exec.Command("cmd", "/c", cmd).Output()
	if err != nil {
		return -1, err
	}
	matches := re.FindStringSubmatch(strings.TrimSpace(string(output)))
	if len(matches) != 3 {
		return -1, errModeFormat
	}
	cols, err := strconv.Atoi(matches[2])
	if err != nil {
		return -1, err
	}
	return cols, nil
}
