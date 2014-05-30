package bitballoon

import(
  "fmt"
  "reflect"
  "testing"
  "net/http"
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
    t.Errorf("Expected Deploys.List to return %v, returned %v",expected, deploys)
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
    t.Errorf("Expected Deploys.List to return %v, returned %v",expected, deploys)
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
