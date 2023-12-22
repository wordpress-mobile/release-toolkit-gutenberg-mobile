# render

The `render` command is responsible for generating the checklists that make up the release process.

## Usage
For further CLI usage information on render, run `gbm-cli render -h`.

Currently, `render` supports 2 subcommands: `checklist` and `aztec`:

## Flags

**`--c`** - Optional: if set any subcommand will send the output to the system clipboard. Otherwise the result is sent to stdout.

### `checklist`

#### Flags

**`-version -v`** - Required: This is the release version. If passed a non-zero patch value, the checklist will generate a patch specific checklist. Otherwise the result will be a scheduled checklist. The version must be a full semantic version of the format `X.Y.Z` or `vX.Y.Z`. Short versions which do not include the patch value are not allowed.

**`-message|-m `** - Optional: This is an optional message that will be displayed in a light blue block at the top of the checklist. This is primarily used to explain the reason for a patch release but can also be used for scheduled releases.

**`-date|-d`** - Optional: The `date` option only applies to scheduled releases. It is just an optional string value used to display the release date. If not provided the checklist generator will use the next Thursday from when the script runs.

**`-host-version|-V`** Optional: This sets the host app version. It defaults to "X.XX". Currently it is only used to generate the suggested notification message to be shared in the apps infrastructure slack channel.

**`--a`** - Optional: If set, the checklist generator will check to see if the Aztec versions are correct and will omit the steps to update Aztec. see `aztec` sub command to render the aztec steps in the event the aztec versions are not valid when cutting the release


#### Usage
To generate the HTML output for a release checklist, run `checklist` as a subcommand and pass a version number with `-v`:

```
gbm-cli render checklist -v 1.106.0
```

The command output can be copied to the clipboard with `--c`:

```
gbm-cli render checklist -v 1.106.0 --c
```


### `aztec`

In most release scenarios, Aztec is not updated. The result is that the "Update Aztec" steps are considered conditional and usually left unchecked. By default, the `checklist` command keeps the conditional steps. However, there is a flag that changes this behavior: adding `--a` to the `checklist` command will remove the conditional Aztec section. It also makes the script reach out to the relevant Aztec configs to see if a release version is required. In this case, the checklist command will add the Aztec steps to the checklist but right after the "Before the Release" section, and also add a warning message about updating Aztec.