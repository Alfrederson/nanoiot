package main

type Stack struct {
	values   []interface{}
	capacity int
}

func (s *Stack) Push(value interface{}) {
	if len(s.values) >= s.capacity {
		s.values = s.values[1:] // drop oldest
	}
	s.values = append(s.values, value) // add newest
}
