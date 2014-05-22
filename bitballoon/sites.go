package bitballoon

import(
  "os"
  "io"
  "io/ioutil"
  "fmt"
  "time"
  "path"
  "path/filepath"
  "strings"
  "errors"
  "bytes"
  "crypto/sha1"
  "encoding/hex"
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

type DeployInfo struct {
  Id string `json:"id"`
  DeployId string `json:"deploy_id"`
  Required []string `json:"required"`
}

type siteUpdate struct {
  Name string `json:"name"`
  CustomDomain string `json:"custom_domain"`
  Password string `json:"password"`
  NotificationEmail string `json:"notification_email"`
  Files *map[string]string `json:"files"`
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

  if site.Zip != "" {
    return s.deployZip(site)
  } else {
    return s.deployDir(site)
  }

  options := &RequestOptions{JsonBody: site.mutableParams()}

  _, err := s.client.Request("PUT", path.Join("/sites", site.Id), options, site)

  return err
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
  files := map[string]string{}

  err := filepath.Walk(site.Dir, func(path string, info os.FileInfo, err error) error {
    if info.IsDir() == false {
      rel, err := filepath.Rel(site.Dir, path)
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
    JsonBody: &siteUpdate{
      Name: site.Name,
      CustomDomain: site.CustomDomain,
      Password: site.Password,
      NotificationEmail: site.NotificationEmail,
      Files: &files,
    },
  }

  fmt.Println("Files", files)

  deployInfo := new(DeployInfo)
  _, err = s.client.Request("PUT", filepath.Join("/sites", site.Id), options, deployInfo)

  if err != nil {
    return err
  }

  lookup := map[string]bool{}

  for _, sha := range(deployInfo.Required) {
    lookup[sha] = true
  }

  for path, sha := range(files) {
    if lookup[sha] == true {
      file, _ := os.Open(filepath.Join(site.Dir, path))
      defer file.Close()

      options = &RequestOptions{
        RawBody: file,
        Headers: &map[string]string{"Content-Type": "application/octet-stream"},
      }
      fmt.Println("Uploading %s", path)
      _, err = s.client.Request("PUT", filepath.Join("/sites", site.Id, "files", path), options, nil)
      if err != nil {
        fmt.Println("Error", err)
        return err
      }
    }
  }


  return err
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

  for key, value := range *site.mutableParams() {
    writer.WriteField(key, value)
  }

  err = writer.Close()
  if err != nil {
    return err
  }

  contentType := "multipar/form-data; boundary=" + writer.Boundary()
  options := &RequestOptions{RawBody: body, Headers: &map[string]string{"Content-Type": contentType}}

  _, err = s.client.Request("PUT", path.Join("/sites", site.Id), options, nil)

  return err
}

func (site *Site) mutableParams() *map[string]string {
  return &map[string]string{
      "name": site.Name,
      "custom_domain": site.CustomDomain,
      "password": site.Password,
      "notification_email": site.NotificationEmail,
  }
}
