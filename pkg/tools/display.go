package tools

import (
	"strconv"
	"strings"

	"github.com/s8508235/tui-dictionary/pkg/log"
)

// DisplayDefinition should fit definitions into a window
func DisplayDefinition(logger *log.Logger, lineLimit, colNum int, defs ...string) (string, error) {
	var buf strings.Builder
	var err error

	_, err = buf.WriteRune('\n')
	if err != nil {
		return "", err
	}
	// for definition of ... line and ending \n
	lineCount := 2
	for i, def := range defs {
		if len(def) > 0 {
			// +1 for def lines +1 for line number
			incr := strings.Count(def, "\n") + len(def)/colNum + 1 + 1
			lineCount += incr
			logger.Logrus.Debugln(i+1, lineCount, lineLimit, incr, def)
			if lineCount >= lineLimit && buf.Len() > 1 {
				lineCount -= incr
				continue
			}
			_, err = buf.WriteString(strconv.Itoa(i + 1))
			if err != nil {
				break
			}
			_, err = buf.WriteString(". ")
			if err != nil {
				break
			}
			_, err = buf.WriteString("\n\t")
			if err != nil {
				break
			}
			_, err = buf.WriteString(strings.ReplaceAll(def, "\n", "\n\t"))
			if err != nil {
				break
			}
			_, err = buf.WriteRune('\n')
			if err != nil {
				break
			}
		} else {
			break
		}
	}
	return buf.String(), err
}
