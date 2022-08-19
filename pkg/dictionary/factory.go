package dictionary

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/s8508235/tui-dictionary/pkg/cloudflare"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/dict"
)

var re = regexp.MustCompile(`(?s)[\s]+`)

type emptyStorage struct{}

// Init initializes emptyStorage
func (s *emptyStorage) Init() error {
	return nil
}

// Visited implements Storage.Visited()
func (s *emptyStorage) Visited(requestID uint64) error {
	return nil
}

// IsVisited implements Storage.IsVisited()
func (s *emptyStorage) IsVisited(requestID uint64) (bool, error) {
	return false, nil
}

// Cookies implements Storage.Cookies()
func (s *emptyStorage) Cookies(u *url.URL) string {
	return ""
}

// SetCookies implements Storage.SetCookies()
func (s *emptyStorage) SetCookies(u *url.URL, cookies string) {}

// Close implements Storage.Close()
func (s *emptyStorage) Close() error {
	return nil
}

func replaceSpaceWithASCII(word string) string {
	return re.ReplaceAllString(word, "%20")
}

func generalWebDictionarySearch(results *[]string, counter *int) func(e *colly.HTMLElement) {
	return func(e *colly.HTMLElement) {
		// if *counter < 3 {
		*results = append(*results, e.Text)
		// } else {
		// 	return
		// }
		*counter += 1
	}
}

// removeRussianAccentMarks remove stress in Russian but this is not best strategy
// https://russianalphabet.online/stress-marks-in-russian/
func removeRussianAccentMarks(word string) string {
	return strings.Map(func(r rune) rune {
		if int32(r) == 769 {
			return -1
		}
		return r
	}, word)
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

func NewCollinsDictionary(logger *log.Logger) (Interface, error) {
	c := colly.NewCollector()
	// don't want to cache anything since it should be a light query
	if err := c.SetStorage(&emptyStorage{}); err != nil {
		return nil, err
	}
	if client, err := cloudflare.NewClient(logger); err != nil {
		return nil, err
	} else {
		c.SetClient(client)
	}
	return &WebDictionaryCrawler{
		Crawler: c,
		Logger:  logger,
		SearchURL: func(word string) string {
			return fmt.Sprintf(collinsURL, re.ReplaceAllString(word, "-"))
		},
		Selector:   collinsSelector,
		SearchFunc: generalWebDictionarySearch,
	}, nil
}

func NewUrbanDictionary(logger *log.Logger) (Interface, error) {
	c := colly.NewCollector()
	// don't want to cache anything since it should be a light query
	if err := c.SetStorage(&emptyStorage{}); err != nil {
		return nil, err
	}
	return &WebDictionaryCrawler{
		Crawler: c,
		Logger:  logger,
		SearchURL: func(word string) string {
			return fmt.Sprintf(urbanURL, replaceSpaceWithASCII(word))
		},
		Selector:   urbanSelector,
		SearchFunc: generalWebDictionarySearch,
	}, nil
}

func NewLearnerDictionary(logger *log.Logger) (Interface, error) {
	c := colly.NewCollector()
	// don't want to cache anything since it should be a light query
	if err := c.SetStorage(&emptyStorage{}); err != nil {
		return nil, err
	}
	return &WebDictionaryCrawler{
		Crawler: c,
		Logger:  logger,
		SearchURL: func(word string) string {
			return fmt.Sprintf(learnerURL, replaceSpaceWithASCII(word))
		},
		Selector:   learnerSelector,
		SearchFunc: generalWebDictionarySearch,
	}, nil
}

func NewWebsterDictionary(logger *log.Logger) (Interface, error) {
	c := colly.NewCollector()
	// don't want to cache anything since it should be a light query
	if err := c.SetStorage(&emptyStorage{}); err != nil {
		return nil, err
	}
	return &WebDictionaryCrawler{
		Crawler: c,
		Logger:  logger,
		SearchURL: func(word string) string {
			return fmt.Sprintf(websterURL, replaceSpaceWithASCII(word))
		},
		Selector:   websterURLSelector,
		SearchFunc: websterSearch,
	}, nil
}

func NewMyPreferDictionary(logger *log.Logger) (*MyPrefer, error) {
	webster, err := NewWebsterDictionary(logger)
	if err != nil {
		return nil, err
	}
	learner, err := NewLearnerDictionary(logger)
	if err != nil {
		return nil, err
	}
	collins, err := NewCollinsDictionary(logger)
	if err != nil {
		return nil, err
	}
	urban, err := NewUrbanDictionary(logger)
	if err != nil {
		return nil, err
	}
	dictionaries := []Interface{collins, webster, learner, urban}
	return &MyPrefer{
		Dictionaries: dictionaries,
	}, nil
}

// NewDictComRussianEnglishDictionary must use with removeRussianAccentMarks, and don't care about stress
func NewDictComRussianEnglishDictionary(logger *log.Logger) (Interface, error) {
	c := colly.NewCollector()
	// don't want to cache anything since it should be a light query
	if err := c.SetStorage(&emptyStorage{}); err != nil {
		return nil, err
	}
	return &WebDictionaryCrawler{
		Crawler: c,
		Logger:  logger,
		SearchURL: func(word string) string {
			// logger.Info([]rune("й"))
			// logger.Info([]rune("е́"))
			// logger.Info([]rune("о́"))
			// logger.Info([]rune("у́"))
			// logger.Info([]rune("я́"))
			s := fmt.Sprintf(dictComRussianEnglishURL, replaceSpaceWithASCII(removeRussianAccentMarks(word)))
			return s
		},
		Selector:   dictComRussianEnglishSelector,
		SearchFunc: generalWebDictionarySearch,
	}, nil
}

func NewRussianDictDictionary(logger *log.Logger) (Interface, error) {
	c := colly.NewCollector()
	// don't want to cache anything since it should be a light query
	if err := c.SetStorage(&emptyStorage{}); err != nil {
		return nil, err
	}
	return &WebDictionaryCrawler{
		Crawler: c,
		Logger:  logger,
		SearchURL: func(word string) string {
			s := fmt.Sprintf(russianDictURL, replaceSpaceWithASCII(removeRussianAccentMarks(word)))
			logger.Infoln(s)
			return s
		},
		Selector:   russianDictSelector,
		SearchFunc: generalWebDictionarySearch,
	}, nil
}

// NewOpenRussianDictionary must use with removeRussianAccentMarks, and care about stress
func NewOpenRussianDictionary(logger *log.Logger) (Interface, error) {
	c := colly.NewCollector()
	// don't want to cache anything since it should be a light query
	if err := c.SetStorage(&emptyStorage{}); err != nil {
		return nil, err
	}
	return &WebDictionaryCrawler{
		Crawler: c,
		Logger:  logger,
		SearchURL: func(word string) string {
			s := fmt.Sprintf(openRussianURL, replaceSpaceWithASCII(removeRussianAccentMarks(word)))
			logger.Infoln(s)
			return s
		},
		Selector:   openRussianSelector,
		SearchFunc: generalWebDictionarySearch,
	}, nil
}

func NewMyPreferRUDictionary(logger *log.Logger) (*MyPrefer, error) {
	dictComRE, err := NewDictComRussianEnglishDictionary(logger)
	if err != nil {
		return nil, err
	}
	RUDict, err := NewRussianDictDictionary(logger)
	if err != nil {
		return nil, err
	}
	openRU, err := NewOpenRussianDictionary(logger)
	if err != nil {
		return nil, err
	}
	dictionaries := []Interface{dictComRE, RUDict, openRU}
	return &MyPrefer{
		Dictionaries: dictionaries,
	}, nil
}
