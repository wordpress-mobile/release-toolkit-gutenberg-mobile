#!/bin/bash

# INSTRUCTIONS FOR TESTING FROM FORKED REPOS
# Prerequisites:
# 1. Fork the following repos to your github user repo:
#    a) Gutenberg-Mobile: https://github.com/wordpress-mobile/gutenberg-mobile
#    * .gitmodules on CURRENT_BRANCH should reference your gutenberg fork, replace 'WordPress' with GITHUB_USERNAME
#    * (https://github.com/wordpress-mobile/gutenberg-mobile/blob/develop/.gitmodules)
#    b) Gutenberg: https://github.com/WordPress/gutenberg
#    c) WordPress-Android: https://github.com/wordpress-mobile/WordPress-Android
#    d) WordPress-iOS: https://github.com/wordpress-mobile/WordPress-iOS
# 2. Insure that each of your forked repos contains the PR labels specified below:
GUTENBERG_MOBILE_PR_LABEL="release-process"
WPANDROID_PR_LABEL="gutenberg-mobile"
WPIOS_PR_LABEL="Gutenberg integration"
# 3. Ensure that each of your repos contains the target branch listed below:
GUTENBERG_MOBILE_TARGET_BRANCH="trunk"
WPANDROID_TARGET_BRANCH="develop"
WPIOS_TARGET_BRANCH="develop"
# 4. Update the repo names below to the user repo name for your fork
 MOBILE_REPO="wordpress-mobile"
# MOBILE_REPO="YOUR_GITHUB_USERNAME"
# 5. Clone the forked gutenberg-mobile repo

set -e

SCRIPT_PATH="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
# Ask for path to gutenberg-mobile directory
# (default is sibling directory of gutenberg-mobile-release-toolkit)
DEFAULT_GB_MOBILE_LOCATION="$SCRIPT_PATH/../gutenberg-mobile"

read -r -p "Please enter the path to the gutenberg-mobile directory [$DEFAULT_GB_MOBILE_LOCATION]:" GB_MOBILE_PATH
GB_MOBILE_PATH=${GB_MOBILE_PATH:-"$DEFAULT_GB_MOBILE_LOCATION"}
echo ""
if [[ ! "$GB_MOBILE_PATH" == *gutenberg-mobile ]]; then
    abort "Error path does not end with gutenberg-mobile"
fi

source ./release_utils.sh

# Execute script commands from gutenberg-mobile directory
cd "$GB_MOBILE_PATH"

# Check that Github CLI is installed
command -v gh >/dev/null || abort "Error: The Github CLI must be installed."

# Check that Github CLI is logged
gh auth status >/dev/null 2>&1 || abort "Error: You are not logged into any GitHub hosts. Run 'gh auth login' to authenticate."

# Check that jq is installed
command -v jq >/dev/null || abort "Error: jq must be installed."

## Check current branch is develop, trunk, or release/* branch
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
CURRENT_HASH=$(git rev-parse HEAD)
confirm_to_proceed "Are you sure you want to create WPApps PRs from the gutenberg-mobile '$CURRENT_BRANCH' branch, commit: $CURRENT_HASH ?"

# Confirm branch is clean
[[ -z "$(git status --porcelain)" ]] || { git status; abort "Uncommitted changes found. Aborting release script..."; }

# Ask for new version number
CURRENT_VERSION_NUMBER=$(jq '.version' package.json --raw-output)
echo "Current Version Number:$CURRENT_VERSION_NUMBER"
read -r -p "Enter a name for your gb-mobile branch: " BRANCH_NAME
if [[ -z "$BRANCH_NAME" ]]; then
    abort "Version number cannot be empty."
fi

# Ask for WPAndroid target branch
read -r -p "Do you want to target $WPANDROID_TARGET_BRANCH branch for WPAndroid PR? (y/n) " -n 1
echo ""
if [[ $REPLY =~ ^[Nn]$ ]]; then
    read -r -p "Enter the branch name you want to target. Make sure a branch with this name already exists in WPAndroid repository: " WPANDROID_TARGET_BRANCH
fi

# Ask for WPiOS target branch
read -r -p "Do you want to target $WPIOS_TARGET_BRANCH branch for WPiOS PR? (y/n) " -n 1
echo ""
if [[ $REPLY =~ ^[Nn]$ ]]; then
    read -r -p "Enter the branch name you want to target. Make sure a branch with this name already exists in WPiOS repository: " WPIOS_TARGET_BRANCH
fi

# Ensure javascript dependencies are up-to-date
read -r -p "Run 'npm ci' to ensure javascript dependencies are up-to-date? (y/n) " -n 1
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    execute "npm" "ci"
fi

