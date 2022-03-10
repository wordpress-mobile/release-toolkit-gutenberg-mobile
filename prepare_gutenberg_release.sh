#!/bin/bash
set -euo pipefail

if [[ -n "${GBM_DEBUG:-}" ]]; then
    set -x
fi

GBM_GUTENBERG_OWNER="${GBM_GUTENBERG_OWNER:=WordPress}"
GBM_WP_MOBILE_OWNER="${GBM_WP_MOBILE_OWNER:=wordpress-mobile}"
GBM_DRY_RUN="${GBM_DRY_RUN:=1}"


## Output helpers
if [[ -t 1 ]]; then
    tty_escape() { printf "\033[%sm" "$1"; }
else
    tty_escape() { :; }
fi

tty_mkbold() { tty_escape "1;$1"; }
tty_underline="$(tty_escape "4;39")"
tty_red="$(tty_escape 31)"
tty_cyan="$(tty_escape 96)"
tty_bold="$(tty_escape "1;")"
tty_green="$(tty_escape 32)"
tty_reset="$(tty_escape 0)"

warn() { printf "${tty_underline}${tty_red}Warning${tty_reset}: %s\n" "$1"; }
abort() { printf "${tty_bold}${tty_red}Error: %s${tty_reset}\n" "$1" && exit 1; }

show_dry_run_warning() {
  cat <<EOF
${tty_red}
################################################################################
#                                                                              #
#                          ❗️ Dry Mode enabled ❗️                              #
#                ~~ Nothing will be pushed to remote repos ~~                  #
#                                                                              #
################################################################################
${tty_reset}
EOF
}

# Parse arguments
usage() {
    cat <<EOF
Usage: $0 ????

EOF
}

gbm_release_head="trunk"
gb_release_head="trunk"

skip_aztec_verify=0
allow_non_stable_branches=0

gbm_release_version="${1:-}"

[[ "${GBM_DRY_RUN}" -gt "0" ]] && show_dry_run_warning


if [[ -z "$gbm_release_version" ]] || ! [[ "$gbm_release_version" =~ [0-9]*\.[0-9]*\.[0-9]* ]]; then
  error "{$tty_red}Error: A valid version is required."
  usage && exit 1
fi

# Validations
if [[ "$allow_non_stable_branches"  -eq "0" ]] && [[ ! "$gbm_release_head" =~ ^trunk$|v[0-9]+\.[0-9]+\.[0-9]+ ]]; then
  warn "Looks like you're trying to release from '$gbm_release_head'. This is not recommended."
  abort "Releases should generally only be based on 'trunk', or a release tag."
fi

