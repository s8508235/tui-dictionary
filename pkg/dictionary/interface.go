package dictionary

import "errors"

var ErrorNoDef = errors.New("no definition found")

type Interface interface {
	Search(word string) ([]string, error)
}
