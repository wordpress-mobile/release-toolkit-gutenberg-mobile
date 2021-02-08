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
GUTENBERG_PR_LABEL="Mobile App Android/iOS"
WPANDROID_PR_LABEL="gutenberg-mobile"
WPIOS_PR_LABEL="Gutenberg integration"
# 3. Ensure that each of your repos contains the target branch listed below:
GUTENBERG_MOBILE_TARGET_BRANCH="trunk"
GUTENBERG_TARGET_BRANCH="master"
WPANDROID_TARGET_BRANCH="develop"
WPIOS_TARGET_BRANCH="develop"
# 4. Update the repo names below to the user repo name for your fork
GUTENBERG_REPO="WordPress"
MOBILE_REPO="wordpress-mobile"
#GUTENBERG_REPO="YOUR_GITHUB_USERNAME"
#MOBILE_REPO="YOUR_GITHUB_USERNAME"
# 5. Clone the forked gutenberg-mobile repo

# Before creating the release, this script performs the following checks:
# - AztecAndroid and WordPress-Aztec-iOS are set to release versions
# - Release is being created off of either develop, main, or release/*
# - Release is being created off of a clean branch
# - Whether there are any open PRs targeting the milestone for the release

set -e

# Read GB-Mobile PR template
PR_TEMPLATE_PATH='./release_pull_request.md'
test -f "$PR_TEMPLATE_PATH" || abort "Error: Could not find PR template at $PR_TEMPLATE_PATH"
PR_TEMPLATE=$(cat "$PR_TEMPLATE_PATH")

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
source ./release_prechecks.sh "$GB_MOBILE_PATH"

# Execute script commands from gutenberg-mobile directory
cd "$GB_MOBILE_PATH"

# Check that Github CLI is installed
command -v gh >/dev/null || abort "Error: The Github CLI must be installed."

# Check that jq is installed
command -v jq >/dev/null || abort "Error: jq must be installed."

# Check that Aztec versions are set to release versions
aztec_version_problems="$(check_android_and_ios_aztec_versions)"
if [[ -n "$aztec_version_problems" ]]; then
    warn "There appear to be problems with the Aztec versions:\n$aztec_version_problems"
    confirm_to_proceed "Do you want to proceed with the release despite the ^above^ problem(s) with the Aztec version?"
else
    ohai "Confirmed that Aztec Libraries are set to release versions. Proceeding..."
fi

## Check current branch is develop, trunk, or release/* branch
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [[ ! "$CURRENT_BRANCH" =~ ^develop$|^trunk$|^release/.* ]]; then
    warn "Releases should generally only be based on 'develop', 'trunk', or an earlier release branch."
    warn "You are currently on the '$CURRENT_BRANCH' branch."
    confirm_to_proceed "Are you sure you want to create a release branch from the '$CURRENT_BRANCH' branch?"
fi

# Confirm branch is clean
[[ -z "$(git status --porcelain)" ]] || { git status; abort "Uncommitted changes found. Aborting release script..."; }

# Ask for new version number
CURRENT_VERSION_NUMBER=$(jq '.version' package.json --raw-output)
echo "Current Version Number:$CURRENT_VERSION_NUMBER"
read -r -p "Enter the new version number: " VERSION_NUMBER
if [[ -z "$VERSION_NUMBER" ]]; then
    abort "Version number cannot be empty."
fi

# Ensure javascript dependencies are up-to-date
read -r -p "Run 'npm ci' to ensure javascript dependencies are up-to-date? (y/n) " -n 1
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    execute "npm" "ci"
fi

# If there are any open PRs with a milestone matching the release version number, notify the user and ask them if they want to proceed
number_milestone_prs=$(check_if_version_has_pending_prs_for_milestone "$VERSION_NUMBER")
if [[ -n "$number_milestone_prs" ]] && [[ "0" != "$number_milestone_prs" ]]; then
    echo "There are currently $number_milestone_prs PRs with a milestone matching $VERSION_NUMBER."
    confirm_to_proceed "Do you want to proceed with cutting the release?"
fi

# Create Git branch
RELEASE_BRANCH="release/$VERSION_NUMBER"
ohai "Create Git branch '$RELEASE_BRANCH' in gutenberg-mobile."
execute "git" "switch" "-c" "$RELEASE_BRANCH" 

# Create Git branch in Gutenberg
GB_RELEASE_BRANCH="rnmobile/release_$VERSION_NUMBER"
ohai "Create Git branch '$GB_RELEASE_BRANCH' in gutenberg."
cd gutenberg
execute "git" "switch" "-c" "$GB_RELEASE_BRANCH" 
cd ..

# Set version numbers
ohai "Set version numbers in package.json files"
for file in 'package.json' 'package-lock.json' 'gutenberg/packages/react-native-aztec/package.json' 'gutenberg/packages/react-native-bridge/package.json' 'gutenberg/packages/react-native-editor/package.json'; do
    TEMP_FILE=$(mktemp)
    execute "jq" ".version = \"$VERSION_NUMBER\"" "$file" > "$TEMP_FILE" "--tab"
    execute "mv" "$TEMP_FILE" "$file"
