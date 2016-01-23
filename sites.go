package netlify

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

// Site represents a netlify Site
type Site struct {
	Id     string `json:"id"`
	UserId string `json:"user_id"`

	// These fields can be updated through the API
	Name              string   `json:"name"`
	CustomDomain      string   `json:"custom_domain"`
	DomainAliases     []string `json:"domain_aliases"`
	Password          string   `json:"password"`
	NotificationEmail string   `json:"notification_email"`

	State   string `json:"state"`
	Plan    string `json:"plan"`
	SSLPlan string `json:"ssl_plan"`
	Premium bool   `json:"premium"`
	Claimed bool   `json:"claimed"`

	Url           string `json:"url"`
	AdminUrl      string `json:"admin_url"`
	DeployUrl     string `json:"deploy_url"`
	ScreenshotUrl string `json:"screenshot_url"`

	SSL      bool `json:"ssl"`
	ForceSSL bool `json:"force_ssl"`

	BuildSettings      *BuildSettings      `json:"build_settings"`
	ProcessingSettings *ProcessingSettings `json:"processing_settings"`

	DeployHook string `json:"deploy_hook"`

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

// Settings for continuous deployment
type BuildSettings struct {
	RepoType   string            `json:"repo_type"`
	RepoURL    string            `json:"repo_url"`
	RepoBranch string            `json:"repo_branch"`
	Cmd        string            `json:"cmd"`
	Dir        string            `json:"dir"`
	Env        map[string]string `json:"env"`

	CreatedAt Timestamp `json:"created_at"`
	UpdatedAt Timestamp `json:"updated_at"`
}

// Settings for post processing
type ProcessingSettings struct {
	CSS struct {
		Minify bool `json:"minify"`
		Bundle bool `json:"bundle"`
	} `json:"css"`
	JS struct {
		Minify bool `json:"minify"`
		Bundle bool `json:"bundle"`
	} `json:"js"`
	HTML struct {
		PrettyURLs bool `json:"pretty_urls"`
	} `json:"html"`
	Images struct {
		Optimize bool `json:"optimize"`
	} `json:"images"`
	Skip bool `json:"skip"`
}

// Attributes for Sites.Create
type SiteAttributes struct {
	Name              string `json:"name"`
	CustomDomain      string `json:"custom_domain"`
	Password          string `json:"password"`
	NotificationEmail string `json:"notification_email"`

	ForceSSL bool `json:"force_ssl"`

	ProcessingSettings bool `json:"processing_options"`

	Repo *RepoOptions `json:"repo"`
}

// Attributes for site.ProvisionCert
type CertOptions struct {
	Certificate    string   `json:"certificate"`
	Key            string   `json:"key"`
	CaCertificates []string `json:"ca_certificates"`
}

// Attributes for setting up continuous deployment
type RepoOptions struct {
	// GitHub API ID or similar unique repo ID
	Id string `json:"id"`

	// Repo path. Full ssh based path for manual repos,
	// username/reponame for GitHub or BitBucket
	Repo string `json:"repo"`

	// Currently "github", "bitbucket" or "manual"
	Provider string `json:"provider"`

	// Directory to deploy after building
	Dir string `json:"dir"`

	// Build command
	Cmd string `json:"cmd"`

	// Branch to pull from
	Branch string `json:"branch"`

	// Build environment variables
	Env *map[string]string `json:"env"`

	// ID of a netlify deploy key used to access the repo
	DeployKeyID string `json:"deploy_key_id"`
}

// Get a single Site from the API. The id can be either a site Id or the domain
// of a site (ie. site.Get("mysite.netlify.com"))
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

// Configure Continuous Deployment for a site
func (site *Site) ContinuousDeployment(repoOptions *RepoOptions) (*Response, error) {
	options := &RequestOptions{JsonBody: map[string]*RepoOptions{"repo": repoOptions}}

	return site.client.Request("PUT", site.apiPath(), options, site)
}

// Provision SSL Certificate for a site. Takes optional CertOptions to set a custom cert/chain/key.
// Without this netlify will generate the certificate automatically.
func (site *Site) ProvisionCert(certOptions *CertOptions) (*Response, error) {
	options := &RequestOptions{JsonBody: certOptions}

	return site.client.Request("POST", path.Join(site.apiPath(), "ssl"), options, nil)
}

// Destroy deletes a site permanently
func (site *Site) Destroy() (*Response, error) {
	resp, err := site.client.Request("DELETE", site.apiPath(), nil, nil)
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
	return resp, err
}

func (site *Site) mutableParams() *SiteAttributes {
	return &SiteAttributes{
		Name:              site.Name,
		CustomDomain:      site.CustomDomain,
		Password:          site.Password,
		NotificationEmail: site.NotificationEmail,
		ForceSSL:          site.ForceSSL,
	}
}
