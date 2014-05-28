package bitballoon

import (
	"bytes"
	"code.google.com/p/goauth2/oauth"
	"encoding/json"
	"strings"
	"strconv"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
)

const (
	libraryVersion = "0.1"
	defaultBaseURL = "https://www.bitballoon.com"
	apiVersion     = "v1"

	userAgent = "bitballoon-go/" + libraryVersion
)

// Config is used to configure the BitBalloon client.
// Typically you'll just want to set an AccessToken
type Config struct {
	AccessToken  string

	ClientId     string
	ClientSecret string

	BaseUrl      string
	UserAgent    string

	HttpClient   *http.Client
}

// The BitBalloon Client
type Client struct {
	client    *http.Client

	BaseUrl   *url.URL
	UserAgent string

	Sites   *SitesService
	Deploys *DeployService
}

// BitBalloon API Response.
// All API methods on the different client services will return a Response object.
// For any list operation this object will hold pagination information
type Response struct {
	*http.Response

	NextPage  int
	PrevPage  int
	FirstPage int
	LastPage  int
}

// RequestOptions for doing raw requests to the BitBalloon API
type RequestOptions struct {
	JsonBody    interface{}
	RawBody     io.Reader
	QueryParams *url.Values
	Headers     *map[string]string
}

// ErrorResponse is returned when a request to the API fails
type ErrorResponse struct {
	Response *http.Response
	Message  string
}

// All List methods takes a ListOptions object controlling pagination
type ListOptions struct {
	Page int
	PerPage int
}

func (o *ListOptions) toQueryParamsMap() *url.Values {
	params := url.Values{}
	if o.Page > 0 {
		params["page"] = []string{strconv.Itoa(o.Page)}
	}
	if o.PerPage > 0 {
		params["per_page"] = []string{strconv.Itoa(o.PerPage)}
	}
	return &params
}

func (r *ErrorResponse) Error() string {
	return r.Message
}

// NewClient returns a new BitBalloon API client
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
		t := &oauth.Transport{
			Token: &oauth.Token{AccessToken: config.AccessToken},
		}
		client.client = t.Client()
	}

	if &config.UserAgent != nil {
		client.UserAgent = config.UserAgent
	} else {
		client.UserAgent = userAgent
	}

	client.Sites = &SitesService{client: client}
	client.Deploys = &DeployService{client: client}

	return client
}

func (c *Client) newRequest(method, apiPath string, options *RequestOptions) (*http.Request, error) {
	if c.client == nil {
		return nil, errors.New("Client has not been authenticated")
	}

	urlPath := path.Join("api", apiVersion, apiPath)
	if options!= nil && options.QueryParams != nil && len(*options.QueryParams) > 0 {
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
	} else {
		req, err = http.NewRequest(method, u.String(), buf)
	}
	if err != nil {
		return nil, err
	}

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

// Request sends an authenticated HTTP request to the BitBalloon API
//
// When error is nil, resp always contains a non-nil Response object
//
// Generally methods on the various services should be used over raw API requests
func (c *Client) Request(method, path string, options *RequestOptions, decodeTo interface{}) (*Response, error) {
	req, err := c.newRequest(method, path, options)
	if err != nil {
		return nil, err
	}

	httpResponse, err := c.client.Do(req)
	defer httpResponse.Body.Close()

	resp := newResponse(httpResponse)

	if err != nil {
		return resp, err
	}

	if err = checkResponse(httpResponse); err != nil {
		return resp, err
	}

	if decodeTo != nil {
		if writer, ok := decodeTo.(io.Writer); ok {
			io.Copy(writer, httpResponse.Body)
		} else {
			err = json.NewDecoder(httpResponse.Body).Decode(decodeTo)
		}
	}
	return resp, err
}

func newResponse(r *http.Response) *Response {
	response := &Response{Response: r}
	response.populatePageValues()
	return response
}

func checkResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}
	errorResponse := &ErrorResponse{Response: r}
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
