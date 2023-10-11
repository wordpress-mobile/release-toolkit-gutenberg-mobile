# GBM CLI

## Overview
The GBM cli tool helps developmental tasks for managing Gutenberg Mobile releases.

The current features include:
- Command to generate the release checklist
- Commands to wrangle Gutenberg Mobile releases

## Prerequisites
Download and install the [Go package](https://go.dev/doc/install). 


## Structure
The project is setup wth the following directories:

### ./bin
The `bin` directory is where the executable files generated by the Go build tools are placed.

### ./cmd
The `cmd` directory defines the various cli commands that make up the CLI took. Under the hood `gbm` uses [go-cobra](https://github.com/spf13/cobra/tree/main). 

### ./pkg
The packages in `pkg` are intended as the "public" interface for the tool. The are primarily used by the `cmd` packages but could be used by other go projects.

### ./templates
The template files in this directory are embedded into the go binary. The can be accessed by using `templates` as the root path segment.
For example using the render package:

```go
 // this can be called anywhere in the project
  checklist := render.Render("template/checklist/checklist.html", data, funcs)
 ```
