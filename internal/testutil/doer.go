package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
)

type FakeDoer struct {
	doer map[string]func(req *http.Request) (*http.Response, error)
	t    *testing.T
}

func (d *FakeDoer) Do(req *http.Request) (*http.Response, error) {
	key := req.Method + " " + req.URL.RequestURI()
	doer, ok := d.doer[key]
	if !ok {
		d.t.Errorf("No match for %s, not previously registered", key)
	}
	return doer(req)
}
func NewFakeDoer(t *testing.T) (d *FakeDoer) {
	return &FakeDoer{make(map[string]func(req *http.Request) (*http.Response, error)), t}
}

func (d *FakeDoer) HandleRequest(expectedMethod string, expectedURI string, statusCode int, response interface{}) {
	d.HandleBodyRequest(expectedMethod, expectedURI, statusCode, response, nil)
}

func (d *FakeDoer) HandleBodyRequest(expectedMethod string, expectedURI string, statusCode int, response interface{}, testBody func(io.ReadCloser) bool) {
	responseBytes, err := json.Marshal(response)
	if err != nil {
		panic("Error setting up test. Could not Marshal response.")
	}

	newFunc := func(req *http.Request) (*http.Response, error) {
		if testBody != nil && !testBody(req.Body) {
			d.t.Errorf("Body not as expected for %s %s. Body: `%v`", expectedMethod, expectedURI, req.Body)
		}
		resp := http.Response{
			StatusCode: statusCode,
			Body:       ioutil.NopCloser(bytes.NewReader(responseBytes)),
			Request:    req,
		}
		return &resp, nil
	}

	key := expectedMethod + " " + expectedURI
	currentFunc, ok := d.doer[key]
	if ok {
		// queue up another request:
		d.doer[key] = func(req *http.Request) (*http.Response, error) {
			// queue up the new function and execute the current function
			d.doer[key] = newFunc
			return currentFunc(req)
		}
	} else {
		d.doer[key] = newFunc
	}
}
