package dictionary

import (
	"github.com/s8508235/tui-dictionary/pkg/log"
)

type MyPrefer struct {
	Logger       *log.Logger
	Dictionaries []Interface
}

func (m *MyPrefer) Search(word string) ([]string, error) {
	result := make([]string, 0, 5)
	for _, dictionary := range m.Dictionaries {
		if len(result) >= 5 {
			return result, nil
		}
		r, err := dictionary.Search(word)
		if err != nil && err != ErrorNoDef {
			return result, err
		}
		result = append(result, r...)
	}
	return result, nil
}