done

# Commit react-native-aztec, react-native-bridge, react-native-editor version update
ohai "Commit react-native-aztec, react-native-bridge, react-native-editor version update version update"
cd gutenberg
git add 'packages/react-native-aztec/package.json' 'packages/react-native-bridge/package.json' 'packages/react-native-editor/package.json'
execute "git" "commit" "-m" "Release script: Update react-native-editor version to $VERSION_NUMBER"
cd ..

# Commit gutenberg-mobile version updates
ohai "Commit gutenberg-mobile version updates"
git add 'package.json' 'package-lock.json'
execute "git" "commit" "-m" "Release script: Update gb mobile version to $VERSION_NUMBER"

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

# Ask if a cherry-pick is needed before bundling (for example if this is a hotfix release)
cd gutenberg
read -r -p "Do you want to cherry-pick a commit from gutenberg? (y/n) " -n 1
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    read -r -p "Enter the commit hash to cherry-pick: " GUTENBERG_COMMIT_HASH_TO_CHERRY_PICK
    execute "git" "cherry-pick" "$GUTENBERG_COMMIT_HASH_TO_CHERRY_PICK"
fi
cd ..

# Commit updates to gutenberg submodule
ohai "Commit updates to gutenberg submodule"
execute "git" "add" "gutenberg"
execute "git" "commit" "-m" "Release script: Update gutenberg ref"

# Update the bundles
ohai "Update the bundles"
npm run bundle || abort "Error: 'npm bundle' failed.\nIf there is an error stating something like \"Command 'bundle' unrecognized.\" above, perhaps try running 'rm -rf node_modules gutenberg/node_modules && npm ci'."

# Commit bundle changes
ohai "Commit bundle changes"
execute "git" "add" "bundle/"
execute "git" "commit" "-m" "Release script: Update bundle for: $VERSION_NUMBER"


#####
# Create PRs
#####

# Verify before creating PRs
confirm_to_proceed "Do you want to proceed with creating a Gutenberg-Mobile PR for the $RELEASE_BRANCH branch and a Gutenberg PR for the $GB_RELEASE_BRANCH branch."
ohai "Proceeding to create PRs..."

#####
# Gutenberg-Mobile PR
#####

