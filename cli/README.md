# GBM CLI

## Overview
The GBM cli tool helps developmental tasks for the Gutenberg Mobile project

The current features include:
- Command to generate the release checklist
- Commands to wrangle Gutenberg Mobile releases


## Structure
The project is setup wth the following directories

### ./cmd
This defines the various cli commands. Under the hood `gbm` uses go-cobra. The inner structure is

```
cmd/
  root.go
  command1/
    root.go (of command1)
    subcommandA.go
    subcommandB.go
    ...

  command2/
    root.go (of command2)
    subcommandA.go
    subcommandB.go
```

Run the help command to see the currently supported commands:

`$ gbm -h`


### ./internal

These are lower level wrappers that provide helpers to the main command flows.

#### exc

The internal `exc` package provides helpers for some common shell commands. Specifically it provides wrappers
for `npm`, `bundler` and `git` (See note below)

#### repo

The internal `repo` package provides helpers to interact wit git repos. It provides functions for local git interactions
and calls to the GitHub Apis. The project uses `go-git` when ever possible. The package drops down to `git` (via the `exc` package ) when `go-git` does not support a git operation.

#### utils
`utils` provides common general purpose utils like logging


### ./pkg

The packages in `pkg` are intended as the "public" interface for the tool. The are primarily used by the `cmd` packages but could be used by other go projects.

#### gbm

`gbm` manages companion prs for any of the Gutenberg Mobile submodules. It can create and update PRs. Currently it only supports updates from Gutenberg.

#### integration

`integration` manages integration PRs for the main apps. It can create and update PRs.

#### release

The `release` package provides functions to create GBM releases. The various release "domains" are reflected in the specific files:

- `gb.go` - Sets up the Gutenberg release PR.
- `gbm.go` - Sets up and updates GBM release PRs.
- `integration.go` - Sets up and updates integration release PRs.
- `publish.go` - provides functions to publish a release on GBM.
- `utils.go` - utilities specific to release tasks.

### ./render

The `render` function provides a light wrapper to render templates in the `./templates` directory


### ./templates

The template files in this directory are embedded into the go binary. The can be accessed by using `templates` as the root path segment.
For example using the render package:

```go
 // this can be called anywhere in the project
  checklist := render.Render("template/checklist/checklist.html", data, funcs)
 ```


### E2E testing

The project recognizes the following env variables

`GBM_WORDPRESS_ORG`
`GBM_WPMOBILE_ORG`
`GBM_AUTOMATTIC_ORG`

Set these to a non production org on Github to perform e2e tests. The repos the should have the same name as the main repos, for example:

`jhnstn/gutenberg`
`jhnstn/gutenberg-mobile`
`jhnstn/WordPress-iOS`
`jhnstn/WordPress-Android`

Note: when testing the release flow make sure the `.gitmodules` config in the test `gutenberg-mobile` repo is pointing to the test org for `gutenberg`
The script will honor the `env` setting to create PRs but currently does not override the `git remote` setting of the submodules.
