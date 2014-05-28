/*
Package bitballoon provides a client for using the BitBalloon API.

To work with the BitBalloon API, start by instantiating a client:

    client := bitballoon.NewClient(&bitballoon.Config{AccessToken: AccessToken})

    // List sites
    sites, resp, err := client.Sites.List(&bitballoon.ListOptions{Page: 1})

    // Create a new site from a Dir
    site, resp, err := client.Sites.Create(&Site{
      Dir: "/path/to/a/site/folder",
    })

    // Wait for the site to process
    err := site.WaitForReady(0)

    // Get a single site
    site, resp, err := client.Sites.Get("my-site-id")

    // Set a Dir to deploy
    site.Dir = "/path/to/site"

    // Deploy the directory to the Site
    resp, err := site.Update()

    // Delete the site
    resp, err := site.Destroy()

*/
package bitballoon
