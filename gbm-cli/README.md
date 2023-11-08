# GBM CLI

## Overview
The GBM CLI tool helps developmental tasks for managing Gutenberg Mobile releases.

The current features include:
- Command to generate the release checklist
- Commands to wrangle Gutenberg Mobile releases

## Installing
Check the latest release in this repository for the binary builds. Currently we only build for MacOS arm64 (apple silicon). The script has only been tested on apple silicon but build of other platforms should work as expected. See the official [go build](https://go.dev/ref/mod#go-install) documentation for alternative builds.

If using apple silicon, download the `gbm-cli` binary from the [latest release(https://github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/releases)
Place the executable in your PATH and reload your shell. Try

```
$ gbm-cli --version
```

To verify installation.

If `go` (above version `1.21`) is installed you can also use:

```
go install github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli@latest
```

This will build and install the executable in the `GOPATH` on your machine.
Note: Verify that `GOPATH` is set before using this method. If it's not set run `export GOPATH=$HOME/go` before calling `go install`

## Authentication

The tool uses the same Github authentication as [`gh`](https://cli.github.com/). If `gh` is installed and authorized there is no need to do anything else.

Otherwise follow these steps:

1. Create a [personal access token](https://github.blog/2013-05-16-personal-api-tokens/)
2. Export the token under the environment variable `GH_TOKEN`

## Development Environment
1. Download and install the [Go package](https://go.dev/doc/install). Check `go.mod` for the current version of go required (Note: anything below `v1.21` will not work)
2. While not required, it is highly recommended to develop with [VSCode](https://code.visualstudio.com/) and install the [Go VSCode](https://marketplace.visualstudio.com/items?itemName=golang.go) extension.

## Releasing

Check the [Releasing](../Releasing.md) doc for more information on creating Gutenberg Mobile releases. Use the following for creating new releases of the CLI tool

When ready to push updates to a new `gbm-cli` version make sure to:
- Increment the version in `./cmd/root.go`
- Merge the PR with the version bump
- Create a new Github Release with the updated version.
- Locally checkout the release tag
- Create a `./bin` directory if you don't have one already
- Run `go build -o ./bin/gbm-cli`
- Add `./bin/gbm-cli` as an artifact to the Github release.

## Testing
For detailed instructions on testing and configuring your development environment, visit [Testing.md](https://github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/blob/cli/update-checklist/cli/Testing.md).


## Structure
The project is setup wth the following directories:


### cmd
The `cmd` directory defines the various cli commands that make up the CLI took. Under the hood `gbm` uses [go-cobra](https://github.com/spf13/cobra/tree/main).

### pkg
The packages in `pkg` are intended as the "public" interface for the tool. The are primarily used by the `cmd` packages but could be used by other go projects.

### templates
The template files in this directory are embedded into the go binary. The can be accessed by using `templates` as the root path segment.
For example using the render package:

```go
 // this can be called anywhere in the project
  checklist := render.Render("template/checklist/checklist.html", data, funcs)
 ```
