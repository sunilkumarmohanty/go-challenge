package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNumberHandler(t *testing.T) {

	type upstreamServerFields struct {
		result     *result
		statusCode int
		timeout    time.Duration
	}

	tests := []struct {
		name           string
		method         string
		wantStatusCode int
		wantResult     string
		invalidURL     string
		usfs           []upstreamServerFields
	}{
		{
			name: "Valid Single URL",
			usfs: []upstreamServerFields{
				upstreamServerFields{
					result: &result{
						Numbers: []int{1, 2, 3, 4},
					},
					statusCode: http.StatusOK,
				},
			},
			wantStatusCode: http.StatusOK,
			wantResult:     `{"numbers":[1,2,3,4]}`,
		},
		{
			name: "Valid 2 URLs",
			usfs: []upstreamServerFields{
				upstreamServerFields{
					result: &result{
						Numbers: []int{1, 2},
					},
					statusCode: http.StatusOK,
				},
				upstreamServerFields{
					result: &result{
						Numbers: []int{3, 4},
					},
					statusCode: http.StatusOK,
				},
			},
			wantStatusCode: http.StatusOK,
			wantResult:     `{"numbers":[1,2,3,4]}`,
		},
		{
			name: "Valid 2 URLs Duplicate Fields",
			usfs: []upstreamServerFields{
				upstreamServerFields{
					result: &result{
						Numbers: []int{1, 2},
					},
					statusCode: http.StatusOK,
				},
				upstreamServerFields{
					result: &result{
						Numbers: []int{3, 4, 5, 1, 2},
					},
					statusCode: http.StatusOK,
				},
			},
			wantStatusCode: http.StatusOK,
			wantResult:     `{"numbers":[1,2,3,4,5]}`,
		},
		{
			name: "Valid 2 URLsand 1 timeout",
			usfs: []upstreamServerFields{
				upstreamServerFields{
					result: &result{
						Numbers: []int{1, 2},
					},
					timeout:    490 * time.Millisecond,
					statusCode: http.StatusOK,
				},
				upstreamServerFields{
					result: &result{
						Numbers: []int{3, 4, 5, 1, 2},
					},
					timeout:    500 * time.Millisecond,
					statusCode: http.StatusOK,
				},
			},
			wantStatusCode: http.StatusOK,
			wantResult:     `{"numbers":[]}`,
		},
		{
			name: "Valid 2 URLs and 1 timeout",
			usfs: []upstreamServerFields{
				upstreamServerFields{
					result: &result{
						Numbers: []int{1, 2},
					},
					statusCode: http.StatusOK,
				},
				upstreamServerFields{
					result: &result{
						Numbers: []int{3, 4, 5, 1, 2},
					},
					timeout:    501 * time.Millisecond,
					statusCode: http.StatusOK,
				},
			},
			wantStatusCode: http.StatusOK,
			wantResult:     `{"numbers":[1,2]}`,
		},
		{
			name: "Valid 2 URLs Duplicate Fields and 1 invalid status code",
			usfs: []upstreamServerFields{
				upstreamServerFields{
					result: &result{
						Numbers: []int{1, 2},
					},
					statusCode: http.StatusOK,
				},
				upstreamServerFields{
					result: &result{
						Numbers: []int{3, 4, 5, 1, 2},
					},
					statusCode: http.StatusInternalServerError,
				},
			},
			wantStatusCode: http.StatusOK,
			wantResult:     `{"numbers":[1,2]}`,
		},
		{
			name:           "Valid 0 URLs",
			wantStatusCode: http.StatusOK,
			wantResult:     `{"numbers":[]}`,
		},
		{
			name:           "Invalid method",
			method:         http.MethodPost,
			wantStatusCode: http.StatusMethodNotAllowed,
			wantResult:     `405 Method Not Allowed`,
		},
		{
			name: "Invalid URL in query string",
			usfs: []upstreamServerFields{
				upstreamServerFields{
					result: &result{
						Numbers: []int{1, 2, 3, 4},
					},
					statusCode: http.StatusOK,
				},
			},
			invalidURL:     ":invalid.invalid",
			wantStatusCode: http.StatusOK,
			wantResult:     `{"numbers":[1,2,3,4]}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			servers := make([]*httptest.Server, 0)
			for _, field := range tt.usfs {

				makeHandler := func(field upstreamServerFields) http.HandlerFunc {
					return func(w http.ResponseWriter, r *http.Request) {
						if field.statusCode != http.StatusOK {
							w.WriteHeader(field.statusCode)
							return
						}
						time.Sleep(field.timeout)
						json.NewEncoder(w).Encode(field.result)
					}
				}(field)
				servers = append(servers, httptest.NewServer(makeHandler))

			}

			var query string
			for _, val := range servers {
				query = fmt.Sprintf("%s&u=%s", query, val.URL)
			}
			if len(tt.invalidURL) != 0 {
				query = fmt.Sprintf("%s&u=%s", query, tt.invalidURL)
			}
			req, err := http.NewRequest(tt.method, "/?"+query, nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(NumberHandler)
			start := time.Now().UnixNano()
			handler.ServeHTTP(rr, req)
			duration := time.Now().UnixNano() - start
			if duration/int64(time.Millisecond) >= 500 {
				t.Errorf("NumberHandler is taking too long. Got %v want less than equal to 500", duration/int64(time.Millisecond))
			}
			if tt.wantStatusCode != rr.Code {
				t.Errorf("NumberHandler returned wrong status code: got %v want %v",
					rr.Code, tt.wantStatusCode)
			}
			if tt.wantResult != strings.TrimSpace(rr.Body.String()) {
				t.Errorf("NumberHandler returned wrong body: got %v want %v",
					rr.Body.String(), string(tt.wantResult))
			}
		})
	}
}
