package netlify

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestSitesService_List(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/api/v1/sites", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `[{"id":"first"},{"id":"second"}]`)
	})

	sites, _, err := client.Sites.List(&ListOptions{})
	if err != nil {
		t.Errorf("Sites.List returned an error: %v", err)
	}

	expected := []Site{{Id: "first"}, {Id: "second"}}
	if !reflect.DeepEqual(sites, expected) {
		t.Errorf("Expected Sites.List to return %v, returned %v", expected, sites)
	}
}

func TestSitesService_List_With_Pagination(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/api/v1/sites", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, map[string]string{"page": "2", "per_page": "10"})
		fmt.Fprint(w, `[{"id":"first"},{"id":"second"}]`)
	})

	sites, _, err := client.Sites.List(&ListOptions{Page: 2, PerPage: 10})
	if err != nil {
		t.Errorf("Sites.List returned an error: %v", err)
	}

	expected := []Site{{Id: "first"}, {Id: "second"}}
	if !reflect.DeepEqual(sites, expected) {
		t.Errorf("Expected Sites.List to return %v, returned %v", expected, sites)
	}
}

func TestSitesService_Get(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/api/v1/sites/my-site", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"id":"my-site"}`)
	})

	site, _, err := client.Sites.Get("my-site")
	if err != nil {
		t.Errorf("Sites.Get returned an error: %v", err)
	}

	if site.Id != "my-site" {
		t.Errorf("Expected Sites.Get to return my-site, returned %v", site.Id)
	}
}