# Create Git branch
GUTENBERG_MOBILE_BRANCH="$BRANCH_NAME"
ohai "Create Git branch '$GUTENBERG_MOBILE_BRANCH' in gutenberg-mobile."
execute "git" "switch" "-c" "$GUTENBERG_MOBILE_BRANCH"

# Create Git branch in Gutenberg
GB_BRANCH="rnmobile/$GUTENBERG_MOBILE_BRANCH"
ohai "Create Git branch '$GB_BRANCH' in gutenberg."
cd gutenberg
execute "git" "switch" "-c" "$GB_BRANCH"
cd ..

# Make sure podfile is updated
ohai "Make sure podfile is updated"
PRE_IOS_COMMAND="npm run core preios"
eval "$PRE_IOS_COMMAND"

# If preios results in changes, commit them
cd gutenberg
if [[ -n "$(git status --porcelain)" ]]; then
    ohai "Commit changes from '$PRE_IOS_COMMAND'"
    execute "git" "commit" "-a" "-m" "Release script: Update with changes from '$PRE_IOS_COMMAND'"
else
    ohai "There were no changes from '$PRE_IOS_COMMAND' to be committed."
fi
cd ..

# Commit updates to gutenberg submodule
if [[ -n "$(git status --porcelain)" ]]; then
    ohai "Commit updates to gutenberg submodule"
    execute "git" "add" "gutenberg"
    execute "git" "commit" "-m" "Release script: Update gutenberg ref"
else
    ohai "There were no changes from submodule update to be committed."
fi

# Update the bundles
ohai "Update the bundles"
npm run bundle || abort "Error: 'npm bundle' failed.\nIf there is an error stating something like \"Command 'bundle' unrecognized.\" above, perhaps try running 'rm -rf node_modules gutenberg/node_modules && npm ci'."

# Commit bundle changes
if [[ -n "$(git status --porcelain)" ]]; then
    ohai "Commit bundle changes"
    execute "git" "add" "bundle/"
    execute "git" "commit" "-m" "Release script: Update bundle for: $BRANCH_NAME"
else
    ohai "There were no changes from bundle update to be committed."
fi


#####
# Create PRs
#####

# Replace version number in GB-Mobile PR template
PR_BODY="## Description
This PR is for publishing the $GUTENBERG_MOBILE_BRANCH branch of gutenberg-mobile to S3 for testing in WPApps repos.



