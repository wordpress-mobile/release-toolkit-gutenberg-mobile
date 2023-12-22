# `gbn-cli`
## Usage
Use `gbm-cli -h` to see the full list of available commands.

## Flags
**`-v, -version`** Displays the current version of the tool

**`-h, --help`** Displays the help menu

### Commands

### `completion`
This is automatically added to the command via the cli library used. It will generate a shell auto completion script. Follow [this guide](https://blog.chmouel.com/posts/cobra-completions/#installation) for how to add the auto completion to your shell.

**Note**: The tool does not implement any custom auto completion and the default generated output has not been tested.

### `render`

#### Usage
The `render` command is responsible for generating the checklists that make up the release process.

#### Flags
**`--c`** - Optional: if set any subcommand will send the output to the system clipboard. Otherwise the result is sent to stdout.

#### Subcommands

## `render checklist`

### Usage
To generate the HTML output for a release checklist, run `checklist` as a subcommand and pass a version number with `-v`:

```
gbm-cli render checklist -v 1.106.0
```

The command output can be copied to the clipboard with `--c`:

```
gbm-cli render checklist -v 1.106.0 --c
```

### Flags
**`-version -v=string`** - Required: This is the release version. If passed a non-zero patch value, the checklist will generate a patch specific checklist. Otherwise the result will be a scheduled checklist. The version must be a full semantic version of the format `X.Y.Z` or `vX.Y.Z`. Short versions which do not include the patch value are not allowed.

**`-message|-m=string`** - Optional: This is an optional message that will be displayed in a light blue block at the top of the checklist. This is primarily used to explain the reason for a patch release but can also be used for scheduled releases.

**`-date|-d=string`** - Optional: The `date` option only applies to scheduled releases. It is just an optional string value used to display the release date. If not provided the checklist generator will use the next Thursday from when the script runs

**`-host-version|-V=string`** Optional: This sets the host app version. It defaults to "X.XX". Currently it is only used to generate the suggested notification message to be shared in the apps infrastructure slack channel.

**`--a`** - Optional: If set, the checklist generator will check to see if the Aztec versions are correct and will omit the steps to update Aztec. see `aztec` sub command to render the aztec steps in the event the aztec versions are not valid when cutting the release

---

### `release`

#### Usage
The `release` command is used to create and review releases

#### Flags
**`--keep`** Optional: Most release commands use a temporary directory which is cleaned up a the end of the command. `--keep` prevents the deletion of the temporary directory. This is useful for development.

#### Subcommands

#### `release prepare gb {version}`

##### Usage
Prepares a release pr on Gutenberg for the version provided. It can be used to create scheduled release or patch releases. A list of pr numbers or commit shas should be provided when creating a patch release.

Examples:

Creating a scheduled release
```
gbm-cli release prepare gb v1.0.0
```

Creating a patch release
```
gbm-cli release prepare gb v1.0.1 --prs=123,124
```

##### Flags
**`--no-tag`** Optional: Prevents adding the tag to Gutenberg. If provided it will also skip the prompt to add the tag.

**`--prs=string,string...`** Optional: Comma separated list of PR numbers. Only used for patch releases to cherry pick the merge commit associated with the pull request.

**`--shas=string,string...`** Optional: Comma separated list of commits. Only used for patch releases to cherry pick commits.


#### `release prepare gbm {version}`

##### Usage
Prepares a release pr on Gutenberg Mobile for the version provided. It can be used to create scheduled release or patch releases.
Examples:

Creating a scheduled release
```
gbm-cli release prepare gbm v1.0.0
```

Creating a patch release
```
gbm-cli release prepare gbm v1.0.1
```

##### Flags


#### `release prepare all {version}`

##### Usage
Prepares both Gutenberg and Gutenberg Mobile release prs.


#### `release integrate {version}`

##### Usage
Create release integration PRs for the provided version. A host version is required if creating a patch version. A release pr will be created on both platforms unless specified by command flags.

#### Flags
**`--android|-a`** Optional: Specifies that an Android release pr is created. If the `--ios` flag is not set then only the Android pr will be created.

**`--ios|-i`** Optional: Specifies that an iOS release pr is created. If the `--android` flag is not set then only the iOS pr will be created.

**`--host-version|-V`** Required for patch releases: The host app version that the release is targeting.


#### `release status {version}`

##### Usage
Outputs the status of the release.

##### Flags
**`watch`** Optional: Periodically refreshes the status output

**`--time|-t`** Optional: Delay in seconds between refreshes (default 10)
