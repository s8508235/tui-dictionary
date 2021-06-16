package dictionary

import (
	"github.com/s8508235/tui-dictionary/pkg/log"
)

type MyPrefer struct {
	Logger  *log.Logger
	Dict    *DICTClient
	Collins *Collins
	Urban   *Urban
}

func (m *MyPrefer) Search(word string) ([]string, error) {
	result := make([]string, 0, 5)
	collinsResult, err := m.Collins.Search(word)
	if err != nil && err != ErrorNoDef {
		return result, err
	}
	m.Logger.Logrus.Debugln("collins", len(collinsResult))
	result = append(result, collinsResult...)
	dictResult, err := m.Dict.Search(word)
	if err != nil && err != ErrorNoDef {
		return result, err
	}
	m.Logger.Logrus.Debugln("dict.org", len(dictResult))
	result = append(result, dictResult...)

	if len(result) >= 5 {
		return result, nil
	}
	m.Logger.Logrus.Info("=== Use unofficial dictionary ===")
	urbanResult, err := m.Urban.Search(word)
	if err != nil && err != ErrorNoDef {
		return result, err
	}
	result = append(result, urbanResult...)
	if len(result) == 0 {
		return result, ErrorNoDef
	}
	return result, nil
}
