package netlify

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff"
)

const MaxFilesForSyncDeploy = 7000
const PreProcessingTimeout = time.Minute * 5

// Deploy represents a specific deploy of a site
type Deploy struct {
	Id     string `json:"id"`
	SiteId string `json:"site_id"`
	UserId string `json:"user_id"`

	// State of the deploy (uploading/uploaded/processing/ready/error)
	State string `json:"state"`

	// Cause of error if State is "error"
	ErrorMessage string `json:"error_message"`

	// Shas of files that needs to be uploaded before the deploy is ready
	Required []string `json:"required"`

	DeployUrl     string `json:"deploy_url"`
	SiteUrl       string `json:"url"`
	ScreenshotUrl string `json:"screenshot_url"`

	CreatedAt Timestamp `json:"created_at"`
	UpdatedAt Timestamp `json:"updated_at"`

	client *Client
}

// DeploysService is used to access all Deploy related API methods
type DeploysService struct {
	site   *Site
	client *Client
}

type uploadError struct {
	err   error
	mutex *sync.Mutex
}

type deployFiles struct {
	Files *map[string]string `json:"files"`
	Async bool               `json:"async"`
}

func (s *DeploysService) apiPath() string {
	if s.site != nil {
		return path.Join(s.site.apiPath(), "deploys")
	} else {
		return "/deploys"
	}
}

// Create a new deploy
//
// Example: site.Deploys.Create("/path/to/site-dir", true)
// If the target is a zip file, it must have the extension .zip
func (s *DeploysService) Create(dirOrZip string) (*Deploy, *Response, error) {
	return s.create(dirOrZip, false)
}

// Create a new draft deploy. Draft deploys will be uploaded and processed, but
// won't affect the active deploy for a site.
func (s *DeploysService) CreateDraft(dirOrZip string) (*Deploy, *Response, error) {
	return s.create(dirOrZip, true)
}

func (s *DeploysService) create(dirOrZip string, draft bool) (*Deploy, *Response, error) {
	if s.site == nil {
		return nil, nil, errors.New("You can only create a new deploy for an existing site (site.Deploys.Create(dirOrZip)))")
	}

	params := url.Values{}
	if draft {
		params["draft"] = []string{"true"}
	}
	options := &RequestOptions{QueryParams: &params}
	deploy := &Deploy{client: s.client}
	resp, err := s.client.Request("POST", s.apiPath(), options, deploy)

	if err != nil {
		return deploy, resp, err
	}

	resp, err = deploy.Deploy(dirOrZip)
	return deploy, resp, err
}

// List all deploys. Takes ListOptions to control pagination.
func (s *DeploysService) List(options *ListOptions) ([]Deploy, *Response, error) {
	deploys := new([]Deploy)

	reqOptions := &RequestOptions{QueryParams: options.toQueryParamsMap()}

	resp, err := s.client.Request("GET", s.apiPath(), reqOptions, deploys)

	for _, deploy := range *deploys {
		deploy.client = s.client
	}

	return *deploys, resp, err
}

// Get a specific deploy.
func (d *DeploysService) Get(id string) (*Deploy, *Response, error) {
	deploy := &Deploy{Id: id, client: d.client}
	resp, err := deploy.Reload()

	return deploy, resp, err
}

func (deploy *Deploy) apiPath() string {
	return path.Join("/deploys", deploy.Id)
}

func (deploy *Deploy) Deploy(dirOrZip string) (*Response, error) {
	if strings.HasSuffix(dirOrZip, ".zip") {
		return deploy.deployZip(dirOrZip)
	} else {
		return deploy.deployDir(dirOrZip)
	}
}

// Reload a deploy from the API
func (deploy *Deploy) Reload() (*Response, error) {
	if deploy.Id == "" {
		return nil, errors.New("Cannot fetch deploy without an ID")
	}
	return deploy.client.Request("GET", deploy.apiPath(), nil, deploy)
}

// Restore an old deploy. Sets the deploy as the active deploy for a site
func (deploy *Deploy) Restore() (*Response, error) {
	return deploy.client.Request("POST", path.Join(deploy.apiPath(), "restore"), nil, deploy)
}

// Alias for restore. Published a specific deploy.
func (deploy *Deploy) Publish() (*Response, error) {
	return deploy.Restore()
}

