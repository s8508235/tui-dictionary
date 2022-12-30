package dictionary

import (
	"sync"

	"golang.org/x/sync/errgroup"
)

type MyPrefer struct {
	Dictionaries []Interface
}

func (m *MyPrefer) Search(word string) ([]string, error) {
	eg := new(errgroup.Group)
	var wg sync.WaitGroup
	resultChan := make(chan string)
	result := make([]string, 0, 5)
	for _, dictionary := range m.Dictionaries {
		dictionary := dictionary
		wg.Add(1)
		eg.Go(func() error {
			defer wg.Done()
			r, err := dictionary.Search(word)
			if err != nil && err != ErrorNoDef {
				return err
			}
			for _, res := range r {
				res = re.ReplaceAllString(res, " ")
				resultChan <- res
			}
			return nil
		})
	}
	eg.Go(func() error {
		wg.Wait()
		close(resultChan)
		return nil
	})
	eg.Go(func() error {
		for definitions := range resultChan {
			result = append(result, definitions)
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return result, err
	}
	if len(result) == 0 {
		return result, ErrorNoDef
	}
	return result, nil
}
