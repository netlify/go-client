package bitballoon

import (
  "net/url"
	"net/http"
	"net/http/httptest"
	"testing"
  "reflect"
)

var (
  mux *http.ServeMux
  client *Client
  server *httptest.Server
)

func setup() {
  mux = http.NewServeMux()
  server = httptest.NewServer(mux)

  client = NewClient(&Config{HttpClient: http.DefaultClient, BaseUrl: server.URL})
}

func teardown() {
  server.Close()
}

func testMethod(t *testing.T, r *http.Request, expected string) {
	if expected != r.Method {
		t.Errorf("Request method = %v, expected %v", r.Method, expected)
	}
}

type values map[string]string

func testFormValues(t *testing.T, r *http.Request, values values) {
	want := url.Values{}
	for k, v := range values {
		want.Add(k, v)
	}

	r.ParseForm()
	if !reflect.DeepEqual(want, r.Form) {
		t.Errorf("Request parameters = %v, want %v", r.Form, want)
	}
}

func TestResponse_populatePageValues(t *testing.T) {
	r := http.Response{
		Header: http.Header{
			"Link": {`<https://www.bitballoon.com/sites?page=1>; rel="first",` +
				` <https://www.bitballoon.com/sites?page=2>; rel="prev",` +
				` <https://www.bitballoon.com/sites?page=4>; rel="next",` +
				` <https://www.bitballoon.com/sites?page=5>; rel="last"`,
			},
		},
	}

	response := newResponse(&r)
	if expected, got := 1, response.FirstPage; expected != got {
		t.Errorf("response.FirstPage: %v, expected %v", got, expected)
	}
	if expected, got := 2, response.PrevPage; expected != got {
		t.Errorf("response.PrevPage: %v, expected %v", got, expected)
	}
	if expected, got := 4, response.NextPage; expected != got {
		t.Errorf("response.NextPage: %v, expected %v", got, expected)
	}
	if expected, got := 5, response.LastPage; expected != got {
		t.Errorf("response.LastPage: %v, expected %v", got, expected)
	}
}
