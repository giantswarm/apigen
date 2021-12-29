package apigen

type StringSet struct {
	data map[string]struct{}
}

func NewStringSet() *StringSet {
	s := StringSet{
		data: map[string]struct{}{},
	}
	return &s
}

func (s *StringSet) Add(str string) {
	s.data[str] = struct{}{}
}

func (s *StringSet) Contains(str string) bool {
	_, exists := s.data[str]
	return exists
}

func (s *StringSet) Remove(str string) {
	delete(s.data, str)
}

func (s *StringSet) TakeOne() (string, bool) {
	if s.Size() == 0 {
		return "", false
	}

	var taken string
	for str, _ := range s.data {
		taken = str
		break
	}

	s.Remove(taken)
	return taken, true
}

func (s *StringSet) Size() int {
	return len(s.data)
}
