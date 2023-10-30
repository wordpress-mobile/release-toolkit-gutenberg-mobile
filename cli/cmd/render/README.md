# render

The `render` command is responsible for generating the checklists that make up the release process.

### Usage
For further CLI usage information on render, run `go run main.go render -h` from the `cli` directory.

Currently, `render` supports 2 subcommands: `checklist` and `aztec`:

#### `checklist`
To generate the HTML output for a release checklist, run `checklist` as a subcommand and pass a version number with `-v`:

```
go run main.go render checklist -v 1.106.0
```

The command output can be copied to the clipboard with `--c`:

```
go run main.go render checklist -v 1.106.0 --c
```

#### `aztec`
In most release scenarios, Aztec is not updated. The result is that the "Update Aztec" steps are considered conditional and usually left unchecked. By default, the `checklist`` command keeps the conditional steps. However, there is a flag that changes this behavior: adding `--a` to the `checklist` command will remove the conditional Aztec section. It also makes the script reach out to the relevant Aztec configs to see if a release version is required. In this case, the checklist command will add the Aztec steps to the checklist but right after the "Before the Release" section, and also add a warning message about updating Aztec.