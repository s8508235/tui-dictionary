package main

import (
	"fmt"
	"os"
	"strings"

	log "github.com/s8508235/tui-dictionary/pkg/log"
	"github.com/sirupsen/logrus"

	"github.com/s8508235/tui-dictionary/pkg/dictionary"
)

func main() {

	logger := log.New()
	// dict, err := dictionary.NewMyPreferDictionary(logger, "tcp", "dict.dict.org:2628", "!")
	dict, err := dictionary.NewLearnerDictionary(logger)
	if err != nil {
		logger.Errorln("Fail to init dictionary:", err)
		return
	}
	logger.SetLevel(logrus.DebugLevel)

	argsWithoutProg := os.Args[1:]

	searchWord := strings.Join(argsWithoutProg, " ")
	results, err := dict.Search(searchWord)
	if err == dictionary.ErrorNoDef {
		fmt.Printf("no definition for: %s\n", searchWord)
	} else if err != nil {
		logger.Errorln("Search error:", err)
		return
	}

	fmt.Printf("definition: %s\n", results)
}
