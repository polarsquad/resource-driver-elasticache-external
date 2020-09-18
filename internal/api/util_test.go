package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/matryer/is"
)

// NOTE: *_mock.go files are generated via the following commands:
// $ mockgen -source=main.go -destination=modeler_mock.go -package=main modeler
// $ mockgen -source=notify.go -destination=notifier_mock.go -package=main notifier

func ExecuteRequest(server Server, method, url string, body interface{}, t *testing.T) *httptest.ResponseRecorder {
	return ExecuteRequestHeader(server, method, url, body, nil, t)
}

func ExecuteRequestHeader(server Server, method, url string, body interface{}, header http.Header, t *testing.T) *httptest.ResponseRecorder {

	server.SetupRoutes()

	var req *http.Request
	var err error
	if body == nil {
		req, err = http.NewRequest(method, url, nil)
	} else {
		if b, ok := body.([]byte); ok {
			req, err = http.NewRequest(method, url, bytes.NewBuffer(b))
		} else if b, ok := body.(*bytes.Buffer); ok {
			req, err = http.NewRequest(method, url, b)
		} else {
			b, marshalErr := json.Marshal(body)
			if marshalErr != nil {
				t.Errorf("creating request: marshaling body to JSON: %v", marshalErr)
			}
			req, err = http.NewRequest(method, url, bytes.NewBuffer(b))
		}
	}
	if err != nil {
		t.Errorf("creating request: %v", err)
	}
	if header != nil {
		for item := range header {
			for _, value := range header.Values(item) {
				req.Header.Add(item, value)
			}
		}
	}
	w := httptest.NewRecorder()

	server.Router.ServeHTTP(w, req)

	return w
}

func TestValidIDs(t *testing.T) {
	is := is.New(t)

	is.True(isValidAsID("valid-id"))                   // "valid-id" is a valid id
	is.True(isValidAsID("01-valid-id-2"))              // "01-valid-id-2" is a valid id
	is.True(isValidAsID("jahgsdo87iq28ui3hdgkuyqxl3")) // "jahgsdo87iq28ui3hdgkuyqxl3" is a valid id

	is.True(!isValidAsID("-invalid-id")) // "-invalid-id" is not a valid id
	is.True(!isValidAsID("Invalid ID"))  // "Invalid ID" is not a not avalid id
	is.True(!isValidAsID(""))            // "" is a not a valid id
	is.True(!isValidAsID("a"))           // "a" is a not a valid id
}

func AreEqualExceptDates(a, b interface{}) bool {
	typeA := reflect.TypeOf(a)
	typeB := reflect.TypeOf(b)
	if typeA != typeB {
		return false
	}
	if typeA == reflect.TypeOf(time.Now()) {
		return true
	}
	if typeA.Kind() != reflect.Struct {
		return reflect.DeepEqual(a, b)
	}
	valueA := reflect.ValueOf(a)
	valueB := reflect.ValueOf(b)
	for i := 0; i < valueA.NumField(); i++ {
		if !AreEqualExceptDates(valueA.Field(i).Interface(), valueB.Field(i).Interface()) {
			return false
		}
	}
	return true
}
