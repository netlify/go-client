package tools

//go:generate gobin -m -run github.com/go-swagger/go-swagger/cmd/swagger generate client -A netlify -f vendor/github.com/netlify/open-api/swagger.yml -t . -c plumbing --with-flatten=full
