# release

The `release` command is a parent command for running the release process. `release` consists of three subcommands, which represent the three phases of the release flow:

### prepare
Used to prepare Gutenberg and Gutenberg Mobile PRs for the release. Contains three subcommands:

- `all`: Prepare both Gutenberg and Gutenberg Mobile PRs release
- `gb`: Prepare Gutenberg PR for a mobile release
- `gbm`: Prepare Gutenberg Mobile PR release

**Flags:**
- `--k`, `--keep`: Keep temporary directory after running command
- `--no-tag`:  Prevent tagging the release
- `-h`, `--help`: Command line help for `prepare`

### integrate
Used to integrate a release into the main apps WordPress-iOS and WordPress-Android. If the Android or iOS flags are set, only that platform will be integrated. Otherwise, both will be integrated.

**Usage**
After the `prepare` command has been run and the CI has finished, the main apps integration PRs can be created:

```
go run main.go release integrate v1.107.0
```

**Flags**
- `-a`, `--android`: Only integrate Android
- `-i`, `--ios`: Only integrate iOS
- `-h`, `--help`: Command line help for `integrate` command

### status