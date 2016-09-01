package netlify

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/cenkalti/backoff"
)

const MaxFilesForSyncDeploy = 1000
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

	Branch    string `json:"branch,omitempty"`
	CommitRef string `json:"commit_ref,omitempty"`

	client *Client
	logger *logrus.Entry
}

func (d Deploy) log() *logrus.Entry {
	if d.logger == nil {
		d.logger = d.client.log.WithFields(logrus.Fields{
			"function": "deploy",
			"id":       d.Id,
			"site_id":  d.SiteId,
			"user_id":  d.UserId,
		})
	}

	return d.logger.WithField("state", d.State)
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

func (u *uploadError) Set(err error) {
	if err != nil {
		u.mutex.Lock()
		defer u.mutex.Unlock()
		if u.err != nil {
			u.err = err
		}
	}
}

func (u *uploadError) Get() error {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	return u.err
}

type deployFiles struct {
	Files     *map[string]string `json:"files"`
	Async     bool               `json:"async"`
	Branch    string             `json:"branch,omitempty"`
	CommitRef string             `json:"commit_ref,omitempty"`
}

func (s *DeploysService) apiPath() string {
	if s.site != nil {
		return path.Join(s.site.apiPath(), "deploys")
	}
	return "/deploys"
}

// Create a new deploy
//
// Example: site.Deploys.Create("/path/to/site-dir", true)
// If the target is a zip file, it must have the extension .zip
func (s *DeploysService) Create(dirOrZip string) (*Deploy, *Response, error) {
	return s.create(dirOrZip, false)
}

// CreateDraft a new draft deploy. Draft deploys will be uploaded and processed, but
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
	if sharedError.Get() != nil {
		return errors.New("Canceled because upload has already failed")
	}

	log := deploy.log().WithFields(logrus.Fields{
		"dir":  dir,
		"path": path,
	})

	log.Infof("Uploading file: %v", path)
	file, err := os.Open(filepath.Join(dir, path))
	defer file.Close()

	if err != nil {
		log.Warnf("Error opening file %v: %v", path, err)
		return err
	}

	info, err := file.Stat()

	if err != nil {
		log.Warnf("Error getting file size %v: %v", path, err)
		return err
	}

	options := &RequestOptions{
		RawBody:       file,
		RawBodyLength: info.Size(),
		Headers:       &map[string]string{"Content-Type": "application/octet-stream"},
	}

	fileUrl, err := url.Parse(path)
	if err != nil {
		log.Warnf("Error parsing url %v: %v", path, err)
		return err
	}

	resp, err := deploy.client.Request("PUT", filepath.Join(deploy.apiPath(), "files", fileUrl.Path), options, nil)
	if resp != nil && resp.Response != nil && resp.Body != nil {
		resp.Body.Close()
	}
	if err != nil {
		log.Warnf("Error while uploading %v: %v", path, err)
		return err
	}

	log.Infof("Finished uploading file: %s", path)
	return err
}

// deployDir scans the given directory and deploys the files
// that have changed on Netlify.
func (deploy *Deploy) deployDir(dir string) (*Response, error) {
	return deploy.DeployDirWithGitInfo(dir, "", "")
}

