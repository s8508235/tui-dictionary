package main

import (
	"fmt"

	"github.com/s8508235/tui-dictionary/pkg/dictionary"
	"github.com/sirupsen/logrus"
)

type integrationTest struct {
	logger             *logrus.Logger
	failedDictionaries map[string][]string
	dictionaries       []dictionary.Interface
}

func checkEmptyMap(m map[string][]string) bool {
	for _, v := range m {
		if len(v) > 0 {
			return false
		}
	}
	return true
}

func main() {
	englishWord := "divest"
	splitedEnglishWord := "tie up"
	logger := logrus.New()
	logger.Level = logrus.DebugLevel
	oxford, err := dictionary.NewOxfordLearnerDictionary(logger)
	if err != nil {
		logger.Error(err)
	}
	cambridge, err := dictionary.NewCambridgeDictionary(logger)
	if err != nil {
		logger.Error(err)
	}
	webster, err := dictionary.NewWebsterDictionary(logger)
	if err != nil {
		logger.Error(err)
	}
	learner, err := dictionary.NewLearnerDictionary(logger)
	if err != nil {
		logger.Error(err)
	}

	intgTest := &integrationTest{
		logger:             logger,
		dictionaries:       []dictionary.Interface{oxford, cambridge, webster, learner},
		failedDictionaries: make(map[string][]string),
	}
	intgTest.testDictionary("normal", englishWord)
	intgTest.testDictionary("splitted", splitedEnglishWord)

	if checkEmptyMap(intgTest.failedDictionaries) {
		logger.Info("integration test success")
	} else {
		logger.Errorf("integration test failed with %s", intgTest.failedDictionaries)
	}
}

func (i *integrationTest) testDictionary(testCategory, word string) {
	for _, dict := range i.dictionaries {
		name := dict.GetName()
		if _, exist := i.failedDictionaries[name]; !exist {
			i.failedDictionaries[name] = make([]string, 0)
		}
		_, err := dict.Search(word)
		if err == dictionary.ErrorNoDef {
			i.failedDictionaries[name] = append(i.failedDictionaries[name], fmt.Sprintf("[%s]: no definitions", testCategory))
		} else if err != nil {
			i.failedDictionaries[name] = append(i.failedDictionaries[name], fmt.Sprintf("[%s]: %s", testCategory, err))
		}
	}
}
