package dictionary

import (
	"regexp"

	"github.com/gocolly/colly/v2"
	"github.com/s8508235/tui-dictionary/pkg/log"
	"golang.org/x/net/dict"
)

var re = regexp.MustCompile(`(?s)[\s]+`)

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

func NewCollinsDictionary(logger *log.Logger) *Collins {
	return &Collins{
		Crawler: colly.NewCollector(),
		Logger:  logger,
	}
}

func NewUrbanDictionary(logger *log.Logger) *Urban {
	return &Urban{
		Crawler: colly.NewCollector(),
		Logger:  logger,
	}
}

func NewMyPreferDictionary(logger *log.Logger, network, addr, prefix string) (*MyPrefer, error) {
	dict, err := NewDICTClient(logger, network, addr, prefix)
	if err != nil {
		return nil, err
	}
	collins := NewCollinsDictionary(logger)
	urban := NewUrbanDictionary(logger)
	return &MyPrefer{
		Logger:  logger,
		Dict:    dict,
		Collins: collins,
		Urban:   urban,
	}, nil
}
