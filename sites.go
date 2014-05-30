package bitballoon

import (
	"errors"
	"path"
	"time"
)

var (
	defaultTimeout time.Duration = 5 * 60 // 5 minutes
)

// SitesService is used to access all Site related API methods
type SitesService struct {
	client *Client
}

// Site represents a BitBalloon Site
type Site struct {
	Id     string `json:"id"`
	UserId string `json:"user_id"`

	// These fields can be updated through the API
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

	// Access deploys for this site
	Deploys *DeploysService

	client *Client
}

// Info returned when creating a new deploy
type DeployInfo struct {
	Id       string   `json:"id"`
	DeployId string   `json:"deploy_id"`
	Required []string `json:"required"`
}

// Attributes for Sites.Create
type SiteAttributes struct {
	Name              string `json:"name"`
	CustomDomain      string `json:"custom_domain"`
	Password          string `json:"password"`
	NotificationEmail string `json:"notification_email"`
}

// Get a single Site from the API. The id can be either a site Id or the domain
// of a site (ie. site.Get("mysite.bitballoon.com"))
func (s *SitesService) Get(id string) (*Site, *Response, error) {
	site := &Site{Id: id, client: s.client}
	site.Deploys = &DeploysService{client: s.client, site: site}
	resp, err := site.Reload()

	return site, resp, err
}

// Create a new empty site.
func (s *SitesService) Create(attributes *SiteAttributes) (*Site, *Response, error) {
	site := &Site{client: s.client}
	site.Deploys = &DeploysService{client: s.client, site: site}

	reqOptions := &RequestOptions{JsonBody: attributes}

	resp, err := s.client.Request("POST", "/sites", reqOptions, site)

	return site, resp, err
}

// List all sites you have access to. Takes ListOptions to control pagination.
func (s *SitesService) List(options *ListOptions) ([]Site, *Response, error) {
	sites := new([]Site)

	reqOptions := &RequestOptions{QueryParams: options.toQueryParamsMap()}

	resp, err := s.client.Request("GET", "/sites", reqOptions, sites)

	for _, site := range *sites {
		site.client = s.client
		site.Deploys = &DeploysService{client: s.client, site: &site}
	}

	return *sites, resp, err
}

func (site *Site) apiPath() string {
	return path.Join("/sites", site.Id)
}

func (site *Site) Reload() (*Response, error) {
	if site.Id == "" {
		return nil, errors.New("Cannot fetch site without an ID")
	}
	return site.client.Request("GET", site.apiPath(), nil, site)
}

// Update will update the fields that can be updated through the API
func (site *Site) Update() (*Response, error) {
	options := &RequestOptions{JsonBody: site.mutableParams()}

	return site.client.Request("PUT", site.apiPath(), options, site)
}

// Destroy deletes a site permanently
func (site *Site) Destroy() (*Response, error) {
	return site.client.Request("DELETE", site.apiPath(), nil, nil)
}

func (site *Site) mutableParams() *map[string]string {
	return &map[string]string{
		"name":               site.Name,
		"custom_domain":      site.CustomDomain,
		"password":           site.Password,
		"notification_email": site.NotificationEmail,
	}
}
