package tools

import (
	"errors"
	"regexp"
	"unicode"

	"github.com/s8508235/tui-dictionary/pkg/entity"
)

var (
	ErrWord     = errors.New("should be a word")
	ErrCyrillic = errors.New("should be a cyrillic")
)

// WordValidate validate input word before searching dictionary
func WordValidate(s string, lang entity.DictionaryLanguage) error {
	switch lang {
	case entity.English:
		match, err := regexp.MatchString(`(?s)^[a-zA-Z\s]+$`, s)
		if err != nil {
			return err
		}
		if match {
			return nil
		} else {
			return ErrWord
		}
	case entity.Russian:
		for _, c := range s {
			if !unicode.Is(unicode.Cyrillic, c) && int32(c) != 769 && !unicode.IsSpace(c) {
				return ErrCyrillic
			}
		}
		return nil
	default:
		return entity.ErrUnknownLanguage
	}
}
