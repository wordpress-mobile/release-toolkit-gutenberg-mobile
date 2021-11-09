# Making a release

The `bundle` directory contains the production version of the project's Javascript. This is what the WordPress apps use to avoid having to build Gutenberg.

You can rebuild those files at any time by running

```
npm run bundle
```

This is useful in case you want to use an unreleased version of the bundle in the apps. For instance, on a PR that's a work in progress, you might want to include to a specific gutenberg-mobile branch in the apps with an updated bundle so reviewers can see the latest changes before approving them (and releasing a new version).

# Release Checklist Script

Use the './release-checklist.sh' command to generate a release checklist.
The checklist will be published as an issue on the [Gutenberg Mobile GitHub repository](https://github.com/wordpress-mobile/gutenberg-mobile/issues).

The script will generate a checklist for each of the [release types](#different-types-of-releases).

With out any arguemnts, the script will prompt for all the checklist options. You can use command line options to skip the prompts. To see the full list of options, run

```
./release-checklist.sh -h
```

The script can also be used to update an existing checklist to add:

1. Aztec update steps
2. Incoming changes

Use the `-i` option with the id of the issue to use the update feature. Then include either the `-a` or `-u` options to add or update the checklist with Aztec or incoming changes steps, respectively.

# Release Automation Script

The `./release-automation.sh` script is used to automate the release process.
It will create 4 branches/PRs:

1. `release/X.XX.X` on [mobile-gutenberg](https://github.com/wordpress-mobile/gutenberg-mobile)
2. `rnmobile/release_X.XX.X` on [wordpress/gutenberg](https://github.com/WordPress/gutenberg)
3. `gutenberg/integrate_release_X.X.X` on [wordpress-mobile/WordPress-Android](https://github.com/wordpress-mobile/WordPress-Android
4. `gutenberg/integrate_release_X.X.X` on [wordpress-mobile/WordPress-iOS](https://github.com/wordpress-mobile/WordPress-iOS)

The latter two PRs can be used to create installable builds for testing.

**Note** The WPAndroid PR requires the `Build Android RN Bridge & Publish to S3` step to be completed before the PR can be confirmed.

# Different types of releases

## Best practices

It's best practice to use the automation script (mentioned in the release template above) for all releases types (alpha, regular, betafix, hotfix). When wrangling a betafix or hotfix, it's important to merge the fix to Gutenberg `trunk` independently of the release process. When the release is cut (by the automation script) the commit(s) that make up the betafix or hotfix should then be cherry-picked onto the Gutenberg release branch.

## 1. Alpha

### When

Whenever a build is needed for testing (usually a few days prior to a Regular release)

### Branches

For example, when the next release will be `1.11.0`.

| Repo             | Cut From | Branch Name                               |
| ---------------- | -------- | ----------------------------------------- |
| gutenberg        | trunk    | rnmobile/release_1.11.0-alpha1            |
| gutenberg-mobile | develop  | release/1.11.0-alpha1                     |
| WPAndroid        | develop  | gutenberg/integrate_release_1.11.0-alpha1 |
| WPiOS            | develop  | gutenberg/integrate_release_1.11.0-alpha1 |

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
| gutenberg-mobile | develop  | release/1.11.0                     | trunk & develop |
| WPAndroid        | develop  | gutenberg/integrate_release_1.11.0 | develop         |
| WPiOS            | develop  | gutenberg/integrate_release_1.11.0 | develop         |

## 3. Betafix

### When

A fix is targeting a main app version that is not yet released (meaning the release branch is cut but it's still in beta) and a new gutenberg-mobile release is needed.

### Branches

For example when releasing gutenberg-mobile `1.11.1` while main apps version `22.2.0` is in beta which currently has gutenberg-mobile `1.11.0` in it.
At the same time there could also be a regular release going on for example for gutenberg-mobile version `1.12.0`.

| Repo             | Cut From                | Branch Name                        | Merging To                                                       |
| ---------------- | ----------------------- | ---------------------------------- | ---------------------------------------------------------------- |
| gutenberg        | rnmobile/release_1.11.0 | rnmobile/release_1.11.1            | trunk & (maybe also) rnmobile/release_1.12.0                     |
| gutenberg-mobile | release/1.11.0          | release/1.11.1                     | trunk & develop & (maybe also) release/1.12.0                    |
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
| gutenberg        | rnmobile/release_1.11.0 | rnmobile/release_1.11.1            | trunk & (maybe also) rnmobile/release_1.12.1                     |
| gutenberg-mobile | release/1.11.0          | release/1.11.1                     | trunk & develop & (maybe also) release/1.12.1                    |
| WPAndroid        | release/22.2.0          | gutenberg/integrate_release_1.11.1 | release/22.2.1 & (maybe also) gutenberg/integrate_release_1.12.1 |
| WPiOS            | release/22.2.0          | gutenberg/integrate_release_1.11.1 | release/22.2.1 & (maybe also) gutenberg/integrate_release_1.12.1 |

### Automation script differences

1. If necessary create new patch version branches `release/X.Y.1` in WPiOS and WPAndroid.

Rest should be same the as Betafix

### Release checklist template differences

1. Include `Hotfix` in the heading
1. After the fix is merged and if there is an ongoing regular release, betafix or hotfix then the changes should be brought back to those branches as well.

Rest should be same the as Betafix

# Handling Aztec Updates

Before creating the release, make sure Aztec is up to date. Verify `[gutenberg-mobile/RNTAztecView.podspec](#)` and `gutenberg-mobile/gutenberg/packages/react-native-aztec/RNTAztecView.podspec` refer to the same `Wordpress-Aztex-iOS` version and are poiting to a stable tagged release (e.g. 1.14.1). If not, we may need to create a new Aztec release.

The `./release_automation.sh` will verify if the Aztec version is correct. If not run `./release_checklist.sh -a -i {release issue id}` to update the release checklist.

### Updating Aztec iOS

Follow [this process](https://github.com/wordpress-mobile/AztecEditor-iOS/blob/develop/Documentation/ReleaseProcess.md) for iOS. Afterwards update Aztec version references within `gutenberg-mobile/RNTAztecView.podspec` and `gutenberg-mobile/gutenberg/packages/react-native-aztec/RNTAztecView.podspec` to the new WordPress-Aztec-iOS version.

### Updating Aztec Android

For Android, release are created via releases. Go to [AztecEditor-Android/releases](https://github.com/wordpress-mobile/AztecEditor-Android/releases) and draft a new release with the new version tag. Use the version for the release title and add the changelog as the descriptions. The binaries are created when the release is published.

# Handling Incoming Changes

If addiotional changes (e.g. bug fixes) were merged into `gutenberg-mobile/release/X.XX.X` or `gutenberg/rnmobile/release-X.XX.X` branches add an incoming changes checklist to the release issue. The addtion can be added by running:

```
 `./release_checklist.sh -u -i {release issue id}` -m "Some info about the incoming change".
```

# Verify Localization Strings

Bundling will generate two location strings files:

- [bundle/android/strings.xml](https://href.li/?https://github.com/wordpress-mobile/gutenberg-mobile/blob/develop/bundle/android/strings.xml)
- [bundle/ios/GutenbergNativeTranslations.swift](https://href.li/?https://github.com/wordpress-mobile/gutenberg-mobile/blob/develop/bundle/ios/GutenbergNativeTranslations.swift)

These files will only be modified if there are any string changes. Any modfidications to these files should be verified for correctness:

- Check that extra strings from non-native files are not being added.
- Confirm that any strings referenced in the code are not being removed.

More info can be found in this [issue](https://href.li/?https://github.com/wordpress-mobile/gutenberg-mobile/issues/3466).

**If any issue is found, it will require manually modifying the files and pushing the changes to the release branch.**
