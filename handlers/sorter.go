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
	//uniqueNumber map helps us identify if a number is already in the sorted list
	uniqueNumbers := make(map[int]struct{})
	//the sorted number list
	numbers := make([]int, 0)
	//listen to the receiver channel until it is closed
	for res := range s.receiver {
		for _, val := range res.Numbers {
			if _, ok := uniqueNumbers[val]; !ok {
				uniqueNumbers[val] = struct{}{}
				numbers = append(numbers, val)
			}
		}
		//use of go sort package
		sort.Ints(numbers)
		s.result <- &result{numbers}
	}
	//notify completion
	s.done <- true
}