# Replace version number in GB-Mobile PR template
PR_BODY=${PR_TEMPLATE//v1.XX.Y/$VERSION_NUMBER}

execute "git" "push" "-u" "origin" "HEAD"

# Create Draft GB-Mobile Release PR in GitHub
GB_MOBILE_PR_URL=$(execute "gh" "pr" "create" \
"--title" "Release $VERSION_NUMBER" \
"--body" "$PR_BODY" \
"--repo" "$MOBILE_REPO/gutenberg-mobile" \
"--head" "$MOBILE_REPO:$RELEASE_BRANCH" \
"--base" "$GUTENBERG_MOBILE_TARGET_BRANCH" \
"--label" "$GUTENBERG_MOBILE_PR_LABEL" \
"--draft")

#####
# Gutenberg PR
#####

# Get Checklist from Gutenberg PR template
cd gutenberg
GUTENBERG_PR_TEMPLATE_PATH=".github/PULL_REQUEST_TEMPLATE.md"
test -f "$GUTENBERG_PR_TEMPLATE_PATH" || abort "Error: Could not find PR template at $GUTENBERG_PR_TEMPLATE_PATH" 
# Get the checklist from the gutenberg PR template by removing everything before the '## Checklist:' line
CHECKLIST_FROM_GUTENBERG_PR_TEMPLATE=$(sed -e/'## Checklist:'/\{ -e:1 -en\;b1 -e\} -ed < "$GUTENBERG_PR_TEMPLATE_PATH")

# Construct body for Gutenberg release PR
GUTENBERG_PR_BEGINNING="## Description
Release $VERSION_NUMBER of the react-native-editor and Gutenberg-Mobile.

For more information about this release and testing instructions, please see the related Gutenberg-Mobile PR: $GB_MOBILE_PR_URL"
GUTENBERG_PR_BODY="$GUTENBERG_PR_BEGINNING

$CHECKLIST_FROM_GUTENBERG_PR_TEMPLATE"

execute "git" "push" "-u" "origin" "HEAD"

# Create Draft Gutenberg Release PR in GitHub
GUTENBERG_PR_URL=$(execute "gh" "pr" "create" \
"--title" "Mobile Release v$VERSION_NUMBER" \
"--body" "$GUTENBERG_PR_BODY" \
"--repo" "$GUTENBERG_REPO/gutenberg" \
"--head" "$GUTENBERG_REPO:$GB_RELEASE_BRANCH" \
"--base" "$GUTENBERG_TARGET_BRANCH" \
"--label" "$GUTENBERG_PR_LABEL" \
"--draft")
cd ..

echo "PRs Created"
echo "==========="
printf "Gutenberg-Mobile PR %s \n Gutenberg %s \n" "$GB_MOBILE_PR_URL" "$GUTENBERG_PR_URL" | column -t

confirm_to_proceed "Do you want to proceed with creating main apps (WPiOS and WPAndroid) PRs?"
ohai "Proceeding to create main apps PRs..."

GB_MOBILE_PR_REF=$(git rev-parse HEAD)

WP_APPS_PR_TITLE="Integrate gutenberg-mobile release $VERSION_NUMBER"

WP_APPS_PR_BODY="## Description
This PR incorporates the $VERSION_NUMBER release of gutenberg-mobile.  
For more information about this release and testing instructions, please see the related Gutenberg-Mobile PR: $GB_MOBILE_PR_URL

Release Submission Checklist

- [ ] I have considered if this change warrants user-facing release notes and have added them to \`RELEASE-NOTES.txt\` if necessary."

WP_APPS_INTEGRATION_BRANCH="gutenberg/integrate_release_$VERSION_NUMBER"


#####
# WPAndroid PR
#####
read -r -p "Do you want to target $WPANDROID_TARGET_BRANCH branch for WPAndroid PR? (y/n) " -n 1
echo ""
if [[ $REPLY =~ ^[Nn]$ ]]; then
    read -r -p "Enter the branch name you want to target. Make sure a branch with this name already exists in WPAndroid repository: " WPANDROID_TARGET_BRANCH
fi

TEMP_WP_ANDROID_DIRECTORY=$(mktemp -d)
ohai "Clone WordPress-Android into '$TEMP_WP_ANDROID_DIRECTORY'"
execute "git" "clone" "-b" "$WPANDROID_TARGET_BRANCH" "--depth=1" "git@github.com:$MOBILE_REPO/WordPress-Android.git" "$TEMP_WP_ANDROID_DIRECTORY"

cd "$TEMP_WP_ANDROID_DIRECTORY"

execute "git" "submodule" "update" "--init" "--recursive" "--depth=1" "--recommend-shallow"

ohai "Create after_x.xx.x branch in WordPress-Android"
execute "git" "switch" "-c" "gutenberg/after_$VERSION_NUMBER" 

execute "git" "push" "-u" "origin" "HEAD"

ohai "Create release branch in WordPress-Android"
execute "git" "switch" "-c" "$WP_APPS_INTEGRATION_BRANCH"

ohai "Update gutenberg-mobile ref"
cd libs/gutenberg-mobile
execute "git" "fetch" "--recurse-submodules=no" "origin" "$GB_MOBILE_PR_REF"
execute "git" "checkout" "$GB_MOBILE_PR_REF"
execute "git" "submodule" "update"
cd ../..

execute "git" "add" "libs/gutenberg-mobile"
execute "git" "commit" "-m" "Release script: Update gutenberg-mobile ref"

ohai "Update strings"
execute "python" "tools/merge_strings_xml.py"
# If merge_strings_xml.py results in changes, commit them
if [[ -n "$(git status --porcelain)" ]]; then
    ohai "Commit changes from 'python tools/merge_strings_xml.py'"
    execute "git" "add" "WordPress/src/main/res/values/strings.xml"
    execute "git" "commit" "-m" "Release script: Update strings"
else
    ohai "There were no changes from 'python tools/merge_strings_xml.py' to be committed."
fi

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

read -r -p "Do you want to target $WPIOS_TARGET_BRANCH branch for WPiOS PR? (y/n) " -n 1
echo ""
if [[ $REPLY =~ ^[Nn]$ ]]; then
    read -r -p "Enter the branch name you want to target. Make sure a branch with this name already exists in WPiOS repository: " WPIOS_TARGET_BRANCH
fi

TEMP_WP_IOS_DIRECTORY=$(mktemp -d)
ohai "Clone WordPress-iOS into '$TEMP_WP_IOS_DIRECTORY'"
execute "git" "clone" "-b" "$WPIOS_TARGET_BRANCH" "--depth=1" "git@github.com:$MOBILE_REPO/WordPress-iOS.git" "$TEMP_WP_IOS_DIRECTORY"

cd "$TEMP_WP_IOS_DIRECTORY"

ohai "Create after_x.xx.x branch in WordPress-iOS"
execute "git" "switch" "-c" "gutenberg/after_$VERSION_NUMBER" 

execute "git" "push" "-u" "origin" "HEAD"

ohai "Create release branch in WordPress-iOS"
execute "git" "switch" "-c" "gutenberg/integrate_release_$VERSION_NUMBER" 

ohai "Update GitHub organization and gutenberg-mobile ref"
test -f "Podfile" || abort "Error: Could not find Podfile"
sed -i'.orig' -E "s/(podspec_prefix = \"https:\/\/raw.githubusercontent.com\/)wordpress-mobile(\/gutenberg-mobile\/#{tag_or_commit}\")/\1$MOBILE_REPO\2/" Podfile || abort "Error: Failed updating GitHub organization in Podfile"
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

echo "Main apps PRs created"
echo "==========="
printf "WPAndroid %s \n WPiOS %s \n" "$WP_ANDROID_PR_URL" "$WP_IOS_PR_URL" | column -t

