package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/s8508235/tui-dictionary/pkg/log"
	"github.com/sirupsen/logrus"

	"github.com/s8508235/tui-dictionary/pkg/dictionary"
)

func main() {
	st := time.Now()
	logger := log.New()
	dict, err := dictionary.NewCambridgeDictionary(logger)
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
		return
	} else if err != nil {
		logger.Errorln("Search error:", err)
		return
	}

	fmt.Printf("definition: %s cost %f seconds\n", results, time.Since(st).Seconds())
}
