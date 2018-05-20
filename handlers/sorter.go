package handlers

import (
	"sort"
)

type sorter struct {
	receiver chan *result
	result   chan *result
	done     chan bool
}

func (s *sorter) do() {
	uniqueNumbers := make(map[int]struct{})
	numbers := make([]int, 0)
	for res := range s.receiver {
		for _, val := range res.Numbers {
			if _, ok := uniqueNumbers[val]; !ok {
				uniqueNumbers[val] = struct{}{}
				numbers = append(numbers, val)
			}
		}
		sort.Ints(numbers)
		s.result <- &result{numbers}
	}
	s.done <- true
}
