package bitballoon

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io"
	"io/ioutil"
	"fmt"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var (
	defaultTimeout time.Duration = 5 * 60 // 5 minutes
)

type SitesService struct {
	client *Client
}

type Site struct {
	Id     string `json:"id"`
	UserId string `json:"user_id"`

	Name              string `json:"name"`
	CustomDomain      string `json:"custom_domain"`
	Password          string `json:"password"`
	NotificationEmail string `json:"notification_email"`

	State   string `json:"state"`
	Premium bool   `json:"premium"`
	Claimed bool   `json:"claimed"`

	Url           string `json:"url"`
	AdminUrl      string `json:"admin_url"`
	DeployUrl     string `json:"deploy_url"`
	ScreenshotUrl string `json:"screenshot_url"`

	CreatedAt Timestamp `json:"created_at"`
	UpdatedAt Timestamp `json:"updated_at"`

	Zip string
	Dir string

	client *Client
}

type DeployInfo struct {
	Id       string   `json:"id"`
	DeployId string   `json:"deploy_id"`
	Required []string `json:"required"`
}

type siteUpdate struct {
	Name              string             `json:"name"`
	CustomDomain      string             `json:"custom_domain"`
	Password          string             `json:"password"`
	NotificationEmail string             `json:"notification_email"`
	Files             *map[string]string `json:"files"`
}

func (s *SitesService) Get(id string) (*Site, error) {
	site := &Site{Id: id, client: s.client}
	err := site.refresh()

	return site, err
}

func (s *SitesService) List() ([]Site, error) {
	sites := new([]Site)

	_, err := s.client.Request("GET", "/sites", nil, sites)

	for _, site := range(*sites) {
		site.client = s.client
	}

	return *sites, err
}

func (site *Site) apiPath() string {
	return path.Join("/sites", site.Id)
}

func (site *Site) refresh() error {
	if site.Id == "" {
		return errors.New("Cannot fetch site without an ID")
	}
	_, err := site.client.Request("GET", site.apiPath(), nil, site)

	return err
}

func (site *Site) Update() error {

	if site.Zip != "" {
		return site.deployZip()
	} else {
		return site.deployDir()
	}

	options := &RequestOptions{JsonBody: site.mutableParams()}

	_, err := site.client.Request("PUT", site.apiPath(), options, site)

	return err
}

func (site *Site) WaitForReady(timeout time.Duration) error {
	if site.State == "current" {
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

			site, err := site.client.Sites.Get(site.Id)
			if site != nil {
				fmt.Println("Site state is now: ", site.State)
			}
			if err != nil || (site != nil && site.State == "current") {
				done <- err
				break
			}
		}
	}()

	err := <-done
	return err
}

func (site *Site) deployDir() error {
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
			Name:              site.Name,
			CustomDomain:      site.CustomDomain,
			Password:          site.Password,
			NotificationEmail: site.NotificationEmail,
			Files:             &files,
		},
	}

	deployInfo := new(DeployInfo)
	_, err = site.client.Request("PUT", site.apiPath(), options, deployInfo)

	if err != nil {
		return err
	}

	lookup := map[string]bool{}

	for _, sha := range deployInfo.Required {
		lookup[sha] = true
	}

	for path, sha := range files {
		if lookup[sha] == true {
			file, _ := os.Open(filepath.Join(site.Dir, path))
			defer file.Close()

			options = &RequestOptions{
				RawBody: file,
				Headers: &map[string]string{"Content-Type": "application/octet-stream"},
			}
			fmt.Println("Uploading %s", path)
			_, err = site.client.Request("PUT", filepath.Join(site.apiPath(), "files", path), options, nil)
			if err != nil {
				fmt.Println("Error", err)
				return err
			}
		}
	}

	return err
}

func (site *Site) deployZip() error {
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

	_, err = site.client.Request("PUT", site.apiPath(), options, nil)

	return err
}

func (site *Site) mutableParams() *map[string]string {
	return &map[string]string{
		"name":               site.Name,
		"custom_domain":      site.CustomDomain,
		"password":           site.Password,
		"notification_email": site.NotificationEmail,
	}
}
