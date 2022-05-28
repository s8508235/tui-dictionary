package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aaaton/golem/v4"
	"github.com/aaaton/golem/v4/dicts/en"
	"github.com/briandowns/spinner"
	"github.com/c-bata/go-prompt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/s8508235/tui-dictionary/model"
	"github.com/s8508235/tui-dictionary/pkg/dictionary"
	"github.com/s8508235/tui-dictionary/pkg/log"
	"github.com/s8508235/tui-dictionary/pkg/tools"
	"github.com/sirupsen/logrus"
)

func targetCompleter(fileNameList []string) prompt.Completer {
	return func(d prompt.Document) []prompt.Suggest {
		s := make([]prompt.Suggest, 0, len(fileNameList))
		for _, fileName := range fileNameList {
			s = append(s, prompt.Suggest{Text: fileName})
		}
		return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
	}
}

func initialModel(logger *logrus.Logger, lemmatizer *golem.Lemmatizer,
	dictionary dictionary.Interface, out *os.File, target string) model.Dictionary {

	searchWord := textinput.New()
	searchWord.Placeholder = "test"
	searchWord.Focus()
	return model.Dictionary{
		Logger:     logger,
		Target:     target,
		Choices:    make([]string, 0),
		Selected:   make(map[int]struct{}),
		OutFile:    out,
		Lemmatizer: lemmatizer,
		Dictionary: dictionary,
		SearchWord: searchWord,
		Spinner:    spinner.New(spinner.CharSets[43], 100*time.Millisecond), // Build our new spinner
	}
}

func main() {
	logger := log.New()
	// logger.SetLevel(logrus.DebugLevel)
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
	files, err := ioutil.ReadDir("./")
	if err != nil {
		logger.Errorln("Fail to read current directory:", err)
		return
	}
	fileNameList := make([]string, 0)
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".txt" {
			fileNameList = append(fileNameList, file.Name())
		}
	}
	logger.Debug(strings.Join(fileNameList, ","))
	// enter target -> loop (enter word, select definition)
	target := prompt.Input(
		"Target: ",
		targetCompleter(fileNameList),
		prompt.OptionShowCompletionAtStart(),
		prompt.OptionCompletionOnDown(),
	)
	tools.Exit()
	shouldPadding := false
	if _, err := os.Stat(target); err == nil {
		logger.Debugln("target exist")
		shouldPadding = true
	} else if !os.IsNotExist(err) {
		logger.Error(err)
		return
	}
	if filepath.Ext(target) == "" {
		target += ".txt"
	} else if filepath.Ext(target) != ".txt" {
		fmt.Println("\n\033[31mplease enter .txt as file extension\033[0m")
		return
	}
	out, err := os.OpenFile(filepath.Clean(target), os.O_CREATE|os.O_RDWR|os.O_APPEND|os.O_SYNC, 0600)
	if err != nil {
		logger.Errorln("Fail to open output file", err)
		return
	}
	if shouldPadding {
		lastByte := make([]byte, 2)
		end, err := out.Seek(0, io.SeekEnd)
		if err != nil {
			logger.Error(err)
		}
		if end > 1 {
			n, err := out.ReadAt(lastByte, end-2)
			if n != 2 {
				fmt.Printf("\n\033[31mfile corrupt\033[0m")
				return
			} else if err != nil {
				logger.Error(err)
				return
			}
			if string(lastByte) != "\n\n" {
				if _, err := out.WriteString("\n\n"); err != nil {
					return
				}
			}
			logger.Info("last byte: <", lastByte, ">")
		}
	}
	p := tea.NewProgram(initialModel(logger, lemmatizer, dict, out, target), tea.WithAltScreen())
	// p := tea.NewProgram(initialModel(logger, lemmatizer, dict, out, target))

	if err := p.Start(); err != nil {
		logger.Fatal(err)
	}
}
