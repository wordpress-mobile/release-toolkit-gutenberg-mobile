# Gutenberg Mobile CLI
The Gutenberg Mobile CLI (`gbm-cli`) tool is available from this repo. The tool is recommended for handling Gutenberg Mobile releases. See [Installing the CLI tool](#installing.md) for more information.

# Release Checklist Template

Use the `gbm-cli` tool to generate the release checklist:

```
gbm-cli render checklist -v {version} --c
```

The `--c` flag should send the rendered checklist output to your system clipboard. If not try manually capturing the output without the `--c` flag.

# Using gbm-cli for releasing

 To create the Gutenberg and Gutenberg Mobile PRs run:

 ```
 $ gbm-cli release prepare all {version}
 ```

The Android PR can be created right away but the [Build iOS RN XCFramework & Publish to S3](https://github.com/wordpress-mobile/gutenberg-mobile/blob/trunk/.buildkite/pipeline.yml#L167) must complete before creating the iOS PR.

The status for both builds can be checked by running:

```
$ gbm-cli release status {version}
```

Once the platform builds are ready run:

```
$ gbm-cli release integrate {gutenberg-mobile-version}
```

## Patch releases

The cli tool

# Different types of releases

## 1. Alpha

Note: The `gbm-cli` tool does not currently support alpha releases (see #217). Use the legacy `release_automation.sh` script if needed.

### When

Whenever a build is needed for testing (usually a few days prior to a Regular release)

### Branches

For example, when the next release will be `1.11.0`.

| Repo             | Cut From | Branch Name                               |
| ---------------- | -------- | ----------------------------------------- |
| gutenberg        | trunk    | rnmobile/release_1.11.0-alpha1            |
| gutenberg-mobile | trunk    | release/1.11.0-alpha1                     |
| WPAndroid        | trunk    | gutenberg/integrate_release_1.11.0-alpha1 |
| WPiOS            | trunk    | gutenberg/integrate_release_1.11.0-alpha1 |

### Automation script differences

Compared to a Regular release, the differences here are:

- When the script asks for the new version number, don't forget to add the `-alpha` suffix (e.g. `1.11.0-alpha1`).
- All PRs created by the release script should be edited to clarify that they are temporary and will be deleted when testing is finished.

### Release checklist template differences

The release checklist is not used for alpha releases. When testing is finished, please close all PRs and delete all branches created by the script.

## 2. Regular

### When

On Thursdays, the week before the main apps (WPiOS & WPAndroid) have cut their releases, every 2 weeks.

### Branches

For example when releasing gutenberg-mobile `1.11.0`.

| Repo             | Cut From | Branch Name                        | Merging To      |
| ---------------- | -------- | ---------------------------------- | --------------- |
| gutenberg        | trunk    | rnmobile/release_1.11.0            | trunk           |
| gutenberg-mobile | trunk    | release/1.11.0                     | trunk           |
| WPAndroid        | trunk    | gutenberg/integrate_release_1.11.0 | trunk           |
| WPiOS            | trunk    | gutenberg/integrate_release_1.11.0 | trunk           |

## 3. Betafix

### When

A fix is targeting a main app version that is not yet released (meaning the release branch is cut but it's still in beta) and a new gutenberg-mobile release is needed.

### Branches

For example when releasing gutenberg-mobile `1.11.1` while main apps version `22.2.0` is in beta which currently has gutenberg-mobile `1.11.0` in it.
At the same time there could also be a regular release going on for example for gutenberg-mobile version `1.12.0`.

| Repo             | Cut From                | Branch Name                        | Merging To                                                       |
| ---------------- | ----------------------- | ---------------------------------- | ---------------------------------------------------------------- |
| gutenberg        | rnmobile/release_1.11.0 (tag) | rnmobile/release_1.11.1            | trunk & (maybe also) rnmobile/release_1.12.0                     |
| gutenberg-mobile | v1.11.0 (tag)          | release/1.11.1                     | trunk & (maybe also) release/1.12.0                              |
| WPAndroid        | release/22.2.0          | gutenberg/integrate_release_1.11.1 | release/22.2.0 & (maybe also) gutenberg/integrate_release_1.12.0 |
| WPiOS            | release/22.2.0          | gutenberg/integrate_release_1.11.1 | release/22.2.0 & (maybe also) gutenberg/integrate_release_1.12.0 |

### Automation script differences

1. Before running the script switch to the relevant branch to cut from in gutenberg-mobile repo.
1. Run [release_automation.sh](./release_automation.sh) as usual.
1. When asked by the script enter the relevant branch names to cut from (to target) in other repos.
1. If a commit that is fixing the issue is already merged to gutenberg, when asked by the script enter the commit hash to be cherry-picked.

### Release checklist template differences

1. Include `Betafix` in the heading.
1. `after_X.XX.X` branches can be ignored.

## 4. Hotfix

### When

A fix is targeting a main app version that is already released and a new gutenberg-mobile release is needed.

### Branches

For example when releasing gutenberg-mobile `1.11.1` while main apps version `22.2.0` is released which currently has gutenberg-mobile `1.11.0` in it.
At the same time there could also be a regular release, a betafix or even another hotfix going on for example for gutenberg-mobile version `1.12.1`.

| Repo             | Cut From                | Branch Name                        | Merging To                                                       |
| ---------------- | ----------------------- | ---------------------------------- | ---------------------------------------------------------------- |
| gutenberg        | rnmobile/release_1.11.0 (tag) | rnmobile/release_1.11.1            | trunk & (maybe also) rnmobile/release_1.12.1                     |
| gutenberg-mobile | v1.11.0 (tag)          | release/1.11.1                     | trunk & (maybe also) release/1.12.1                              |
| WPAndroid        | release/22.2.0          | gutenberg/integrate_release_1.11.1 | release/22.2.1 & (maybe also) gutenberg/integrate_release_1.12.1 |
| WPiOS            | release/22.2.0          | gutenberg/integrate_release_1.11.1 | release/22.2.1 & (maybe also) gutenberg/integrate_release_1.12.1 |

### Automation script differences

1. If necessary create new patch version branches `release/X.Y.1` in WPiOS and WPAndroid.

Rest should be same the as Betafix

### Release checklist template differences

1. Include `Hotfix` in the heading
1. After the fix is merged and if there is an ongoing regular release, betafix or hotfix then the changes should be brought back to those branches as well.

Rest should be same the as Betafix
