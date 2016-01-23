package netlify

// DeployKey for use with continuous deployment setups
type DeployKey struct {
	Id        string `json:"id"`
	PublicKey string `json:"public_key"`

	CreatedAt Timestamp `json:"created_at"`
}

// DeployKeysService is used to access all DeployKey related API methods
type DeployKeysService struct {
	site   *Site
	client *Client
}

// Create a new deploy key for use with continuous deployment
func (d *DeployKeysService) Create() (*DeployKey, *Response, error) {
	deployKey := &DeployKey{}

	resp, err := d.client.Request("POST", "/deploy_keys", &RequestOptions{}, deployKey)

	return deployKey, resp, err
}
