# CONTRIBUTING

The go-client is an [netlify/open-api][open-api] derived http client generated using [go-swagger][go-swagger].  Starting with version [`2.0.0`](https://github.com/netlify/go-client/releases/tag/v2.0.0) it is managed with [Go 1.11 modules][go-modules], and all external tools used for generation are managed with [gobin][gobin] + Go modules.  The [`swagger.yml`] is consumed as a vendored build-time asset via a [git submodule](https://git-scm.com/book/en/v2/Git-Tools-Submodules).  See [GMBE:Tools as dependencies](https://github.com/go-modules-by-example/index/tree/master/010_tools) and [GMBE:Using `gobin` to install/run tools](https://github.com/go-modules-by-example/index/tree/master/017_using_gobin) for a deeper explanation.

### Advantages:

- Everyone will use the same version of the `swagger` CLI.
- Simplify the direct usage of `go tools` (code generation controlled by [`go generate`](https://blog.golang.org/generate))
- Simplify the internals of the included Makefile
- Decouple the development of the go client from the swagger spec.

### Requirements

- [Go 1.11 or higher](https://blog.golang.org/go1.11)
- Ensure your `$GOPATH/bin` is added to your `$PATH` (`export PATH=$GOPATH/bin:$PATH`).
- The only globally installed tool required is [gobin][gobin].  It will be installed/updated when you run `Make deps`.

### Start development

```console
# outside of your GOPATH or in the GOPATH with GO111MODULE=on (see )
$ git clone git@github.com:netlify/go-client.git
$ cd go-client
$ make deps
$ make all
$ make test
```

Or with the go tools:

```console
$ go get -u github.com/myitcv/gobin # if you haven't already
$ git clone git@github.com:netlify/go-client.git
$ cd go-client
$ go mod download
$ go generate
$ go build ./...
$ go test ./...
```

### Updating the swagger spec

To update the swagger spec, update the tag that the open-api submodule points at:

```console
$ cd vendor/github.com/netlify/open-api
$ git checkout v0.10.0
$ cd ../../../../
$ make all
```

Using a tool like [marwan-at-work/mod](https://github.com/marwan-at-work/mod) to bump the version of the package if you have a major version change.

## License

By contributing to Netlify Node Client, you agree that your contributions will be licensed
under its [MIT license](LICENSE).


[godoc-img]: https://godoc.org/github.com/netlify/go-client/?status.svg
[godoc]: https://godoc.org/github.com/netlify/go-client
[goreport-img]: https://goreportcard.com/badge/github.com/netlify/go-client
[goreport]: https://goreportcard.com/report/github.com/netlify/go-client
[git-img]: https://img.shields.io/github/release/netlify/go-client.svg
[git]: https://github.com/netlify/go-client/releases/latest
[gobin]: https://github.com/myitcv/gobin
[modules]: https://github.com/golang/go/wiki/Modules
[open-api]: https://github.com/netlify/open-api
[go-swagger]: https://github.com/go-swagger/go-swagger
[go-modules]: https://github.com/golang/go/wiki/Modules
[swagger]: https://github.com/netlify/open-api/blob/master/swagger.yml
