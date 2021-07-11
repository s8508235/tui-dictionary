package tools

import (
	"errors"
	"regexp"
)

var (
	ErrWord = errors.New("should be a word")
)

// WordValidate validate input word before searching dictionary
func WordValidate(s string) error {
	match, err := regexp.MatchString(`(?s)^[a-zA-Z\s]+$`, s)
	if err != nil {
		return err
	}
	if match {
		return nil
	} else {
		return ErrWord
	}
}
