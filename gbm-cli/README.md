# GBM CLI

## Overview
The GBM CLI tool helps developmental tasks for managing Gutenberg Mobile releases.

The current features include:
- Command to generate the release checklist
- Commands to wrangle Gutenberg Mobile releases

## Installing
Check the [latest release](https://github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/releases/latest) in this repository for the binary builds.
Currently we only build for MacOS arm64 (apple silicon). The script has only been tested on apple silicon but builds of other platforms should work as expected. See the official [go build](https://go.dev/ref/mod#go-install) documentation for alternative builds.

If using apple silicon, fetch the `gbm-cli` binary from the [latest release](https://github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/releases/latest)
Note: MacOS is very strict about downloading unsigned binaries. There is a work around to allowing them but the easier route is to use `wget`:

```
$ wget https://github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/releases/latest/download/gbm-cli
chmod +x ./gbm-cli
```

Then place the executable in your `$PATH` and reload your shell. To verify installation run:

```
$ gbm-cli --version
```

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
2. Export the token under the environment variable `GH_TOKEN` or `GITHUB_TOKEN`

## Commands

### `gbn-cli`
#### Usage
Use `gbm-cli -h` to see the full list of available commands.

#### Flags

**`-v, -version`** Displays the current version of the tool

**`-h, --help`** Displays the help menu

### `render`

#### Usage
The `render` command is responsible for generating the checklists that make up the release process.

#### Flags

**`--c`** - Optional: if set any subcommand will send the output to the system clipboard. Otherwise the result is sent to stdout.

#### Subcommands

#### `render checklist`

##### Usage
To generate the HTML output for a release checklist, run `checklist` as a subcommand and pass a version number with `-v`:

```
gbm-cli render checklist -v 1.106.0
```

The command output can be copied to the clipboard with `--c`:

```
gbm-cli render checklist -v 1.106.0 --c
```

##### Flags

**`-version -v`** - Required: This is the release version. If passed a non-zero patch value, the checklist will generate a patch specific checklist. Otherwise the result will be a scheduled checklist. The version must be a full semantic version of the format `X.Y.Z` or `vX.Y.Z`. Short versions which do not include the patch value are not allowed.

**`-message|-m `** - Optional: This is an optional message that will be displayed in a light blue block at the top of the checklist. This is primarily used to explain the reason for a patch release but can also be used for scheduled releases.

**`-date|-d`** - Optional: The `date` option only applies to scheduled releases. It is just an optional string value used to display the release date. If not provided the checklist generator will use the next Thursday from when the script runs

**`-host-version|-V`** Optional: This sets the host app version. It defaults to "X.XX". Currently it is only used to generate the suggested notification message to be shared in the apps infrastructure slack channel.

**`--a`** - Optional: If set, the checklist generator will check to see if the Aztec versions are correct and will omit the steps to update Aztec. see `aztec` sub command to render the aztec steps in the event the aztec versions are not valid when cutting the release


### `release`

#### Usage

#### Flags

#### Subcommands




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
