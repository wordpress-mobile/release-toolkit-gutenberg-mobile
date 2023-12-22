# release

The `release` command is a parent command for running the release process. `release` consists of three subcommands, which represent the three phases of the release flow:

### prepare
Used to prepare Gutenberg and Gutenberg Mobile PRs for the release. Contains three subcommands:

- `all`: Prepare both Gutenberg and Gutenberg Mobile PRs release
- `gb`: Prepare Gutenberg PR for a mobile release
- `gbm`: Prepare Gutenberg Mobile PR release

**Usage**

Prepare a release for both platforms:

```
gbm-cli release prepare all v1.107.0
```

Prepare a release for Gutenberg only:

```
gbm-cli release prepare gb v1.107.0
```

Prepare a release for Gutenberg Mobile only:

```
gbm-cli release prepare gbm v1.107.0
```


**Flags:**
- `--k`, `--keep`: Keep temporary directory after running command
- `--no-tag`:  Prevent tagging the release
- `-h`, `--help`: Command line help for `prepare`


### integrate
Used to integrate a release into the main apps WordPress-iOS and WordPress-Android. If the Android or iOS flags are set, only that platform will be integrated. Otherwise, both will be integrated.

**Usage**

After the `prepare` command has been run and the CI has finished, the main apps integration PRs can be created:

```
gbm-cli release integrate v1.107.0
```

**Flags**
- `-V`, `--host-version`: Target host version, required if creating a patch release
- `-a`, `--android`: Only integrate Android
- `-i`, `--ios`: Only integrate iOS
- `-h`, `--help`: Command line help for `integrate` command

### status

Command used to check the status of any given release:

**Usage**

```
gbm-cli release status 1.07.0
```