package commands

import (
  "log"
  "github.com/BitBalloon/bitballoon-go/bitballoon"
  "github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
  Use:   "create",
  Short: "create a new BitBalloon site",
  Long:  "creates a new site and returns an ID you can deploy to",
}

var siteName, customDomain, password, notificationEmail string

func init() {
  createCmd.Run = create

  createCmd.Flags().StringVarP(&siteName, "name", "n", "", "Name of the site (must be a valid subdomain: <name>.bitballoon.com)")
  createCmd.Flags().StringVarP(&customDomain, "domain", "d", "", "Custom domain for the site (only works for premium sites)")
  createCmd.Flags().StringVarP(&password, "password", "p", "", "Password for the site")
  createCmd.Flags().StringVarP(&notificationEmail, "email", "e", "", "Notification email for form submissions (only works for premium sites)")
}

func create(cmd *cobra.Command, args []string) {
  client := newClient()
  site, _, err := client.Sites.Create(&bitballoon.SiteAttributes{
    Name: siteName,
    CustomDomain: customDomain,
    Password: password,
    NotificationEmail: notificationEmail,
  })

  if err != nil {
    log.Fatalf("Error creating site: %v", err)
  }

  log.Println("Site created")
  log.Printf("URL: %v", site.Url)
  log.Printf("ID: %v", site.Id)
}
