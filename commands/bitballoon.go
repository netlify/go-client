package commands

import (
	"os"
	"log"
	"github.com/spf13/cobra"
	"github.com/bitballoon/bitballoon-go/bitballoon"
)

var SiteId, AccessToken string

var BitBalloonCmd = &cobra.Command{
	Use:   "bitballoon",
	Short: "",
	Long:  "",
}

func Execute() {
	AddCommands()
	BitBalloonCmd.Execute()
}

func AddCommands() {
	BitBalloonCmd.AddCommand(createCmd)
	BitBalloonCmd.AddCommand(updateCmd)
	BitBalloonCmd.AddCommand(deployCmd)
}

func init() {
	BitBalloonCmd.PersistentFlags().StringVarP(&SiteId, "site", "s", "", "site domain or id")
	BitBalloonCmd.PersistentFlags().StringVarP(&AccessToken, "token", "t", "", "API acccess token (https://www.bitballoon.com/applications)")
}

func newClient() *bitballoon.Client {
	if AccessToken == "" {
		log.Fatalln("No API access token, get one at https://www.bitballoon.com/applications and use the --token option")
	}

	config := &bitballoon.Config{AccessToken: AccessToken}

	endpoint := os.Getenv("BB_API_ENDPOINT")

	if endpoint != "" {
		config.BaseUrl = endpoint
	}

	return bitballoon.NewClient(config)
}
