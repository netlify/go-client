package netlify

import (
	"bytes"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestDeploysService_List(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/api/v1/deploys", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `[{"id":"first"},{"id":"second"}]`)
	})

	deploys, _, err := client.Deploys.List(&ListOptions{})
	if err != nil {
		t.Errorf("Deploys.List returned an error: %v", err)
	}

	expected := []Deploy{{Id: "first"}, {Id: "second"}}
	if !reflect.DeepEqual(deploys, expected) {
		t.Errorf("Expected Deploys.List to return %v, returned %v", expected, deploys)
	}
}

func TestDeploysService_List_For_Site(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/api/v1/sites/first-site/deploys", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `[{"id":"first"},{"id":"second"}]`)
	})

	site := &Site{Id: "first-site", client: client}
	site.Deploys = &DeploysService{client: client, site: site}

	deploys, _, err := site.Deploys.List(&ListOptions{})
	if err != nil {
		t.Errorf("Deploys.List returned an error: %v", err)
	}

	expected := []Deploy{{Id: "first"}, {Id: "second"}}
	if !reflect.DeepEqual(deploys, expected) {
		t.Errorf("Expected Deploys.List to return %v, returned %v", expected, deploys)
	}
}

func TestDeploysService_Get(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/api/v1/deploys/my-deploy", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"id":"my-deploy"}`)
	})

	deploy, _, err := client.Deploys.Get("my-deploy")
	if err != nil {
		t.Errorf("Sites.Get returned an error: %v", err)
	}

	if deploy.Id != "my-deploy" {
		t.Errorf("Expected Sites.Get to return my-deploy, returned %v", deploy.Id)
	}
}

func TestDeploysService_Create(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/api/v1/sites/my-site/deploys", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")

		r.ParseForm()
		if _, ok := r.Form["draft"]; ok {
			t.Errorf("Draft should not be a query parameter for a normal deploy")
		}

		fmt.Fprint(w, `{"id":"my-deploy"})`)
	})

	mux.HandleFunc("/api/v1/deploys/my-deploy", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")

		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)

		expected := `{"files":{"index.html":"3c7d0500e11e9eb9954ad3d9c2a1bd8b0fa06d88","style.css":"7b797fc1c66448cd8685c5914a571763e8a213da"},"async":false}`
		if expected != strings.TrimSpace(buf.String()) {
			t.Errorf("Expected JSON: %v\nGot JSON: %v", expected, buf.String())
		}

		fmt.Fprint(w, `{"id":"my-deploy"}`)
	})

	site := &Site{Id: "my-site"}
	deploys := &DeploysService{client: client, site: site}
	deploy, _, err := deploys.Create("test-site/folder")

	if err != nil {
		t.Errorf("Deploys.Create returned and error: %v", err)
	}

	if deploy.Id != "my-deploy" {
		t.Errorf("Expected Deploys.Create to return my-deploy, returned %v", deploy.Id)
	}
}

func TestDeploysService_CreateDraft(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/api/v1/sites/my-site/deploys", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")

		r.ParseForm()

		value := r.Form["draft"]
		if len(value) == 0 {
			t.Errorf("No draft query parameter, should be specified")
			return
		}

		draft := value[0]
		if draft != "true" {
			t.Errorf("Draft should be true but was %v", r.Form["draft"])
			return
		}

		fmt.Fprint(w, `{"id":"my-deploy"})`)
	})

	mux.HandleFunc("/api/v1/deploys/my-deploy", func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)

		expected := `{"files":{"index.html":"3c7d0500e11e9eb9954ad3d9c2a1bd8b0fa06d88","style.css":"7b797fc1c66448cd8685c5914a571763e8a213da"},"async":false}`
		if expected != strings.TrimSpace(buf.String()) {
			t.Errorf("Expected JSON: %v\nGot JSON: %v", expected, buf.String())
		}

		fmt.Fprint(w, `{"id":"my-deploy"}`)
	})

	site := &Site{Id: "my-site"}
	deploys := &DeploysService{client: client, site: site}
	deploy, _, err := deploys.CreateDraft("test-site/folder")

	if err != nil {
		t.Errorf("Deploys.Create returned and error: %v", err)
	}

	if deploy.Id != "my-deploy" {
		t.Errorf("Expected Deploys.Create to return my-deploy, returned %v", deploy.Id)
	}
}

func TestDeploysService_Create_Zip(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/api/v1/sites/my-site/deploys", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		r.ParseForm()
		if _, ok := r.Form["draft"]; ok {
			t.Errorf("Draft should not be a query parameter for a normal deploy")
		}

		fmt.Fprint(w, `{"id":"my-deploy"})`)
	})

	mux.HandleFunc("/api/v1/deploys/my-deploy", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")

		if r.Header["Content-Type"][0] != "application/zip" {
			t.Errorf("Deploying a zip should set the content type to application/zip")
			return
		}

		fmt.Fprint(w, `{"id":"my-deploy"}`)
	})

	site := &Site{Id: "my-site"}
	deploys := &DeploysService{client: client, site: site}
	deploy, _, err := deploys.Create("test-site/archive.zip")

	if err != nil {
		t.Errorf("Deploys.Create returned and error: %v", err)
	}

	if deploy.Id != "my-deploy" {
		t.Errorf("Expected Deploys.Create to return my-deploy, returned %v", deploy.Id)
	}
}

func TestDeploysService_CreateDraft_Zip(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/api/v1/sites/my-site/deploys", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		r.ParseForm()
		if val, ok := r.Form["draft"]; ok == false || val[0] != "true" {
			t.Errorf("Draft should be a true parameter for a draft deploy")
		}
		fmt.Fprint(w, `{"id":"my-deploy"})`)
	})

	mux.HandleFunc("/api/v1/deploys/my-deploy", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")

		if r.Header["Content-Type"][0] != "application/zip" {
			t.Errorf("Deploying a zip should set the content type to application/zip")
			return
		}

		fmt.Fprint(w, `{"id":"my-deploy"}`)
	})

	site := &Site{Id: "my-site"}
	deploys := &DeploysService{client: client, site: site}
	deploy, _, err := deploys.CreateDraft("test-site/archive.zip")

	if err != nil {
		t.Errorf("Deploys.Create returned an error: %v", err)
	}

	if deploy.Id != "my-deploy" {
		t.Errorf("Expected Deploys.Create to return my-deploy, returned %v", deploy.Id)
	}
}
