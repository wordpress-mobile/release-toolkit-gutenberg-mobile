# Making a release

The `bundle` directory contains the production version of the project's Javascript. This is what the WordPress apps use to avoid having to build Gutenberg.

You can rebuild those files at any time by running

```
npm run bundle
```

This is useful in case you want to use an unreleased version of the bundle in the apps. For instance, on a PR that's a work in progress, you might want to include to a specific gutenberg-mobile branch in the apps with an updated bundle so reviewers can see the latest changes before approving them (and releasing a new version).

# Release Timeline

Currently, we are experimenting with syncing the Gutenberg Mobile release with the main app release schedule (p9ugOq-1LE-p2). We are also experimenting with performing manual tests every week (p9ugOq-1MA-p2), with a round of "smoke" test before cutting the release and a full round of tests after the release has been integrated into the main apps. A typical Gutenberg Mobile release schedule might look like the following:

- Tuesday of release week: perform a round of "smoke" tests and message all targeted PRs that the release will be cut on Thursday.
- Thursday of release week: start process to cut the release and integrate it into the main apps. (Main apps are cut upcoming Monday)
- Tuesday after release week: perform full round of writing flow and sanity tests.
- Remainder of main app release period: monitor main app release P2 posts for issues found.

# Release Checklist Template

When you are ready to cut a new release, use the following template.

For the post title, use this (replacing `X.XX.X` with the applicable release number):

```
Gutenberg Mobile X.XX.X – Release Scenario
```

For the body of the post, just copy this checklist and again replace all occurrences of `X.XX.X` with the applicable release number.

<details><summary>Click to expand</summary>
<p>
  
