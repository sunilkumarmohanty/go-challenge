package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
)

type task struct {
	url    string
	sortCh chan *result
	ctx    context.Context
}

func newTask(ctx context.Context, u string, sortCh chan *result) *task {
	if _, err := url.Parse(u); err != nil {
		log.Printf("Invalid url %v", u)
		return nil
	}
	return &task{
		url:    u,
		sortCh: sortCh,
		ctx:    ctx,
	}
}

func (t *task) do() {

	req, err := http.NewRequest(http.MethodGet, t.url, nil)
	if err != nil {
		log.Printf("Error creating request %v. Error:%v", t.url, err)
		return
	}
	req = req.WithContext(t.ctx)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error getting response %v. Error: %v", t.url, err)
		return
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		log.Printf("%v returned invalid status code %v", t.url, res.StatusCode)
		return
	}
	var numbers result

	if err = json.NewDecoder(res.Body).Decode(&numbers); err != nil {
		log.Printf("%v returned invalid data. Error:%v", t.url, err)
		return
	}
	// Send the numbers to the sorter
	t.sortCh <- &numbers
}
