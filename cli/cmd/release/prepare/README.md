# prepare

The `prepare` command is responsible for generating the checklists that make up the release process.

Used to prepare Gutenberg and Gutenberg Mobile PRs for the release. Contains three subcommands:

- `all`: Prepare both Gutenberg and Gutenberg Mobile PRs release
- `gb`: Prepare Gutenberg PR for a mobile release
- `gbm`: Prepare Gutenberg Mobile PR release

### Usage

Prepare a release for both platforms:

```
go run main.go release prepare all v1.107.0
```

Prepare a release for Gutenberg only:

```
go run main.go release prepare gb v1.107.0
```

Prepare a release for Gutenberg Mobile only:

```
go run main.go release prepare gbm v1.107.0
```
