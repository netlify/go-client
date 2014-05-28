package bitballoon

import(
  "io/ioutil"
  "os"
  "time"
  "path"
  "path/filepath"
  "errors"
  "strings"
  "crypto/sha1"
  "encoding/hex"
)

// Deploy represents a specific deploy of a site
type Deploy struct {
  Id     string `json:"id"`
  SiteId string `json:"site_id"`
  UserId string `json:"user_id"`

  State   string `json:"state"`

  // Shas of files that needs to be uploaded before the deploy is ready
  Required []string `json:"required"`

  DeployUrl     string `json:"deploy_url"`
  ScreenshotUrl string `json:"screenshot_url"`

  CreatedAt Timestamp `json:"created_at"`
  UpdatedAt Timestamp `json:"updated_at"`


  client *Client
}

type DeployService struct {
  site   *Site
  client *Client
}

type deployFiles struct {
  Files  *map[string]string `json:"files"`
}

func (s *DeployService) apiPath() string {
  if s.site != nil {
    return path.Join(s.site.apiPath(), "deploys")
  } else {
    return "/deploys"
  }
}

func (s *DeployService) Create(dirOrZip string) (*Deploy, *Response, error) {
  if s.site == nil {
    return nil, nil, errors.New("You can only create a new deploy for an existing site (site.Deploys.Create(dirOrZip)))")
  }

  if strings.HasSuffix(dirOrZip, ".zip") {
    return s.deployZip(dirOrZip)
  } else {
    return s.deployDir(dirOrZip)
  }
}

func (s *DeployService) List(options *ListOptions) ([]Deploy, *Response, error) {
  deploys := new([]Deploy)

  reqOptions := &RequestOptions{QueryParams: options.toQueryParamsMap()}

  resp, err := s.client.Request("GET", s.apiPath(), reqOptions, deploys)

  for _, deploy := range(*deploys) {
    deploy.client = s.client
  }

  return *deploys, resp, err
}

func (d *DeployService) Get(id string) (*Deploy, *Response, error) {
  deploy := &Deploy{Id: id, client: d.client}
  resp, err := deploy.refresh()

  return deploy, resp, err
}

func (deploy *Deploy) apiPath() string {
  return path.Join("/deploys", deploy.Id)
}

func (deploy *Deploy) refresh() (*Response, error) {
  if deploy.Id == "" {
    return nil, errors.New("Cannot fetch deploy without an ID")
  }
  return deploy.client.Request("GET", deploy.apiPath(), nil, deploy)
}

func (s *DeployService) deployDir(dir string) (*Deploy, *Response, error) {
  files := map[string]string{}

  err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
    if info.IsDir() == false {
      rel, err := filepath.Rel(dir, path)
      if err != nil {
        return err
      }

      if strings.HasPrefix(rel, ".") || strings.Contains(rel, "/.") {
        return nil
      }

      sha := sha1.New()
      data, err := ioutil.ReadFile(path)

      if err != nil {
        return err
      }

      sha.Write(data)

      files[rel] = hex.EncodeToString(sha.Sum(nil))
    }

    return nil
  })

  options := &RequestOptions{
    JsonBody: &deployFiles{
      Files:             &files,
    },
  }

  deploy := &Deploy{client: s.client}
  resp, err := s.client.Request("POST", s.apiPath(), options, deploy)

  if err != nil {
    return deploy, resp, err
  }

  lookup := map[string]bool{}

  for _, sha := range deploy.Required {
    lookup[sha] = true
  }

  for path, sha := range files {
    if lookup[sha] == true {
      file, _ := os.Open(filepath.Join(dir, path))
      defer file.Close()

      options = &RequestOptions{
        RawBody: file,
        Headers: &map[string]string{"Content-Type": "application/octet-stream"},
      }
      resp, err = s.client.Request("PUT", filepath.Join(deploy.apiPath(), "files", path), options, nil)
      if err != nil {
        return deploy, resp, err
      }
    }
  }

  return deploy, resp, err
}

func (s *DeployService) deployZip(zip string) (*Deploy, *Response, error) {
  zipPath, err := filepath.Abs(zip)
  if err != nil {
    return nil, nil, err
  }

  zipFile, err := os.Open(zipPath)
  defer zipFile.Close()

  if (err != nil) {
    return nil, nil, err
  }

  options := &RequestOptions{RawBody: zipFile, Headers: &map[string]string{"Content-Type": "application/zip"}}

  deploy := &Deploy{client: s.client}
  resp, err := s.client.Request("POST", s.apiPath(), options, deploy)

  return deploy, resp, err
}

func (deploy *Deploy) WaitForReady(timeout time.Duration) error {
  if deploy.State == "ready" {
    return nil
  }

  if timeout == 0 {
    timeout = defaultTimeout
  }

  timedOut := false
  time.AfterFunc(timeout*time.Second, func() {
    timedOut = true
  })

  done := make(chan error)

  go func() {
    for {
      time.Sleep(1 * time.Second)

      if timedOut {
        done <- errors.New("Timeout while waiting for processing")
        break
      }

      _, err := deploy.refresh()
      if err != nil || (deploy.State == "ready") {
        done <- err
        break
      }
    }
  }()

  err := <-done
  return err
}
