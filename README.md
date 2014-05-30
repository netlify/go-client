# BitBallon API Client in Go

See [the bitballoon package on godoc](http://godoc.org/github.com/BitBalloon/bitballoon-go) for full library documentation.

## Quick Start

First `go get github.com/BitBalloon/bitballoon-go` then use in your go project.

```go
import "github.com/BitBalloon/bitballoon-go"

client := bitballoon.NewClient(&bitballoon.Config{AccessToken: AccessToken})

// Create a new site
site, resp, err := client.Sites.Create(&SiteAttributes{
  Name: "site-subdomain",
  CustomDomain: "www.example.com",
  Password: "secret",
  NotificationEmail: "me@example.com",
})

// Deploy a directory
deploy, resp, err := site.Deploys.Create("/path/to/directory")

// Wait for the deploy to process
err := deploy.WaitForReady(0)

// Get a single site
site, resp, err := client.Sites.Get("my-site-id")

// Set the domain of the site
site.CustomDomain = "www.example.com"

// Update the site
resp, err := site.Update()

// Deploy a new version of the site from a zip file
deploy, resp, err := site.Deploys.Create("/path/to/file.zip")
deploy.WaitForReady(0)

// Delete the site
resp, err := site.Destroy()
```
