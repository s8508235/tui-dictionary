package entity

import "errors"

type DictionaryLanguage int
type DictionaryType int

const (
	English DictionaryLanguage = iota
	Russian
)

const (
	EnglishMyPrefer DictionaryType = iota
	RussianMyPrefer
	EnglishMyPreferWithUrban
)

var ErrUnknownLanguage = errors.New("unknown language")
