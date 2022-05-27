package main

import (
	"time"

	"github.com/aaaton/golem/v4"
	"github.com/aaaton/golem/v4/dicts/en"
	"github.com/briandowns/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/s8508235/tui-dictionary/model"
	"github.com/s8508235/tui-dictionary/pkg/dictionary"
	"github.com/s8508235/tui-dictionary/pkg/log"
	"github.com/sirupsen/logrus"
)

func initialModel(logger *logrus.Logger, lemmatizer *golem.Lemmatizer, dictionary dictionary.Interface) model.Dictionary {
	target := textinput.New()
	target.Placeholder = "target"
	target.Focus()
	searchWord := textinput.New()
	searchWord.Placeholder = "test"
	return model.Dictionary{
		Choices:    make([]string, 0),
		Selected:   make(map[int]struct{}),
		Logger:     logger,
		Lemmatizer: lemmatizer,
		Dictionary: dictionary,
		Target:     target,
		SearchWord: searchWord,
		Spinner:    spinner.New(spinner.CharSets[43], 100*time.Millisecond), // Build our new spinner
	}
}

func main() {
	logger := log.New()
	lemmatizer, err := golem.New(en.New())
	if err != nil {
		logger.Errorln("Fail to init lemmatizer:", err)
		return
	}
	dict, err := dictionary.NewMyPreferDictionary(logger)
	if err != nil {
		logger.Errorln("Fail to init dictionary:", err)
		return
	}

	p := tea.NewProgram(initialModel(logger, lemmatizer, dict))

	if err := p.Start(); err != nil {
		logger.Fatal(err)
	}
}
