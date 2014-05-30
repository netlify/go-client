package commands

import (
	"log"
	"github.com/spf13/cobra"
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
	client := newClient()

	var dir string
	if len(args) > 0 {
		dir = args[0]
	} else {
		dir = "."
	}
	path, err := filepath.Abs(dir)
	if err != nil {
		log.Fatalln("Bad directory path")
	}

	// Deploy
	log.Printf("Deploying site: %v - dir: %v", SiteId, path)

	site, _, err := client.Sites.Get(SiteId)

	if err != nil {
		log.Fatalf("Error during deploy: %v", err)
	}

	deploy, _, err := site.Deploys.Create(path)

	if err != nil {
		log.Fatalf("Deploy failed with error: ", err)
	}

	err = deploy.WaitForReady(0)
	if err != nil {
		log.Fatalf("Error dring site processing: ", err)
	}

	log.Println("Site deployed to", site.Url)
}