```html
<!-- wp:paragraph -->
<p>This checklist is based on the <a href="https://github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/blob/develop/Releasing.md#release-checklist-template">Release Checklist Template</a>. If you need a checklist for a new gutenberg-mobile release, please copy from that template.</p>
<!-- /wp:paragraph -->

<!-- wp:paragraph -->
<p>+mobilegutenberg +mobilegutenpagesp2</p>
<!-- /wp:paragraph -->

<!-- wp:heading {"level":3} -->
<h3>Before the Release (Tuesday and Wednesday)</h3>
<!-- /wp:heading -->

<!-- wp:group -->
<div class="wp-block-group"><!-- wp:paragraph -->
<p>o Visit all open gutenberg-mobile PRs that are assigned to X.XX.X milestone and leave a comment with a message similar to the following by Tuesday: </p>
<!-- /wp:paragraph -->

<!-- wp:quote -->
<blockquote class="wp-block-quote"><p>Hey [author]. We will cut the X.XX.X release on [date]. I plan to circle back and bump this PR to the next milestone then, but please let me know if you'd rather us work to include this PR in X.XX.X. Thanks! </p></blockquote>
<!-- /wp:quote --></div>
<!-- /wp:group -->

<!-- wp:paragraph -->
<p>o Midway through the week of the release on Wednesday, create installable builds for WPiOS and WPAndroid based off the current <code>develop</code> branch and complete the <a href="https://github.com/wordpress-mobile/test-cases/tree/master/test-cases/gutenberg/writing-flow">general writing flow test cases</a>. </p>
<!-- /wp:paragraph -->

<!-- wp:heading {"level":3} -->
<h3>Create the Release (Thursday)</h3>
<!-- /wp:heading -->

<!-- wp:paragraph -->
<p>o Verify that <code>gutenberg-mobile/RNTAztecView.podspec</code> and <code>gutenberg-mobile/gutenberg/packages/react-native-aztec/RNTAztecView.podspec</code> refer to the same <code>WordPress-Aztec-iOS</code> version and are pointing to a stable, tagged release (e.g. 1.14.1). If they are not, we may need to <a href="#create-a-new-aztec-release">create a new Aztec</a> release.</p>
<!-- /wp:paragraph -->

<!-- wp:paragraph -->
<p>o Clone the <a href="https://github.com/wordpress-mobile/release-toolkit-gutenberg-mobile">release scripts</a> or pull the latest version if you have already cloned it.</p>
<!-- /wp:paragraph -->

<!-- wp:paragraph -->
<p>o Review the <a href="https://github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/blob/develop/Releasing.md">release script instructions</a>. In your clone of the release scripts, run the script via:  <code>./release_automation.sh</code>. This creates the gutenberg and gutenberg-mobile release PRs as well as WPAndroid and WPiOS integration PRs.<br><br><strong>Note:</strong> You might want to wait a bit before confirming WPAndroid PR creation so gutenberg-mobile can have enough time to finish the <code>Build Android RN Bridge &amp; Publish to S3</code> job on CI which is needed by WPAndroid CI.</p>
<!-- /wp:paragraph -->

<!-- wp:group -->
<div class="wp-block-group"><!-- wp:paragraph -->
<p>o If this is a scheduled release (e.g. X.XX.0) and not a beta/hot fix (e.g. X.XX.2), post a message similar to the following to the <code>#mobile-gutenberg</code> and <code>#mobile-gutenberg-platform</code> Slack channels: </p>
<!-- /wp:paragraph -->

<!-- wp:quote -->
<blockquote class="wp-block-quote"><p>⚠️ The gutenberg-mobile X.XX.X release branches are now cut. Please do not merge any Gutenberg-related changes into the WPiOS or WPAndroid <code>develop</code> branches until <em>after</em> the main apps cut their own releases next week. If you'd like to merge changes now, merge them into the <code>gutenberg/after_X.XX.X</code> branches. </p></blockquote>
<!-- /wp:quote --></div>
<!-- /wp:group -->

<!-- wp:paragraph -->
<p>o Verify the localization strings files (<a href="https://github.com/wordpress-mobile/gutenberg-mobile/blob/develop/bundle/android/strings.xml">bundle/android/strings.xml</a>, <a href="https://github.com/wordpress-mobile/gutenberg-mobile/blob/develop/bundle/ios/GutenbergNativeTranslations.swift">bundle/ios/GutenbergNativeTranslations.swift</a>) have been generated properly. Check that we're not adding extra strings from non-native files and that we're not removing strings that are referenced in the code (more info can be found in this <a href="https://github.com/wordpress-mobile/gutenberg-mobile/issues/3466">issue</a>). <strong>If any issue is found, it will require manually modifying the files and push them to the release branch.</strong> If no strings are updated, it is expected to not see those files modified.</p>
<!-- /wp:paragraph -->
  
<!-- wp:paragraph -->
<p>o In both <code>RELEASE-NOTES.txt</code> and <code>gutenberg/packages/react-native-editor/CHANGELOG.md</code>, replace <code>Unreleased</code> section with the release version and create a new <code>Unreleased</code> section.</p>
<!-- /wp:paragraph -->

<!-- wp:paragraph -->
<p>o Verify the WPAndroid PR build succeeds. If PR CI tasks include a 403 error related to an inability to resolve the <code>react-native-bridge</code> dependency, you must wait for the <code>Build Android RN Bridge &amp; Publish to S3</code> task to succeed in gutenberg-mobile and then restart the WPAndroid CI tasks.</p>
<!-- /wp:paragraph -->  
  
<!-- wp:paragraph -->
<p>o Run the Optional Tests on both the WPiOS and WPAndroid PRs.</p>
<!-- /wp:paragraph -->

<!-- wp:paragraph -->
<p>o Trigger an installable build on WPiOS PR.</p>
<!-- /wp:paragraph -->

<!-- wp:paragraph -->
<p>o Once the installable builds are ready, perform a quick smoke test of the editor on both iOS and Android to verify it launches without crashing. We will perform additional testing after the main apps cut their releases. </p>
<!-- /wp:paragraph -->

<!-- wp:paragraph -->
<p>o Fill in the missing parts of the gutenberg-mobile PR description. When filling in the "Changes" section, link to the most descriptive GitHub issue for any given change and consider adding a short description. Testers rely on this section to gather more details about changes in a release.</p>
<!-- /wp:paragraph -->

<!-- wp:paragraph -->
<p>o Mark all 4 PRs ready for review and request review from your release wrangler buddy.</p>
<!-- /wp:paragraph -->

<!-- wp:paragraph -->
<p>o If this is a release for inclusion in the frozen WPiOS and WPAndroid release branches (i.e. this is a beta/hot fix, e.g. X.XX.2), ping the directly responsible individual handing the release of each platform of the main apps.</p>
<!-- /wp:paragraph -->

<!-- wp:heading {"level":3} -->
<h3 id="create-a-new-aztec-release">Create an Aztec Release (conditional)</h3>
<!-- /wp:heading -->

<!-- wp:paragraph -->
<p>ℹ️ If <code>gutenberg-mobile/RNTAztecView.podspec</code> and <code>gutenberg-mobile/gutenberg/packages/react-native-aztec/RNTAztecView.podspec</code> refer to a commit SHA instead of a stable release (e.g. 1.14.1) or refer to <em>different</em> versions, the steps in this section may need to be completed. </p>
<!-- /wp:paragraph -->

<!-- wp:paragraph -->
<p>o Verify all Aztec PRs attached to the "Next Release" milestone or PRs with changes required for this Gutenberg release have been merged before next steps.</p>
<!-- /wp:paragraph -->

<!-- wp:paragraph -->
<p>o Open a PR on Aztec repo to update the <code>CHANGELOG.md</code> and <code>README.md</code> files with the new version name.</p>
<!-- /wp:paragraph -->

<!-- wp:paragraph -->
<p>o Create a new release and name it with the tag name from step 1. For Aztec-iOS, follow <a href="https://github.com/wordpress-mobile/AztecEditor-iOS/blob/develop/Documentation/ReleaseProcess.md">this process</a>. For Aztec-Android, releases are created via the <a href="https://github.com/wordpress-mobile/AztecEditor-Android/releases">GitHub releases page</a> by hitting the “Draft new release” button, put the tag name to be created in the tag version field and release title field, and also add the changelog to the release description. The binary assets (.zip, tar.gz files) are attached automatically after hitting “Publish release”.</p>
<!-- /wp:paragraph -->

<!-- wp:paragraph -->
<p>o Update Aztec version references within <code>gutenberg-mobile/RNTAztecView.podspec</code> and <code>gutenberg-mobile/gutenberg/packages/react-native-aztec/RNTAztecView.podspec</code> to the new <code>WordPress-Aztec-iOS</code> version.</p>
<!-- /wp:paragraph -->

<!-- wp:heading {"level":3} -->
<h3>Manage Incoming Changes (conditional)</h3>
<!-- /wp:heading -->

<!-- wp:paragraph -->
<p>ℹ️ If additional changes (e.g. bug fixes) were merged into the gutenberg-mobile <code>release/X.XX.X</code> or in gutenberg <code>rnmobile/release-X.XX.X</code> branches, the steps in this section need to be completed.</p>
<!-- /wp:paragraph -->

<!-- wp:paragraph -->
<p>o After a merge happened in gutenberg-mobile <code>release/X.XX.X</code> or in gutenberg <code>rnmobile/release-X.XX.X</code>, ensure the <code>gutenberg</code> submodule points to the correct hash and the <code>rnmobile/release-X.XX.X</code> in the gutenberg repo branch has been updated.</p>
<!-- /wp:paragraph -->

<!-- wp:paragraph -->
<p>o If there were changes in gutenberg repo, make sure to cherry-pick the changes that landed in the <code>trunk</code> branch back to the release branch and don't forget to run <code>npm run bundle</code> in gutenberg-mobile again if necessary.</p>
<!-- /wp:paragraph -->

<!-- wp:paragraph -->
<p>o Add the new change to the "Extra PRs that Landed After the Release Was Cut" section of the gutenberg-mobile PR description.</p>
<!-- /wp:paragraph -->

<!-- wp:paragraph -->
<p>o Re-run the Optional Tests on both the WPiOS and WPAndroid PRs.</p>
<!-- /wp:paragraph -->

<!-- wp:heading {"level":3} -->
<h3>Integrate the Release (Friday or earlier)</h3>
<!-- /wp:heading -->

<!-- wp:paragraph -->
<p>o Verify the <code>gutenberg</code> ref within the gutenberg-mobile release branch is pointed to the latest commit in the gutenberg release branch.</p>
<!-- /wp:paragraph -->

<!-- wp:paragraph -->
<p>o Create and push a <code>rnmobile/X.XX.X</code> git tag for the head of gutenberg release branch. </p>
<!-- /wp:paragraph -->

<!-- wp:paragraph -->
<p>o Ensure that the bundle files are updated to include any changes to the release branch by running <code>npm run bundle</code> in gutenberg-mobile release branch and committing any changes. </p>
<!-- /wp:paragraph -->

<!-- wp:paragraph -->
<p>o Merge the <strong>gutenberg-mobile</strong> PR to <code>trunk</code>. Use "Create a merge commit" option when merging, otherwise there could be conflicts between <code>trunk</code> and release branch in the next release. WARNING: Do not merge the <strong>gutenberg</strong> PR into <code>trunk</code> at this point.</p>
<!-- /wp:paragraph -->

<!-- wp:paragraph -->
<p>o <a href="https://github.com/wordpress-mobile/gutenberg-mobile/releases/new?tag=vX.XX.X&amp;target=trunk&amp;title=Release%20X.XX.X">Create a new gutenberg-mobile GitHub Release</a>. Include a list of changes in the Release description.</p>
<!-- /wp:paragraph -->

<!-- wp:paragraph -->
<p>o In WPiOS, update the reference to point to the <em>tag</em> of the Release created in the previous task. </p>
<!-- /wp:paragraph -->

<!-- wp:paragraph -->
<p>o In WPAndroid, update the <code>gutenbergMobileVersion</code> in <code>build.gradle</code> to point to the <em>tag</em> of the Release used in the previous task. </p>
<!-- /wp:paragraph -->

<!-- wp:paragraph -->
<p>o Re-run the Optional Tests on both the WPiOS and WPAndroid PRs.</p>
<!-- /wp:paragraph -->

<!-- wp:paragraph -->
<p>o Main apps PRs should be ready to merge to their <code>develop</code> branches now. Merge them or get them merged.</p>
<!-- /wp:paragraph -->

<!-- wp:group -->
<div class="wp-block-group"><!-- wp:paragraph -->
<p>o Once everything is merged, send a message similar to the following to our friends in the <code>#platform9</code> Slack channel. If the release is a beta/hot fix (e.g. X.XX.2), be sure to directly mention the relevant Excellence Wranglers for the release and modify the following template as needed.</p>
<!-- /wp:paragraph -->

<!-- wp:quote -->
<blockquote class="wp-block-quote"><p>Hey team. I wanted to let you know that the mobile Gutenberg team has finished integrating the X.XX.X Gutenberg release into the WPiOS and WPAndroid `develop` branches. The integration is ready for the next release cut/build creation when you are available. Please let me know if you have any questions. Thanks! </p></blockquote>
<!-- /wp:quote --></div>
<!-- /wp:group -->

<!-- wp:heading {"level":3} -->
<h3>Sync the Release to Development Branches</h3>
<!-- /wp:heading -->

<!-- wp:paragraph -->
<p>o If there are any conflicts in the gutenberg PR, merge <code>trunk</code> into it and resolve them.</p>
<!-- /wp:paragraph -->

<!-- wp:paragraph -->
<p>o In gutenberg-mobile, create a branch from <code>trunk</code> with a name like <code>merge_release_X.XX.X_to_develop</code> and open PR to <code>develop</code>. If there are any merge conflicts, merge <code>develop</code> into the PR and resolve them.</p>
<!-- /wp:paragraph -->

<!-- wp:paragraph -->
<p>o Merge the gutenberg PR to <code>trunk</code>.</p>
<!-- /wp:paragraph -->

<!-- wp:paragraph -->
<p>o Update the <code>gutenberg</code> reference in the gutenberg-mobile <code>merge_release_X.XX.X_to_develop</code> PR so it points to merge commit in gutenberg <code>trunk</code> for the gutenberg PR merged in the previous task.</p>
<!-- /wp:paragraph -->

<!-- wp:paragraph -->
<p>o Merge the gutenberg-mobile PR to <code>develop</code>.</p>
<!-- /wp:paragraph -->

<!-- wp:heading {"level":3} -->
<h3>Clean Up Pending Work (After main apps cut)</h3>
<!-- /wp:heading -->

<!-- wp:paragraph -->
<p>⚠️ This section may only be completed <em>after</em> the main apps cut their own release branches. </p>
<!-- /wp:paragraph -->

<!-- wp:paragraph -->
<p>o Update the <code>gutenberg/after_X.XX.X</code> branches and open a PR against <code>develop</code>. If the branches are empty we’ll just delete them. The PR can actually get created as soon as something gets merged to the <code>gutenberg/after_X.XX.X</code> branches. Merge the <code>gutenberg/after_X.XX.X</code> PR(s) only <em>AFTER</em> the main apps have cut their release branches.</p>
<!-- /wp:paragraph -->

<!-- wp:heading {"level":3} -->
<h3>Test the Release</h3>
<!-- /wp:heading -->

<!-- wp:paragraph -->
<p>ℹ️ Use the main WP apps to complete each the tasks below for both iOS and Android. </p>
<!-- /wp:paragraph -->

<!-- wp:paragraph -->
<p>o Test the new changes that are included in the release PR.</p>
<!-- /wp:paragraph -->

<!-- wp:paragraph -->
<p>o Complete the <a href="https://github.com/wordpress-mobile/test-cases/tree/master/test-cases/gutenberg/writing-flow">general writing flow test cases</a>.</p>
<!-- /wp:paragraph -->

<!-- wp:paragraph -->
<p>o Complete the <a href="https://github.com/wordpress-mobile/test-cases/blob/trunk/test-cases/gutenberg/unsupported-block-editing.md#unsupported-block-editing---test-cases">Unsupported Block Editor test cases</a>.</p>
<!-- /wp:paragraph -->

<!-- wp:paragraph -->
<p>o Verify the <a href="https://manakinp2.wordpress.com/team-rotations/sanity-testing-rotations/">scheduled team members</a> completed the <a href="https://github.com/wordpress-mobile/test-cases/blob/trunk/test-suites/gutenberg/sanity-test-suites.md">sanity test suites</a>.</p>
<!-- /wp:paragraph -->

<!-- wp:heading {"level":3} -->
<h3>Finish the Release</h3>
<!-- /wp:heading -->

<!-- wp:paragraph -->
<p>o Update the <a href="https://docs.google.com/spreadsheets/d/15U4v6zUBmPGagksHX_6ZfVA672-1qB2MO8M7HYBOOgQ/edit?usp=sharing">Release Incident Spreadsheet</a> with any fixes that occurred after the release branches were cut.</p>
<!-- /wp:paragraph -->

<!-- wp:paragraph -->
<p>o If this is a scheduled release (e.g. X.XX.0), message the next release wrangler in the <code>#mobile-gutenberg-platform</code> Slack channel <strong>providing them with a tentative schedule</strong> for the next release. This will help ensure a smooth hand off and sets expectations for when they should begin their work. </p>
<!-- /wp:paragraph -->

<!-- wp:paragraph -->
<p>o Celebrate! 🎉</p>
<!-- /wp:paragraph -->
```


</p>
</details>

# Different types of releases

## Best practices

It's best practice to use the automation script (mentioned in the release template above) for all releases types (regular, betafix, hotfix). When wrangling a betafix or hotfix, it's important to merge the fix to Gutenberg `trunk` independently of the release process. When the release is cut (by the automation script) the commit(s) that make up the betafix or hotfix should then be cherry-picked onto the Gutenberg release branch.

## 1. Regular

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

## 2. Betafix

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

## 3. Hotfix

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
