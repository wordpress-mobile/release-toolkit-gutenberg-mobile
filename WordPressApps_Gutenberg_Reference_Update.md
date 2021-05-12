# Generating WordPress-Android/iOS gutenberg reference update PRs

The `./wp_gutenberg_ref_update_prs.sh` script can be used for automatically generating Pull Requests in the [WordPress-Android ](https://github.com/wordpress-mobile/WordPress-Android)and [WordPress-iOS](https://github.com/wordpress-mobile/WordPress-iOS) GitHub repositories that integrate changes from a specific [gutenberg-mobile](https://github.com/wordpress-mobile/gutenberg-mobile) branch. 

---

  - [Preqrequisites](#preqrequisites)
  - [Usage](#usage)
  - [Expected Script Runtime (Approximate):](#expected-script-runtime-approximate)
  - [Testing](#testing)
  - [FAQ](#faq)
    - [Why do we need to generate a gutenberg-mobile PR? I just want Apps integration PRs for testing.](#why-do-we-need-to-generate-a-gutenberg-mobile-pr-i-just-want-apps-integration-prs-for-testing)
    - [Why do we need a gutenberg branch?](#why-do-we-need-a-gutenberg-branch)
    - [Why not just use the exising `./release_automation.sh` script for generating builds?](#why-not-just-use-the-exising-release_automationsh-script-for-generating-builds)
    - [My WordPress-Android PR has failed CI for missing gutenberg-mobile bridge on S3, why?](#my-wordpress-android-pr-has-failed-ci-for-missing-gutenberg-mobile-bridge-on-s3-why)
    - [What do I do if the script exited halfway for some reason?](#what-do-i-do-if-the-script-exited-halfway-for-some-reason)

---

## Preqrequisites:

1. To be able to run the automation scripts make sure you have installed:

[Github CLI](https://github.com/cli/cli)
```sh
brew install gh
```
[jq](https://github.com/stedolan/jq)
```sh
brew install jq
```

2. Locally clone an up to date version of this release-toolkit-gutenberg-mobile repository (`develop` branch)
```sh
git clone git@github.com:wordpress-mobile/release-toolkit-gutenberg-mobile.git
```

3. Local clone of the gutenberg-mobile repository, checked out to the specific branch that you would like to generate WordPress Apps Pull Requests for.

```sh
git clone git@github.com:wordpress-mobile/gutenberg-mobile.git
git checkout <GB_MOBILE_BRANCH_TO_GENERATE_PRS_FROM>
```
---
## Usage

Run the script: `./wp_gutenberg_ref_update_prs.sh`

You will be prompted for:
1. The local path to your gutenberg-mobile clone (press enter for default: sibling directory of script)
2. The new branch name for your gutenberg-mobile branch that will be used for bundle updates (press enter for default: `$CURRENT_VERSION_NUMBER-$CURRENT_BRANCH-${CURRENT_HASH:0:6}`). Note if branch already exists, script will fail.
3. The WordPress-Android branch to target (default is `develop`)
4. The WordPress-iOS branch to target (default is `develop`)
5. Whether to run `npm ci` in your gutenberg-mobile cloned directory before beginning
   1. If it is your first time running the script click yes to running `npm ci`
   2. If you are troubleshooting the script, or you are sure npm dependencies of your gutenberg-mobile local directory are up to date, you can type 'N' to skip this step.


The output of the script will be links to PRs in the gutenberg-mobile, WordPress-Android, and WordPress-iOS repositories, as well as the name of a gutenberg repository branch. Don't forget to clean up branches / PRs if they are only used for testing!

---
## Expected Script Runtime (Approximate):

1. Tested *without* running `npm ci`: ~16 minutes 56s
2. Tested *with* running  `npm ci`: ~18 minutes 22s

---
## Testing

If you would like to try out the script before running against the wordpress-mobile GitHub repos, you can follow the instructions at the top of the `./wp_gutenberg_ref_update_prs.sh` script file for running the script against forked repos. 

---
## FAQ

### Why do we need to generate a gutenberg-mobile PR? I just want Apps integration PRs for testing.
 
 The "[Build Android RN Bridge & Publish to S3](https://github.com/wordpress-mobile/gutenberg-mobile/blob/ed7a64d9d8d82af942f52628ae4b64d8f1010c6a/.circleci/config.yml#L256-L284)" CI job only happens for gutenberg-mobile commits associated with an open PR, and we need that RN Bridge on S3 in order to run WordPress-Android Pull Request CI properly

### Why do we need a gutenberg branch?

Prior to generating gutenberg mobile bundle we run a `npm run core preios` command that may generate Podfile related updates in the gutenberg repo. I think we need to include these changes, but if I am wrong please open a PR to remove this step

### Why not just use the exising `./release_automation.sh` script for generating builds?

The  `./release_automation.sh` includes extra checks for making sure that gutenberg-mobile branch is ready for generating a release. Some of these checks would prevent quickly running the script against a non release ready branch for quickly generating WPApps test PRs. See [here](https://github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/blob/d718e1c0732f1c422d427f0fbe0eaa968f978da9/release_automation.sh#L29).

### My WordPress-Android PR has failed CI for missing gutenberg-mobile bridge on S3, why?

he Gutenberg-Mobile "Build Android RN Bridge & Publish to S3" job on CI must finish before WordPress-Android CI will be able to access the gutenberg mobile bridge from S3. Please verify the "Build Android RN Bridge & Publish to S3" job on your Gutenberg-Mobile PR is finished, and then restart failed tests on your WordPress-Android PR and this failure should go away.

### What do I do if the script exited halfway for some reason?

While troubleshooting, you should return to your pre-script state by navigating to your gutenberg-mobile directory checking out the branch or commit you originally wished to use in your WordPress Apps PRs. You may also need to delete branches locally in you gutenberg-mobile and gutenberg-mobile/gutenberg directories, and delete branches and PRs from the wordpress-mobile (or your fork) GitHub remote repositories. 
