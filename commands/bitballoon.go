package commands

import (
	"github.com/spf13/cobra"
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
	BitBalloonCmd.AddCommand(deployCmd)
}

func init() {
	BitBalloonCmd.PersistentFlags().StringVarP(&SiteId, "site", "s", "", "site domain or id")
	BitBalloonCmd.PersistentFlags().StringVarP(&AccessToken, "token", "t", "", "API acccess token (https://www.bitballoon.com/applications)")
}
