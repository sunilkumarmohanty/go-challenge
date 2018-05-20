package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type result struct {
	Numbers []int `json:"numbers"`
}

const (
	timeout = 490
)

//NumberHandler handles the /numbers endpoint
func NumberHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Header().Add("Allow", http.MethodGet)
		w.Write([]byte("405 Method Not Allowed"))
		return
	}

	ctx := r.Context()
	// Timeout set in context with a buffer of 10 ms to return the result
	ctx, cancel := context.WithTimeout(ctx, timeout*time.Millisecond)
	defer cancel()
	//result channel is used to get the result after each successful retrieval f records
	resultCh := make(chan *result)
	//a message in done channel marks the end of all retrieval task
	done := make(chan bool)
	//res contains the recent result
	res := &result{Numbers: []int{}}
	urls := r.URL.Query()["u"]
	w.Header().Set("Content-Type", "application/json")
	if len(urls) == 0 {
		json.NewEncoder(w).Encode(result{
			Numbers: []int{},
		})
	} else {
		//retrieval of all records is done in a separate go subroutine
		go retrieve(ctx, urls, resultCh, done)
		for {
			select {
			case <-ctx.Done():
				//returns whatever the result has been computed when the ctx timeout occurs
				json.NewEncoder(w).Encode(res)
				return
			case <-done:
				json.NewEncoder(w).Encode(res)
				return
			case res = <-resultCh:
				// this just stores the latest result in res
				continue
			}
		}
	}
}

//retrieve is responsible for retrieving and sorting all the numbers received from the urls
func retrieve(ctx context.Context, urls []string, resultCh chan *result, done chan bool) {
	var wg sync.WaitGroup
	//sortCh is listened to by the sorter sub routing
	sortCh := make(chan *result)
	sorter := sorter{
		receiver: sortCh,
		result:   resultCh,
		done:     make(chan bool),
	}
	// Start the sorter job in a new subroutine
	go sorter.do()
	for _, u := range urls {
		//newTask validates the url and returns an instance if the url is valid
		t := newTask(ctx, u, sortCh)
		if t == nil {
			continue
		}
		wg.Add(1)
		//each of the task is run in a separate subroutine
		go func() {
			t.do()
			wg.Done()
		}()
	}
	//waits till all the tasks are done
	wg.Wait()
	//closes the sorter channel
	close(sorter.receiver)
	<-sorter.done
	//notifies the main subprocess that all the work is done
	done <- true
}
