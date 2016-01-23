/*
Package netlify provides a client for using the netlify API.

To work with the netlify API, start by instantiating a client:

    client := netlify.NewClient(&netlify.Config{AccessToken: AccessToken})

    // List sites
    sites, resp, err := client.Sites.List(&netlify.ListOptions{Page: 1})

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

    // Configure Continuous Deployment for a site

    // First get a deploy key
    deployKey, resp, err := site.DeployKeys.Create()
    // Then make sure the public key (deployKey.PublicKey)
    // has access to the repository

    // Configure the repo
    resp, err = site.ContinuousDeployment(&nefliy.RepoOptions{
      Repo: "netlify/netlify-home",
      Provider: "github",
      Dir: "_site",
      Cmd: "gulp build",
      Branch: "master",
      DeployKeyId: deployKey.Id
    })
    if err != nil {
      // Now make sure to add this URL as a POST webhook to your
      // repository:
      site.DeployHook
    }


    // Deleting a site
    resp, err := site.Destroy()
*/
package netlify
