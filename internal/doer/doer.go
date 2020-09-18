package doer

import "net/http"

// Doer is a way of abstracting the HttpClient.Do method.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}
