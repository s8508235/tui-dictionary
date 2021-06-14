package main

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/aaaton/golem/v4"
	"github.com/aaaton/golem/v4/dicts/en"
	"github.com/s8508235/tui-dictionary/model"
	"github.com/s8508235/tui-dictionary/pkg/database"
	"github.com/s8508235/tui-dictionary/pkg/dictionary"
	"github.com/s8508235/tui-dictionary/pkg/log"
	"github.com/tcnksm/go-input"
	"github.com/vmihailenco/msgpack"
	"gopkg.in/ini.v1"
	"gorm.io/gorm"
)

var errWord = errors.New("should be a word")

func main() {
	logger := log.New()
	cfg, err := ini.Load("app.ini")
	if err != nil {
		logger.Logrus.Errorln("Fail to read file", err)
		os.Exit(1)
	}
	dictionaryType := cfg.Section("").Key("dictionary").String()
	target := strings.ToLower(strings.ReplaceAll(cfg.Section("").Key("target").String(), " ", "-"))

	logLevel := cfg.Section("").Key("level").String()
	logger.SetLogLevel(logLevel)

	if len(target) == 0 {
		target = "target"
	}
	fmt.Println("Use dictionary:", dictionaryType)
	fmt.Println("Target:", target)
	fmt.Println("===== Press Ctrl+C to exit =====")
	db, err := database.NewSqlLiteConnection(target, logger)
	if err != nil {
		logger.Logrus.Errorln("Fail to open db", err)
		os.Exit(1)
	}
	db.AutoMigrate(&model.Dictionary{})
	sqlDB, err := db.DB()
	if err != nil {
		logger.Logrus.Errorln("Fail to access sql/db", err)
		os.Exit(1)
	}
	if err := sqlDB.Ping(); err != nil {
		logger.Logrus.Errorln("Fail to ping db", err)
		os.Exit(1)
	}
	defer sqlDB.Close()

	ui := &input.UI{
		Writer: os.Stdout,
		Reader: os.Stdin,
	}

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		logger.Logrus.Errorln("Fail to init lemmatizer", err)
		os.Exit(1)
	}
	dict := dictionary.NewCollinsDictionary(logger)

	query := "word"

	for {
		inputWord, err := ui.Ask(query, &input.Options{
			Required:    true,
			Loop:        true,
			HideDefault: true,
			HideOrder:   true,
			ValidateFunc: func(s string) error {
				match, err := regexp.MatchString(`(?s)^[a-zA-Z\s]+$`, s)
				logger.Logrus.Debug(match, err)
				if err != nil {
					logger.Logrus.Errorln("Fail to match pattern", err)
					return err
				}
				if match {
					return nil
				} else {
					logger.Logrus.Errorln(errWord)
					return errWord
				}
			},
		})
		if err != nil && err != input.ErrInterrupted {
			logger.Logrus.Errorln("Fail to ask", err)
			os.Exit(1)
		} else if err == input.ErrInterrupted {
			fmt.Println()
			os.Exit(0)
		}
		searchWord := lemmatizer.Lemma(inputWord)
		var word model.Dictionary
		result := db.Where("word = ?", searchWord).First(&word)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			results, count := dict.Search(searchWord)
			if count == 0 {
				fmt.Printf("no definition for : %s\n", searchWord)
				continue
			}
			defs, err := displayDefinition(results)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Printf("definition: %s\n", defs)
			b, err := msgpack.Marshal(results)
			if err != nil {
				fmt.Println(err)
			}
			db.Create(&model.Dictionary{Word: searchWord, Definition: b})
		} else {

			var results [3]string
			msgpack.Unmarshal(word.Definition, &results)
			defs, err := displayDefinition(results)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Printf("definition: %s\n", defs)
		}
	}
	// r := mux.NewRouter()
	// r.HandleFunc("/search/{word:\\w+}", searchDictionary(dict, db, lemmatizer)).Methods("GET")
	// http.ListenAndServe(":8087", r)
}

func displayDefinition(defs [3]string) (string, error) {
	var buf strings.Builder
	var err error
	_, err = buf.WriteRune('\n')
	if err != nil {
		return buf.String(), err
	}
	for i := 0; i < 3; i++ {
		if len(defs[i]) > 0 {
			_, err = buf.WriteString(strconv.Itoa(i + 1))
			if err != nil {
				break
			}
			_, err = buf.WriteString(". ")
			if err != nil {
				break
			}
			_, err = buf.WriteString(defs[i])
			if err != nil {
				break
			}
		} else {
			break
		}
	}
	return buf.String(), err
}

// func searchDictionary(dict dictionary.Dictionary, db *gorm.DB, lemmatizer *golem.Lemmatizer) func(w http.ResponseWriter, req *http.Request) {
// 	return func(w http.ResponseWriter, req *http.Request) {
// 		vars := mux.Vars(req)
// 		w.WriteHeader(http.StatusOK)
// 		searchWord := lemmatizer.Lemma(vars["word"])
// 		// fmt.Printf("trigger %s\n", searchWord)
// 		// fmt.Fprintf(w, "word: %v\n", )

// 		var word model.Dictionary
// 		result := db.Where("word = ?", searchWord).First(&word)
// 		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
// 			results, count := dict.Search(searchWord)
// 			if count == 0 {
// 				fmt.Fprintf(w, "no definition for : %s\n", searchWord)
// 				return
// 			}
// 			fmt.Fprintf(w, "definition: %v\n", results)
// 			b, err := msgpack.Marshal(results)
// 			if err != nil {
// 				fmt.Println(err)
// 			}
// 			db.Create(&model.Dictionary{Word: searchWord, Definition: b})
// 		} else {

// 			var defs [3]string
// 			msgpack.Unmarshal(word.Definition, &defs)
// 			fmt.Fprintf(w, "definition: %v\n", defs)
// 		}
// 	}
// }
