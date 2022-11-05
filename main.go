package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aaaton/golem/v4"
	"github.com/aaaton/golem/v4/dicts/en"
	"github.com/c-bata/go-prompt"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/erikgeiser/promptkit"
	"github.com/erikgeiser/promptkit/selection"
	"github.com/s8508235/tui-dictionary/model"
	"github.com/s8508235/tui-dictionary/pkg/dictionary"
	"github.com/s8508235/tui-dictionary/pkg/entity"
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
	dictionary dictionary.Interface, out io.Writer, lang entity.DictionaryLanguage, target string) model.Dictionary {

	searchWord := textinput.New()
	if lang == entity.Russian {
		searchWord.Placeholder = "о́пыт"
	} else {
		searchWord.Placeholder = "test"
	}
	searchWord.Focus()
	s := spinner.New()
	// https://github.com/briandowns/spinner
	s.Spinner = spinner.Spinner{
		Frames: []string{"[>>> >]", "[]>>>> []", "[] >>>> []", "[] >>>> []", "[] >>>> []", "[] >>>>[]", "[>> >>]"},
		FPS:    100 * time.Millisecond,
	}
	return model.Dictionary{
		Logger:     logger,
		Target:     target,
		Language:   lang,
		Choices:    make([]string, 0),
		Selected:   make(map[int]struct{}),
		Out:        out,
		Lemmatizer: lemmatizer,
		Dictionary: dictionary,
		SearchWord: searchWord,
		Spinner:    s,
	}
}

func main() {
	logger := log.New()
	// logger.SetLevel(logrus.DebugLevel)
	files, err := os.ReadDir("./")
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

	sp := selection.New("Choose a language:", []*selection.Choice{
		{
			Index:  0,
			String: "English to English",
			Value:  "en",
		},
		{
			Index:  1,
			String: "Russian to English",
			Value:  "ru",
		},
	})
	sp.Filter = nil
	var choice *selection.Choice

	if choice, err = sp.RunPrompt(); err != nil && err != promptkit.ErrAborted {
		logger.Error("Error: %v\n", err)
		os.Exit(1)
	} else if err == promptkit.ErrAborted {
		logger.Info("Exit without choosing the language")
		os.Exit(0)
	}
	var language entity.DictionaryLanguage
	lang, ok := choice.Value.(string)
	if !ok {
		logger.Error("Error: %v\n", err)
		os.Exit(1)
	}
	switch lang {
	case "en":
		language = entity.English
	case "ru":
		language = entity.Russian
	default:
		logger.Error(entity.ErrUnknownLanguage)
		os.Exit(1)
	}

	var dict dictionary.Interface
	switch language {
	case entity.English:
		dict, err = dictionary.NewMyPreferDictionary(logger)
		if err != nil {
			logger.Errorln("Fail to init dictionary:", err)
			return
		}
	case entity.Russian:
		dict, err = dictionary.NewMyPreferRUDictionary(logger)
		if err != nil {
			logger.Errorln("Fail to init dictionary:", err)
			return
		}
	default:
		logger.Error(entity.ErrUnknownLanguage)
		os.Exit(1)
	}
	lemmatizer, err := golem.New(en.New())
	if err != nil {
		logger.Errorln("Fail to init lemmatizer:", err)
		return
	}
	// enter target -> loop (enter word, select definition)
	target := prompt.Input(
		"Target: ",
		targetCompleter(fileNameList),
		prompt.OptionShowCompletionAtStart(),
		prompt.OptionCompletionOnDown(),
	)
	tools.Exit()
	shouldPadding := false
	var out io.Writer
	if target == "/dev/null" {
		out = io.Discard
	} else {
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
		outFile, err := os.OpenFile(filepath.Clean(target), os.O_CREATE|os.O_RDWR|os.O_APPEND|os.O_SYNC, 0600)
		if err != nil {
			logger.Errorln("Fail to open output file", err)
			return
		}
		defer func() {
			if err := outFile.Close(); err != nil {
				logger.Errorf("Error closing file: %s\n", err)
			}
		}()
		if shouldPadding {
			lastByte := make([]byte, 2)
			end, err := outFile.Seek(0, io.SeekEnd)
			if err != nil {
				logger.Error(err)
			}
			if end > 1 {
				n, err := outFile.ReadAt(lastByte, end-2)
				if n != 2 {
					fmt.Printf("\n\033[31mfile corrupt\033[0m")
					return
				} else if err != nil {
					logger.Error(err)
					return
				}
				if string(lastByte) != "\n\n" {
					if _, err := outFile.WriteString("\n\n"); err != nil {
						return
					}
				}
			}
		}
		out = outFile
	}
	p := tea.NewProgram(initialModel(logger, lemmatizer, dict, out, language, target), tea.WithAltScreen())
	// p := tea.NewProgram(initialModel(logger, lemmatizer, dict, out, language, target))

	if m, err := p.StartReturningModel(); err != nil {
		logger.Fatal(err)
	} else if m, ok := m.(model.Dictionary); ok {
		if m.GetError() != nil {
			logger.Error(m.GetError())
		} else {
			logger.Infoln("normally exit")
		}
	}
}