func (deploy *Deploy) uploadFile(dir, path string, sharedError uploadError) error {
	sharedError.mutex.Lock()
	if sharedError.err != nil {
		sharedError.mutex.Unlock()
		return errors.New("Canceled because upload has already failed")
	}
	sharedError.mutex.Unlock()

	log.Printf("Uploading file: %v", path)
	file, err := os.Open(filepath.Join(dir, path))
	defer file.Close()

	if err != nil {
		log.Printf("Error opening file %v: %v", path, err)
		return err
	}

	info, err := file.Stat()

	if err != nil {
		log.Printf("Error getting file size %v: %v", path, err)
		return err
	}

	options := &RequestOptions{
		RawBody:       file,
		RawBodyLength: info.Size(),
		Headers:       &map[string]string{"Content-Type": "application/octet-stream"},
	}

	fileUrl, err := url.Parse(path)
	if err != nil {
		log.Printf("Error parsing url %v: %v", path, err)
		return err
	}

	resp, err := deploy.client.Request("PUT", filepath.Join(deploy.apiPath(), "files", fileUrl.Path), options, nil)
	if resp != nil && resp.Response != nil && resp.Body != nil {
		resp.Body.Close()
	}
	if err != nil {
		log.Printf("Error while uploading %v: %v", path, err)
		return err
	}

	return err
}

func (deploy *Deploy) deployDir(dir string) (*Response, error) {
	files := map[string]string{}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() == false && info.Mode().IsRegular() {
			rel, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}

			if strings.HasPrefix(rel, ".") || strings.Contains(rel, "/.") || strings.HasPrefix(rel, "__MACOS") {
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

	if err != nil {
		return nil, err
	}

	fileOptions := &deployFiles{
		Files: &files,
	}

	if len(files) > MaxFilesForSyncDeploy {
		fileOptions.Async = true
	}

	options := &RequestOptions{
		JsonBody: fileOptions,
	}

	resp, err := deploy.client.Request("PUT", deploy.apiPath(), options, deploy)

	if err != nil {
		return resp, err
	}

	if len(files) > MaxFilesForSyncDeploy {
		start := time.Now()
		for {
			resp, err := deploy.client.Request("GET", deploy.apiPath(), nil, deploy)
			if err != nil {
				log.Printf("Error fetching deploy: %v", err)
				time.Sleep(5 * time.Second)
			}
			resp.Body.Close()
			log.Printf("Deploy state: %v\n", deploy.State)
			if deploy.State == "prepared" || deploy.State == "ready" {
				break
			}
			if deploy.State == "error" {
				return resp, errors.New("Error: preprocessing deploy failed")
			}
			if start.Add(PreProcessingTimeout).Before(time.Now()) {
				return resp, errors.New("Error: preprocessing deploy timed out")
			}
			time.Sleep(2 * time.Second)
		}
	}

	lookup := map[string]bool{}

	for _, sha := range deploy.Required {
		lookup[sha] = true
	}

	// Use a channel as a semaphore to limit # of parallel uploads
	sem := make(chan int, deploy.client.MaxConcurrentUploads)
	var wg sync.WaitGroup

	sharedErr := uploadError{err: nil, mutex: &sync.Mutex{}}
	for path, sha := range files {
		if lookup[sha] == true && err == nil {
			sem <- 1
			wg.Add(1)
			go func(path string) {
				sharedErr.mutex.Lock()
				if sharedErr.err != nil {
					sharedErr.mutex.Unlock()
					<-sem
					wg.Done()
					return
				}
				sharedErr.mutex.Unlock()

				b := backoff.NewExponentialBackOff()
				b.MaxElapsedTime = 2 * time.Minute
				err := backoff.Retry(func() error { return deploy.uploadFile(dir, path, sharedErr) }, b)
				if err != nil {
					sharedErr.mutex.Lock()
					sharedErr.err = err
					sharedErr.mutex.Unlock()
					<-sem
					wg.Done()
					return
				}
				wg.Done()
				<-sem
			}(path)
		}
	}

	wg.Wait()

	if sharedErr.err != nil {
		return resp, sharedErr.err
	}

	return resp, err
}

func (deploy *Deploy) deployZip(zip string) (*Response, error) {
	zipPath, err := filepath.Abs(zip)
	if err != nil {
		return nil, err
	}

	zipFile, err := os.Open(zipPath)
	defer zipFile.Close()

	if err != nil {
		return nil, err
	}

	info, err := zipFile.Stat()

	if err != nil {
		return nil, err
	}

	options := &RequestOptions{
		RawBody:       zipFile,
		RawBodyLength: info.Size(),
		Headers:       &map[string]string{"Content-Type": "application/zip"},
	}

	resp, err := deploy.client.Request("PUT", deploy.apiPath(), options, deploy)

	return resp, err
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

			_, err := deploy.Reload()
			if err != nil || (deploy.State == "ready") {
				done <- err
				break
			}
		}
	}()

	err := <-done
	return err
}
