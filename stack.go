package main

type Stack struct {
	values   []interface{}
	capacity int
}

func (s *Stack) Push(value interface{}) {
	if len(s.values) >= s.capacity {
		s.values = s.values[:len(s.values)-1]
	}
	s.values = append([]interface{}{value}, s.values...)
}
