package entity

import "errors"

type DictionaryLanguage int

const (
	English DictionaryLanguage = iota
	Russian
)

var ErrUnknownLanguage = errors.New("unknown language")
