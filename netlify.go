package netlify

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"

	oauth "golang.org/x/oauth2"
)

const (
	libraryVersion = "0.1"
	defaultBaseURL = "https://api.netlify.com"
	apiVersion     = "v1"

	userAgent = "netlify-go/" + libraryVersion

	DefaultMaxConcurrentUploads = 10
)

// Config is used to configure the netlify client.
// Typically you'll just want to set an AccessToken
type Config struct {
	AccessToken string

	ClientId     string
	ClientSecret string

	BaseUrl   string
	UserAgent string

	HttpClient     *http.Client
	RequestTimeout time.Duration

	MaxConcurrentUploads int
}

func (c *Config) Token() (*oauth.Token, error) {
	return &oauth.Token{AccessToken: c.AccessToken}, nil
}

// The netlify Client
type Client struct {
	client *http.Client
	log    *logrus.Logger

	BaseUrl   *url.URL
	UserAgent string

	Sites      *SitesService
	Deploys    *DeploysService
	DeployKeys *DeployKeysService

	MaxConcurrentUploads int
}

// netlify API Response.
// All API methods on the different client services will return a Response object.
// For any list operation this object will hold pagination information
type Response struct {
	*http.Response

	NextPage  int
	PrevPage  int
	FirstPage int
	LastPage  int
}

// RequestOptions for doing raw requests to the netlify API
type RequestOptions struct {
	JsonBody      interface{}
	RawBody       io.Reader
	RawBodyLength int64
	QueryParams   *url.Values
	Headers       *map[string]string
}

// ErrorResponse is returned when a request to the API fails
type ErrorResponse struct {
	Response *http.Response
	Message  string
}

func (r *ErrorResponse) Error() string {
	return r.Message
}

// All List methods takes a ListOptions object controlling pagination
type ListOptions struct {
	Page    int
	PerPage int
}

func (o *ListOptions) toQueryParamsMap() *url.Values {
	params := url.Values{}
	if o != nil {
		if o.Page > 0 {
			params.Set("page", strconv.Itoa(o.Page))
		}
		if o.PerPage > 0 {
			params.Set("per_page", strconv.Itoa(o.PerPage))
		}
	}
	return &params
}

// NewClient returns a new netlify API client
func NewClient(config *Config) *Client {
	client := &Client{}

	if config.BaseUrl != "" {
		client.BaseUrl, _ = url.Parse(config.BaseUrl)
	} else {
		client.BaseUrl, _ = url.Parse(defaultBaseURL)
	}

	if config.HttpClient != nil {
		client.client = config.HttpClient
	} else if config.AccessToken != "" {
		client.client = oauth.NewClient(oauth.NoContext, config)
		if config.RequestTimeout > 0 {
			client.client.Timeout = config.RequestTimeout
		}
	}

	if &config.UserAgent != nil {
		client.UserAgent = config.UserAgent
	} else {
		client.UserAgent = userAgent
	}

	if config.MaxConcurrentUploads != 0 {
		client.MaxConcurrentUploads = config.MaxConcurrentUploads
	} else {
		client.MaxConcurrentUploads = DefaultMaxConcurrentUploads
	}

	logrus.SetOutput(ioutil.Discard)
	client.log = logrus.StandardLogger()
	client.Sites = &SitesService{client: client}
	client.Deploys = &DeploysService{client: client}

	client.log.WithFields(logrus.Fields{
		"base_url":               client.BaseUrl.String(),
		"user_agent":             client.UserAgent,
		"max_concurrent_uploads": client.MaxConcurrentUploads,
	}).Debug("created client")

	return client
}

func (c *Client) SetLogger(log *logrus.Logger) {
	if log != nil {
		c.log = logrus.StandardLogger()
	}
	c.log = log
}

func (c *Client) newRequest(method, apiPath string, options *RequestOptions) (*http.Request, error) {
	if c.client == nil {
		return nil, errors.New("Client has not been authenticated")
	}

	urlPath := path.Join("api", apiVersion, apiPath)
	if options != nil && options.QueryParams != nil && len(*options.QueryParams) > 0 {
		urlPath = urlPath + "?" + options.QueryParams.Encode()
	}
	rel, err := url.Parse(urlPath)
	if err != nil {
		return nil, err
	}

	u := c.BaseUrl.ResolveReference(rel)

	buf := new(bytes.Buffer)

	if options != nil && options.JsonBody != nil {
		err := json.NewEncoder(buf).Encode(options.JsonBody)
		if err != nil {
			return nil, err
		}
	}

	var req *http.Request

	if options != nil && options.RawBody != nil {
		req, err = http.NewRequest(method, u.String(), options.RawBody)
		req.ContentLength = options.RawBodyLength
	} else {
		req, err = http.NewRequest(method, u.String(), buf)
	}
	if err != nil {
		return nil, err
	}

	req.Close = true

	req.TransferEncoding = []string{"identity"}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", c.UserAgent)

	if options != nil && options.JsonBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if options != nil && options.Headers != nil {
		for key, value := range *options.Headers {
			req.Header.Set(key, value)
		}
	}

	return req, nil
}

