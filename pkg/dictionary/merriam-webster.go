package dictionary

import (
	"github.com/gocolly/colly/v2"
)

const (
	websterURL         = "https://www.merriam-webster.com/dictionary/%s"
	websterURLSelector = "div.sb span.dt span.dtText"
)

func websterSearch(results *[]string, counter *int) func(e *colly.HTMLElement) {
	return func(e *colly.HTMLElement) {
		result := e.DOM.ReplaceWithSelection(e.DOM.ChildrenFiltered("strong.mw_t_bc")).Text()
		if *counter < 3 {
			*results = append(*results, result)
		} else {
			return
		}
		*counter += 1
	}
}
