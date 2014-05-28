package commands

import (
	"fmt"
	"github.com/bitballoon/bitballoon-go/bitballoon"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "deploy a site to BitBalloon",
	Long:  "deploys an existing site or creates a new site",
}

func init() {
	deployCmd.Run = deploy
}

func deploy(cmd *cobra.Command, args []string) {
	if AccessToken == "" {
		fmt.Println("No API access token, get one at https://www.bitballoon.com/applications and use the --token option")
		return
	}

	var dir string
	if len(args) > 0 {
		dir = args[0]
	} else {
		dir = "."
	}
	path, err := filepath.Abs(dir)
	if err != nil {
		fmt.Println("Bad directory path")
		return
	}

	// Deploy
	fmt.Println("Deploying site: %v - dir: %v", SiteId, path)

	config := &bitballoon.Config{AccessToken: AccessToken}

	endpoint := os.Getenv("BB_API_ENDPOINT")

	if endpoint != "" {
		config.BaseUrl = endpoint
	}

	client := bitballoon.NewClient(config)
	site, _, err := client.Sites.Get(SiteId)

	if err != nil {
		fmt.Println("Error during deploy: %v", err)
		return
	}

	deploy, _, err := site.Deploys.Create(path)

	if err != nil {
		fmt.Println("Deploy failed with error: ", err)
		return
	}

	err = deploy.WaitForReady(0)
	if err != nil {
		fmt.Println("Error dring site processing: ", err)
		return
	}

	fmt.Println("Site deployed to", site.Url)
}
