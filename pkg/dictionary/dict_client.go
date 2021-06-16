package dictionary

import (
	"net/textproto"

	"github.com/s8508235/tui-dictionary/pkg/log"
	"golang.org/x/net/dict"
)

type DICTClient struct {
	Client           *dict.Client
	Logger           *log.Logger
	DictionaryPrefix string
}

func (d *DICTClient) Search(word string) ([]string, error) {
	result := make([]string, 0, 3)
	defs, err := d.Client.Define(d.DictionaryPrefix, word)
	if err != nil {
		textprotoError, valid := err.(*textproto.Error)
		if !valid || textprotoError.Code != 552 {
			d.Logger.Logrus.Error(err)
			return result, err
		}
		return result, ErrorNoDef
	}
	if len(defs) == 0 {
		return result, ErrorNoDef
	}
	for idx, def := range defs {
		if idx >= 3 {
			break
		}
		result = append(result, string(def.Text))
	}
	return result, nil
}