// Request sends an authenticated HTTP request to the netlify API
//
// When error is nil, resp always contains a non-nil Response object
//
// Generally methods on the various services should be used over raw API requests
func (c *Client) Request(method, path string, options *RequestOptions, decodeTo interface{}) (*Response, error) {
	var httpResponse *http.Response
	req, err := c.newRequest(method, path, options)
	if err != nil {
		return nil, err
	}

	if c.idempotent(req) && (options == nil || options.RawBody == nil) {
		httpResponse, err = c.doWithRetry(req, 3)
	} else {
		httpResponse, err = c.client.Do(req)
	}

	resp := newResponse(httpResponse)

	if err != nil {
		return resp, err
	}

	if err = checkResponse(httpResponse); err != nil {
		return resp, err
	}

	if decodeTo != nil {
		defer httpResponse.Body.Close()
		if writer, ok := decodeTo.(io.Writer); ok {
			io.Copy(writer, httpResponse.Body)
		} else {
			err = json.NewDecoder(httpResponse.Body).Decode(decodeTo)
		}
	}
	return resp, err
}

func (c *Client) idempotent(req *http.Request) bool {
	switch req.Method {
	case "GET", "PUT", "DELETE":
		return true
	default:
		return false
	}
}

func (c *Client) rewindRequestBody(req *http.Request) error {
	if req.Body == nil {
		return nil
	}
	body, ok := req.Body.(io.Seeker)
	if ok {
		_, err := body.Seek(0, 0)
		return err
	}
	return errors.New("Body is not a seeker")
}

func (c *Client) doWithRetry(req *http.Request, tries int) (*http.Response, error) {
	httpResponse, err := c.client.Do(req)

	tries--

	if tries > 0 && (err != nil || httpResponse.StatusCode >= 400) {
		if err := c.rewindRequestBody(req); err != nil {
			return c.doWithRetry(req, tries)
		}
	}

	return httpResponse, err
}

func newResponse(r *http.Response) *Response {
	response := &Response{Response: r}
	if r != nil {
		response.populatePageValues()
	}
	return response
}

func checkResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}
	errorResponse := &ErrorResponse{Response: r}
	if r.StatusCode == 403 || r.StatusCode == 401 {
		errorResponse.Message = "Access Denied"
		return errorResponse
	}

	data, err := ioutil.ReadAll(r.Body)
	if err == nil && data != nil {
		errorResponse.Message = string(data)
	} else {
		errorResponse.Message = r.Status
	}

	return errorResponse
}

// populatePageValues parses the HTTP Link response headers and populates the
// various pagination link values in the Reponse.
func (r *Response) populatePageValues() {
	if links, ok := r.Response.Header["Link"]; ok && len(links) > 0 {
		for _, link := range strings.Split(links[0], ",") {
			segments := strings.Split(strings.TrimSpace(link), ";")

			// link must at least have href and rel
			if len(segments) < 2 {
				continue
			}

			// ensure href is properly formatted
			if !strings.HasPrefix(segments[0], "<") || !strings.HasSuffix(segments[0], ">") {
				continue
			}

			// try to pull out page parameter
			url, err := url.Parse(segments[0][1 : len(segments[0])-1])
			if err != nil {
				continue
			}
			page := url.Query().Get("page")
			if page == "" {
				continue
			}

			for _, segment := range segments[1:] {
				switch strings.TrimSpace(segment) {
				case `rel="next"`:
					r.NextPage, _ = strconv.Atoi(page)
				case `rel="prev"`:
					r.PrevPage, _ = strconv.Atoi(page)
				case `rel="first"`:
					r.FirstPage, _ = strconv.Atoi(page)
				case `rel="last"`:
					r.LastPage, _ = strconv.Atoi(page)
				}

			}
		}
	}
}