PR Submission Checklist
- [ ] Update target branch away from trunk unless this is a release PR
- [ ] I have considered if this change warrants user-facing release notes and have added them to \`RELEASE-NOTES.txt\` if necessary."
execute "git" "push" "-u" "origin" "HEAD"

# Create Draft GB-Mobile Release PR in GitHub
GB_MOBILE_PR_URL=$(execute "gh" "pr" "create" \
"--title" "App Integration for $GUTENBERG_MOBILE_BRANCH" \
"--body" "$PR_BODY" \
"--repo" "$MOBILE_REPO/gutenberg-mobile" \
"--head" "$MOBILE_REPO:$GUTENBERG_MOBILE_BRANCH" \
"--base" "$GUTENBERG_MOBILE_TARGET_BRANCH" \
"--label" "$GUTENBERG_MOBILE_PR_LABEL" \
"--draft")

cd gutenberg
execute "git" "push" "-u" "origin" "HEAD"
cd ..

echo "Branches pushed to remote"
echo "==========="

ohai "Proceeding to create main apps PRs..."

GB_MOBILE_PR_REF=$(git rev-parse HEAD)

WP_APPS_PR_TITLE="Integrate changes from gutenberg-mobile branch $GUTENBERG_MOBILE_BRANCH"

WP_APPS_PR_BODY="## Description
This PR incorporates the $GUTENBERG_MOBILE_BRANCH branch of gutenberg-mobile.

Release Submission Checklist

- [ ] I have considered if this change warrants user-facing release notes and have added them to \`RELEASE-NOTES.txt\` if necessary."

WP_APPS_INTEGRATION_BRANCH="gutenberg/integrate_$GUTENBERG_MOBILE_BRANCH"


#####
# WPAndroid PR
#####

TEMP_WP_ANDROID_DIRECTORY=$(mktemp -d)
ohai "Clone WordPress-Android into '$TEMP_WP_ANDROID_DIRECTORY'"
execute "git" "clone" "-b" "$WPANDROID_TARGET_BRANCH" "--depth=1" "git@github.com:$MOBILE_REPO/WordPress-Android.git" "$TEMP_WP_ANDROID_DIRECTORY"

cd "$TEMP_WP_ANDROID_DIRECTORY"

# This is still needed because of Android Stories.
execute "git" "submodule" "update" "--init" "--recursive" "--depth=1" "--recommend-shallow"


ohai "Create integration branch in WordPress-Android"
execute "git" "switch" "-c" "$WP_APPS_INTEGRATION_BRANCH"

# Get the last part of the path from GB_MOBILE_PR_URL
PULL_ID=${GB_MOBILE_PR_URL##*/}

ohai "Update build.gradle file with the latest version"
test -f "build.gradle" || abort "Error: Could not find build.gradle"
sed -i'.orig' -E "s/ext.gutenbergMobileVersion = '(.*)'/ext.gutenbergMobileVersion = '${PULL_ID}-$GB_MOBILE_PR_REF'/" build.gradle || abort "Error: Failed updating gutenbergMobileVersion in build.gradle"

execute "git" "add" "build.gradle"
execute "git" "commit" "-m" "Gutenberg ref update script: Update build.gradle gutenbergMobileVersion to ref"

ohai "Push integration branch"
execute "git" "push" "-u" "origin" "HEAD"

# Create Draft WPAndroid Release PR in GitHub
ohai "Create Draft WPAndroid Release PR in GitHub"
WP_ANDROID_PR_URL=$(execute "gh" "pr" "create" \
"--title" "$WP_APPS_PR_TITLE" \
"--body" "$WP_APPS_PR_BODY" --repo "$MOBILE_REPO/WordPress-Android" \
"--head" "$MOBILE_REPO:$WP_APPS_INTEGRATION_BRANCH" \
"--base" "$WPANDROID_TARGET_BRANCH" \
"--label" "$WPANDROID_PR_LABEL" \
"--draft")

ohai "WPAndroid PR Created: $WP_ANDROID_PR_URL"
echo ""


#####
# WPiOS PR
#####

TEMP_WP_IOS_DIRECTORY=$(mktemp -d)
ohai "Clone WordPress-iOS into '$TEMP_WP_IOS_DIRECTORY'"
execute "git" "clone" "-b" "$WPIOS_TARGET_BRANCH" "--depth=1" "git@github.com:$MOBILE_REPO/WordPress-iOS.git" "$TEMP_WP_IOS_DIRECTORY"

cd "$TEMP_WP_IOS_DIRECTORY"

ohai "Create integration branch in WordPress-iOS"
execute "git" "switch" "-c" "$WP_APPS_INTEGRATION_BRANCH"

ohai "Update GitHub organization and gutenberg-mobile ref"
test -f "Podfile" || abort "Error: Could not find Podfile"
sed -i'.orig' -E "s/wordpress-mobile(\/gutenberg-mobile)/$MOBILE_REPO\1/" Podfile || abort "Error: Failed updating GitHub organization in Podfile"
sed -i'.orig' -E "s/gutenberg :(commit|tag) => '(.*)'/gutenberg :commit => '$GB_MOBILE_PR_REF'/" Podfile || abort "Error: Failed updating gutenberg ref in Podfile"
execute "rake" "dependencies"


execute "git" "add" "Podfile" "Podfile.lock"
execute "git" "commit" "-m" "Release script: Update gutenberg-mobile ref"

ohai "Push integration branch"
execute "git" "push" "-u" "origin" "HEAD"

# Create Draft WPiOS Release PR in GitHub
ohai "Create Draft WPiOS Release PR in GitHub"
WP_IOS_PR_URL=$(execute "gh" "pr" "create" \
"--title" "$WP_APPS_PR_TITLE" \
"--body" "$WP_APPS_PR_BODY" \
"--repo" "$MOBILE_REPO/WordPress-iOS" \
"--head" "$MOBILE_REPO:$WP_APPS_INTEGRATION_BRANCH" \
"--base" "$WPIOS_TARGET_BRANCH" \
"--label" "$WPIOS_PR_LABEL" \
"--draft")

ohai "WPiOS PR Created: $WP_IOS_PR_URL"
echo ""
echo "GB-Mobile PR and branch info:"
echo "==========="

printf "Gutenberg-Mobile PR %s \n Gutenberg Branch %s \n" "$GB_MOBILE_PR_URL" "$GB_BRANCH" | column -t
echo ""
echo "Main apps PRs created"
echo "==========="
printf "WPAndroid %s \n WPiOS %s \n" "$WP_ANDROID_PR_URL" "$WP_IOS_PR_URL" | column -t
echo ""
echo "Please don't forget to delete test branches / PRs when finished. Thank you!"
