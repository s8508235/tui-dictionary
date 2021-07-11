package dictionary

import (
	"fmt"
	"regexp"

	"github.com/gocolly/colly/v2"
	"github.com/s8508235/tui-dictionary/pkg/log"
	"golang.org/x/net/dict"
)

var re = regexp.MustCompile(`(?s)[\s]+`)

func generalWebDictionarySearch(results *[]string, counter *int) func(e *colly.HTMLElement) {
	return func(e *colly.HTMLElement) {
		if *counter < 3 {
			*results = append(*results, e.Text)
		} else {
			return
		}
		*counter += 1
	}
}

func NewDICTClient(logger *log.Logger, network, addr, prefix string) (*DICTClient, error) {
	client, err := dict.Dial(network, addr)
	if err != nil {
		return nil, err
	}

	return &DICTClient{
		network:          network,
		addr:             addr,
		Client:           client,
		Logger:           logger,
		DictionaryPrefix: prefix,
	}, nil
}

func NewCollinsDictionary(logger *log.Logger) Interface {
	return &WebDictionaryCrawler{
		Crawler: colly.NewCollector(),
		Logger:  logger,
		SearchURL: func(word string) string {
			return fmt.Sprintf(collinsURL, re.ReplaceAllString(word, "-"))
		},
		Selector:   collinsSelector,
		SearchFunc: generalWebDictionarySearch,
	}
}

func NewUrbanDictionary(logger *log.Logger) Interface {
	return &WebDictionaryCrawler{
		Crawler: colly.NewCollector(),
		Logger:  logger,
		SearchURL: func(word string) string {
			return fmt.Sprintf(urbanURL, re.ReplaceAllString(word, "%20"))
		},
		Selector:   urbanSelector,
		SearchFunc: generalWebDictionarySearch,
	}
}

func NewLearnerDictionary(logger *log.Logger) Interface {
	return &WebDictionaryCrawler{
		Crawler: colly.NewCollector(),
		Logger:  logger,
		SearchURL: func(word string) string {
			return fmt.Sprintf(learnerURL, re.ReplaceAllString(word, "%20"))
		},
		Selector:   learnerSelector,
		SearchFunc: generalWebDictionarySearch,
	}
}

func NewMyPreferDictionary(logger *log.Logger) (*MyPrefer, error) {
	learner := NewLearnerDictionary(logger)
	collins := NewCollinsDictionary(logger)
	urban := NewUrbanDictionary(logger)
	dictionaries := []Interface{learner, collins, urban}
	return &MyPrefer{
		Logger:       logger,
		Dictionaries: dictionaries,
	}, nil
}
