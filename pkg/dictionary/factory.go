package dictionary

import (
	"fmt"
	"net/url"
	"regexp"

	"github.com/gocolly/colly/v2"
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

func NewCollinsDictionary(logger *log.Logger) (Interface, error) {
	c := colly.NewCollector()
	// don't want to cache anything since it should be a light query
	if err := c.SetStorage(&emptyStorage{}); err != nil {
		return nil, err
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
			return fmt.Sprintf(urbanURL, re.ReplaceAllString(word, "%20"))
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
			return fmt.Sprintf(learnerURL, re.ReplaceAllString(word, "%20"))
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
			return fmt.Sprintf(websterURL, re.ReplaceAllString(word, "%20"))
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
	dictionaries := []Interface{webster, learner, collins, urban}
	return &MyPrefer{
		Dictionaries: dictionaries,
	}, nil
}
