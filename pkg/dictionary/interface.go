package dictionary

type Dictionary interface {
	Search(word string) ([3]string, int)
}
