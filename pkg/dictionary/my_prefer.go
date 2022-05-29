package dictionary

type MyPrefer struct {
	Dictionaries []Interface
}

func (m *MyPrefer) Search(word string) ([]string, error) {
	result := make([]string, 0, 5)
	for _, dictionary := range m.Dictionaries {
		// if len(result) >= 5 {
		// 	return result, nil
		// }
		r, err := dictionary.Search(word)
		if err != nil && err != ErrorNoDef {
			return result, err
		}
		for _, res := range r {
			res = re.ReplaceAllString(res, " ")
			result = append(result, res)
		}
	}

	if len(result) == 0 {
		return result, ErrorNoDef
	}
	return result, nil
}
