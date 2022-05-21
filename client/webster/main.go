package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/s8508235/tui-dictionary/pkg/log"
	"github.com/s8508235/tui-dictionary/pkg/tools"
	"github.com/sirupsen/logrus"

	"github.com/s8508235/tui-dictionary/pkg/dictionary"
)

func main() {

	logger := log.New()

	dict, err := dictionary.NewWebsterDictionary(logger)
	if err != nil {
		logger.Logrus.Errorln("Fail to init dictionary:", err)
		return
	}
	logger.SetLogLevel(logrus.InfoLevel.String())

	argsWithoutProg := os.Args[1:]

	cols, err := tools.Cols()
	if err != nil {
		logger.Logrus.Errorln("Fail to get terminal info:", err)
		return
	}
	lines, err := tools.Lines()
	if err != nil {
		logger.Logrus.Errorln("Fail to get terminal info:", err)
		return
	}

	searchWord := strings.Join(argsWithoutProg, " ")
	results, err := dict.Search(searchWord)
	if err == dictionary.ErrorNoDef {
		fmt.Printf("no definition for: %s\n", searchWord)
	} else if err != nil {
		logger.Logrus.Errorln("Search error:", err)
		return
	}

	defs, err := tools.DisplayDefinition(logger, lines, cols, results...)
	if err != nil {
		logger.Logrus.Error(err)
		return
	}
	fmt.Printf("definition: %s\n", defs)
}
