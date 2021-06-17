package dictionary

import (
	"fmt"

	"github.com/s8508235/tui-dictionary/pkg/log"

	"github.com/gocolly/colly/v2"
)

const urbanURL = "https://www.urbandictionary.com/define.php?term=%s"

type Urban struct {
	Crawler *colly.Collector
	Logger  *log.Logger
}

func (c *Urban) Search(word string) ([]string, error) {
	crawler := c.Crawler.Clone()
	result := make([]string, 0, 3)
	count := 0
	crawler.OnHTML("div#content div.def-panel div.meaning", func(e *colly.HTMLElement) {
		if count < 3 {
			result = append(result, e.Text)
		} else {
			return
		}
		count += 1
	})

	crawler.OnError(func(_ *colly.Response, err error) {
		c.Logger.Logrus.Debugln("Something went wrong:", err)
	})

	crawler.OnRequest(func(r *colly.Request) {
		c.Logger.Logrus.Debugln("Visiting", r.URL.String())
	})

	crawler.OnScraped(func(r *colly.Response) {
		c.Logger.Logrus.Debugln("Finished", r.Request.URL.String())
	})

	searchURL := fmt.Sprintf(urbanURL, re.ReplaceAllString(word, "-"))
	err := crawler.Visit(searchURL)

	if err == nil && count == 0 {
		return result, ErrorNoDef
	}

	return result, nil
}
