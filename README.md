# BitBalloon CLI

Fast command line tool for BitBalloon.

## Commands

Assuming you've installed the binary as `bitballoon`:

### bitballoon create

Creates a new site and returns the ID/URL

```bash
$ bitballoon create --token <access-token>
```

### bitballoon deploy

Deploys a folder or zip file to a BitBalloon site

```bash
$ bitballoon deploy /path/to/site --token <access-token> --site <site-id>
```

### bitballoon update

Updates name, domain, password or notification email of a site

```bash
$ bitballoon update --token <access-token> --site <site-id> --domain www.example.com
```
