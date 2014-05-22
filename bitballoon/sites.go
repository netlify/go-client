package bitballoon

import(
  "os"
  "io"
  "fmt"
  "time"
  "path"
  "path/filepath"
  "errors"
  "bytes"
  "mime/multipart"
)

var (
  defaultTimeout time.Duration = 5 * 60 // seconds
)

type SitesService struct {
  client *Client
}

type Site struct {
  Id string `json:"id"`
  UserId string `json:"user_id"`

  Name string `json:"name"`
  CustomDomain string `json:"custom_domain"`
  Password string `json:"password"`
  NotificationEmail string `json:"notification_email"`

  State string `json:"state"`
  Premium bool `json:"premium"`
  Claimed bool `json:"claimed"`

  Url string `json:"url"`
  AdminUrl string `json:"admin_url"`
  DeployUrl string `json:"deploy_url"`
  ScreenshotUrl string `json:"screenshot_url"`

  CreatedAt Timestamp `json:"created_at"`
  UpdatedAt Timestamp `json:"updated_at"`

  Zip string
  Dir string
}

func (s *SitesService) Get(id string) (*Site, error) {
  site := new(Site)

  _, err := s.client.Request("GET", path.Join("/sites", id), nil, site)

  return site, err
}

func (s *SitesService) List() ([]Site, error) {
  sites := new([]Site)

  _, err := s.client.Request("GET", "/sites", nil, sites)

  return *sites, err
}

func (s *SitesService) Update(site *Site) (error) {

  if &site.Zip != nil {
    return s.depoyZip(site)
  } else {
    return s.deployDir(site)
  }

  return nil
}

func (s *SitesService) WaitForReady(site *Site, timeout time.Duration) error {
  if site.State == "current" {
    return nil
  }

  if timeout == 0 {
    timeout = defaultTimeout
  }

  timedOut := false
  time.AfterFunc(timeout * time.Second, func() {
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

      site, err := s.Get(site.Id)
      if site != nil {
        fmt.Println("Site state is now: ", site.State)
      }
      if err != nil || (site != nil && site.State == "current") {
        done <- err
        break
      }
    }
  }()

  err := <- done
  return err
}

func (s *SitesService) deployDir(site *Site) error {
  return nil
}

func (s *SitesService) deployZip(site *Site) error {
  zipPath, err := filepath.Abs(site.Zip)
  if err != nil {
    return err
  }

  body := &bytes.Buffer{}
  writer := multipart.NewWriter(body)

  fileWriter, err := writer.CreateFormFile("zip", filepath.Base(zipPath))
  fileReader, err := os.Open(zipPath)
  defer fileReader.Close()

  if err != nil {
    return err
  }
  io.Copy(fileWriter, fileReader)

  writer.WriteField("name", site.Name)
  writer.WriteField("custom_domain", site.CustomDomain)
  writer.WriteField("password", site.Password)
  writer.WriteField("notification_email", site.NotificationEmail)

  err = writer.Close()
  if err != nil {
    return err
  }

  contentType := "multipar/form-data; boundary=" + writer.Boundary()
  options := &RequestOptions{RawBody: body, Headers: &map[string]string{"Content-Type": contentType}}

  _, err = s.client.Request("PUT", path.Join("/sites", site.Id), options, nil)

  return err
}
