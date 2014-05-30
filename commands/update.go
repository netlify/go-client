package commands

import (
  "log"
/*  "github.com/bitballoon/bitballoon-go/bitballoon"*/
  "github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
  Use:   "update",
  Short: "update attributes of a BitBalloon site",
  Long:  "updates name, domain, password or email",
}

type updateString struct {
  value string
  set bool
}

func (u *updateString) Set(value string) error {
  u.value = value
  u.set = true
  return nil
}

func (u *updateString) String() string {
  return u.value
}

var updateName, updateDomain, updatePassword, updateEmail updateString

func init() {
  updateCmd.Run = update

  updateCmd.Flags().VarP(&updateName, "name", "n", "Name of the site (must be a valid subdomain: <name>.bitballoon.com)")
  updateCmd.Flags().VarP(&updateDomain, "domain", "d", "Custom domain for the site (only works for premium sites)")
  updateCmd.Flags().VarP(&updatePassword, "password", "", "Password for the site")
  updateCmd.Flags().VarP(&updateEmail, "email", "e", "Notification email for form submissions (only works for premium sites)")
}

func update(cmd *cobra.Command, args []string) {
    client := newClient()

    if SiteId == "" {
      log.Fatalln("No site id specified. Use the --site options")
    }

    site, _, err := client.Sites.Get(SiteId)

    if err != nil {
      log.Fatalf("Error updating site: %v", err)
    }

    if updateName.set {
      site.Name = updateName.value
    }

    if updateDomain.set {
      site.CustomDomain = updateDomain.value
    }

    if updatePassword.set {
      site.Password = updatePassword.value
    }

    if updateEmail.set {
      site.NotificationEmail = updateEmail.value
    }

    _, err = site.Update()

    if err != nil {
      log.Fatalf("Error updating site: %v", err)
    }

    log.Printf("Site updated. URL: %v", site.Url)
}
