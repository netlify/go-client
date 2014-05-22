package bitballoon

import(
  "io"
  "path"
  "bytes"
  "errors"
  "net/http"
	"net/url"
  "encoding/json"
  "code.google.com/p/goauth2/oauth"
)

const (
	libraryVersion = "0.1"
	defaultBaseURL = "https://www.bitballoon.com"
  apiVersion     = "v1"

	userAgent      = "bitballoon-go/" + libraryVersion
)

type Config struct {
  ClientId string
  ClientSecret string
  AccessToken string
  BaseUrl string
  UserAgent string
}

type Client struct {
  client *http.Client
  BaseUrl *url.URL
  UserAgent string

  Sites *SitesService
}

type Response struct {
	*http.Response

	NextPage  int
	PrevPage  int
	FirstPage int
	LastPage  int
}

type RequestOptions struct {
  JsonBody interface{}
  RawBody io.Reader
  QueryParams *map[string]string
  Headers *map[string]string
}

func NewClient(config *Config) *Client {
  client := &Client{}

  if &config.BaseUrl != nil {
    client.BaseUrl, _ = url.Parse(config.BaseUrl)
  } else {
    client.BaseUrl, _ = url.Parse(defaultBaseURL)
  }

  if &config.AccessToken != nil {
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

  return client
}

func (c *Client) newRequest(method, apiPath string, options *RequestOptions)  (*http.Request, error) {
  if c.client == nil {
    return nil, errors.New("Client has not been authenticated")
  }

  rel, err := url.Parse(path.Join("api", apiVersion, apiPath))
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

  if options != nil && options.Headers != nil {
    for key, value := range *options.Headers {
      req.Header.Set(key, value)
    }
  }

	return req, nil
}

func (c *Client) Request(method, path string, options *RequestOptions, decodeTo interface{}) (*Response, error) {
  req, err := c.newRequest(method, path, options)
  if err != nil {
    return nil, err
  }

  httpResponse, err := c.client.Do(req)
  defer httpResponse.Body.Close()

  resp := &Response{Response: httpResponse}

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

func checkResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}
	return errors.New("API Error")
}