## TODO change 'add/extract-verify-aztec-script' to 'trunk' after https://github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/pull/98 is merged
if [[ "$skip_aztec_verify" -eq "0" ]]; then
  aztec_verification=$(curl -sSL https://bit.ly/gbm-toolkit--verify-aztec | bash -s "$gbm_release_head")
  if [[ -n "$aztec_verification" ]] && [[ $(echo "$aztec_verification" | wc -l | xargs) -gt "0" ]]; then
    warn "Aztec version verification failed. Please check the version and try again. Errors:"
    echo "\n$aztec_verification"
    exit 1
  fi
fi

# Clone Gutenberg Mobile
# Check to see if there is an existing Gutenberg Mobile release PR
# NOTE: If there is an existing PR, can we assume the changes are ready ?
gbm_release_branch="release/$gbm_release_version"

# Check if the release branch already exists
#gbm_release_branch_exists=$(gh api "repos/${GBM_WP_MOBILE_OWNER}/mobile-gutenberg/branches" --jq '(map(select(.name == "update-release-checklist") | .name)) | first')

#if [[ -n "$gbm_release_branch_exists" ]]; then
#  gbm_clone_branch="$gbm_release_branch"
#else
#  gbm_clone_branch="$gbm_release_head"
#fi

# Initialize Guteberg Mobile for release
git clone "https://github.com/$GBM_WP_MOBILE_OWNER/gutenberg-mobile" \
  --branch "$gbm_release_head" \
  --single-branch \
  --recurse \
  --shallow-submodules \
  --depth 1

cd gutenberg-mobile
git switch -c "$gbm_release_branch"

npm ci

# Setup Gutenberg for release
cd gutenberg
gb_release_branch="rnmobile/release/$gbm_release_version"
git switch -c "$gb_release_branch"

## Update the verison in gutenberg package.json files
for file in \
  'packages/react-native-aztec/package.json' \
  'packages/react-native-bridge/package.json' \
  'packages/react-native-editor/package.json'; do
  jq ".version = \"$gbm_release_version\"" $file | ex -sc "wq!${file}" /dev/stdin
done

git add \
  packages/react-native-aztec/package.json \
  packages/react-native-bridge/package.json \
  packages/react-native-editor/package.json

git commit -m "Release script: Update react-native-aztec, react-native-bridge, and react-native-editor to version $gbm_release_version"

## Update podfile by native 'preios' command
npm run native "preios"
if [[ -n "$(git status --porcelain)" ]]; then
  git commit -a -m "Release script: Update with changes from 'run native preios'"
fi

## Push up to gutenberg repo before we create the GBM PR (if not a dry run)
[[ "$GBM_DRY_RUN" -eq "0" ]] && git push origin HEAD

# Setup Gutenberg Mobile for release
cd ..
for file in 'package.json' 'package-lock.json'; do
  jq ".version = \"$gbm_release_version\"" $file | ex -sc "wq!${file}" /dev/stdin
done
git add package.json package-lock.json
git commit -m "Release script: Update gb mobile to version $gbm_release_version"

#npm run bundle
#git add bundle/
#git commit -m "Release script: Update bundle for $gbm_release_version"

git add gutenberg
git commit -m "Release script: Update gutenberg ref"

if [[ "$GBM_DRY_RUN" -gt "0" ]]; then
  show_dry_run_warning
  echo -e "\n${tty_bold} Dry run complete – see you later 👋 ${tty_reset}"
  exit 0
fi


# Create release PRs

## Gutenberg Mobile PR
git push origin HEAD
gbm_pr_template_url="https://raw.githubusercontent.com/wordpress-mobile/release-toolkit-gutenberg-mobile/trunk/release_pull_request.md"
gbm_pr_template=$(curl -sSL "$gbm_pr_template_url")
if [[ "$gbm_pr_template" =~ 404 ]]; then
  abort "Unable to fetch PR template from $gbm_pr_template_pr. Please verify the file exists."
fi
gbm_pr_body=${gbm_pr_template//v1.XX.Y/"$gbm_release_version"}

gbm_pr_url=$(gh pr create \
--title "Release $gbm_release_version" \
--body "$gbm_pr_body" \
--head "$gbm_release_branch" \
--repo "$GBM_WP_MOBILE_OWNER/gutenberg-mobile" \
--label "release-process" \
--draft | tail -1)


## Gutenberg PR
cd gutenberg
gb_pr_template_path=".github/PULL_REQUEST_TEMPLATE.md"
gb_pr_checklist=$(sed -e/'## Checklist:'/\{ -e:1 -en\;b1 -e\} -ed < "$gb_pr_template_path")
gb_pr_body="## Description
Release $gbm_release_version of the react-native-editor and Gutenberg-Mobile.

For more information about this release and testing instructions, please see the related Gutenberg-Mobile PR: $gbm_pr_url

$gb_pr_checklist"

gh pr create \
--title "Mobile Release v$gbm_release_version" \
--body "$gb_pr_body" \
--head "$gb_release_branch" \
--repo "$GBM_GUTENBERG_OWNER/gutenberg" \
--label "Mobile App - i.e. Android or iOS" \
--draft

## All done!
echo "$gbm_pr_url"