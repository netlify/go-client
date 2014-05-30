/*
Package bitballoon provides a client for using the BitBalloon API.

To work with the BitBalloon API, start by instantiating a client:

    client := bitballoon.NewClient(&bitballoon.Config{AccessToken: AccessToken})

    // List sites
    sites, resp, err := client.Sites.List(&bitballoon.ListOptions{Page: 1})

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
*/
package bitballoon
