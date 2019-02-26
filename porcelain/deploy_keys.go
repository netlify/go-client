package porcelain

import (
	"github.com/netlify/go-client/models"
	"github.com/netlify/go-client/plumbing/operations"
	"github.com/netlify/go-client/porcelain/context"
)

func (n *Netlify) CreateDeployKey(ctx context.Context) (*models.DeployKey, error) {
	authInfo := context.GetAuthInfo(ctx)
	params := operations.NewCreateDeployKeyParams()
	resp, err := n.Netlify.Operations.CreateDeployKey(params, authInfo)
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}