// DeployDirWithGitInfo scans the given directory and deploys the files
// that have changed on Netlify.
//
// This function allows you to supply git information about the deploy
// when it hasn't been set previously be a Continuous Deployment process.
func (deploy *Deploy) DeployDirWithGitInfo(dir, branch, commitRef string) (*Response, error) {
	files := map[string]string{}
	log := deploy.log().WithFields(logrus.Fields{
		"dir":        dir,
		"branch":     branch,
		"commit_ref": commitRef,
	})
	defer log.Infof("Finished deploying directory %s", dir)

	log.Infof("Starting deploy of directory %s", dir)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() == false && info.Mode().IsRegular() {
			rel, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}

			if ignoreFile(rel) {
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
		log.WithError(err).Warn("Failed to walk directory structure")
		return nil, err
	}

	fileOptions := &deployFiles{
		Files:     &files,
		Branch:    branch,
		CommitRef: commitRef,
	}

	if len(files) > MaxFilesForSyncDeploy {
		log.Debugf("More files than sync can deploy %d vs %d", len(files), MaxFilesForSyncDeploy)
		fileOptions.Async = true
	}

	options := &RequestOptions{
		JsonBody: fileOptions,
	}

	log.Debug("Starting to do PUT to origin")
	resp, err := deploy.client.Request("PUT", deploy.apiPath(), options, deploy)
	if err != nil {
		return resp, err
	}

	if len(files) > MaxFilesForSyncDeploy {
		start := time.Now()
		log.Debug("Starting to poll for the deploy to get into ready || prepared state")
		for {
			resp, err := deploy.client.Request("GET", deploy.apiPath(), nil, deploy)
			if err != nil {
				log.WithError(err).Warnf("Error fetching deploy, waiting for 5 seconds before retry: %v", err)
				time.Sleep(5 * time.Second)
			}
			resp.Body.Close()

			log.Debugf("Deploy state: %v\n", deploy.State)
			if deploy.State == "prepared" || deploy.State == "ready" {
				break
			}
			if deploy.State == "error" {
				log.Warnf("deploy is in state error")
				return resp, errors.New("Error: preprocessing deploy failed")
			}
			if start.Add(PreProcessingTimeout).Before(time.Now()) {
				log.Warnf("Deploy timed out waiting for preprocessing")
				return resp, errors.New("Error: preprocessing deploy timed out")
			}
			log.Debug("Waiting for 2 seconds to retry getting deploy")
			time.Sleep(2 * time.Second)
		}
	}

	lookup := map[string]bool{}

	for _, sha := range deploy.Required {
		lookup[sha] = true
	}

	log.Infof("Going to deploy the %d required files", len(lookup))

	// Use a channel as a semaphore to limit # of parallel uploads
	sem := make(chan int, deploy.client.MaxConcurrentUploads)
	var wg sync.WaitGroup

	sharedErr := uploadError{err: nil, mutex: &sync.Mutex{}}
	for path, sha := range files {
		if lookup[sha] == true && err == nil {
			sem <- 1
			go func(path string) {
				wg.Add(1)
				defer func() {
					<-sem
					wg.Done()
				}()
				log.Debugf("Starting to upload %s/%s", path, sha)
				if sharedErr.Get() != nil {
					return
				}

				b := backoff.NewExponentialBackOff()
				b.MaxElapsedTime = 2 * time.Minute
				err := backoff.Retry(func() error { return deploy.uploadFile(dir, path, sharedErr) }, b)
				if err != nil {
					sharedErr.Set(err)
				}
			}(path)
		}
	}

	log.Debugf("Waiting for required files to upload")
	wg.Wait()

	if sharedErr.Get() != nil {
		return resp, sharedErr.err
	}

	return resp, err
}

// deployZip uploads a Zip file to Netlify and deploys the files
// that have changed.
func (deploy *Deploy) deployZip(zip string) (*Response, error) {
	log := deploy.log().WithFields(logrus.Fields{
		"function": "zip",
		"zip_path": zip,
	})
	log.Infof("Starting to deploy zip file %s", zip)
	zipPath, err := filepath.Abs(zip)
	if err != nil {
		return nil, err
	}

	log.Debugf("Opening zip file at %s", zipPath)
	zipFile, err := os.Open(zipPath)
	if err != nil {
		return nil, err
	}
	defer zipFile.Close()

	info, err := zipFile.Stat()
	if err != nil {
		return nil, err
	}

	log.WithFields(logrus.Fields{
		"name": info.Name(),
		"size": info.Size(),
		"mode": info.Mode(),
	}).Debugf("Opened file %s of %s bytes", info.Name(), info.Size())

	options := &RequestOptions{
		RawBody:       zipFile,
		RawBodyLength: info.Size(),
		Headers:       &map[string]string{"Content-Type": "application/zip"},
	}

	log.Debug("Excuting PUT request for zip file")
	resp, err := deploy.client.Request("PUT", deploy.apiPath(), options, deploy)
	if err != nil {
		log.WithError(err).Warn("Error while uploading zip file")
	}

	log.Info("Finished uploading zip file")
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

func ignoreFile(rel string) bool {
	if strings.HasPrefix(rel, ".") || strings.Contains(rel, "/.") || strings.HasPrefix(rel, "__MACOS") {
		if strings.HasPrefix(rel, ".well-known/") {
			return false
		}
		return true
	}
	return false
}
