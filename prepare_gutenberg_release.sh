#!/bin/bash
set -euo pipefail

if [[ -n "${GBM_DEBUG:-}" ]]; then
    set -x
fi

GBM_GUTENBERG_OWNER="${GBM_GUTENBERG_OWNER:=WordPress}"
GBM_WP_MOBILE_OWNER="${GBM_WP_MOBILE_OWNER:=wordpress-mobile}"
GBM_DRY_RUN="${GBM_DRY_RUN:=1}"

usage() {
  echo "
Usage: $0 [options] <version>

Options:
  -a | --allow-non-standard-branch
    Allow non-standard branches, e.g. release-1.0.0-beta.1. Default false

  -z | --skip-aztec-verify
    Skip the aztec verification step. Default false

  -m | --gbm-release-head trunk
    Release from the given branch on gutenberg-mobile.

  -g | --gb-release-head trunk
    Release from the given branch on gutenberg.

  -c | --cherry-pick
    Commit to cherry pick from gutenberg/trunk into the gutenberg release branch. Add multiple cherry-pick commits with multiple -c flags.

  -s|--gb-shallow-since
    Shallow clone the gutenberg repo from the given date, See


  -h | --help
  Show this help.

Examples:
$ # Prepare release 1.0.0 from gutenberg-mobile/trunk and gutenberg/trunk
$ $0 1.0.0

$ # Prepare point release from gutenberg/rnmobile/1.0.0 and cherry pick commits '313lol' and '415rofl'
$ $0 --gb-release-head \"rnmobile/1.0.0\" --cherry-pick \"313lol,415rofl\" 1.0.1
"
exit 2
}

# Set up default options
gbm_release_head="trunk"
gb_release_head="trunk"
cherry_pick_commits=""
gb_shallow_since=$(date -u +'%Y-%m-%dT%H:%M:%SZ')

skip_aztec_verify=0
allow_non_standard_branches=0

for i in "$@"; do
  case $i in
    -m|--gbm-release-head)
      gbm_release_head="${2:-}"
      shift
      shift
      ;;
    -g|--gb-release-head)
      gb_release_head="${2:-}"
      shift
      shift
      ;;
    -c|--cherry-pick)
      cherry_pick_commits+="${2:-},"
      shift
      shift
      ;;
    -s|--gb-shallow-since)
      gb_shallow_since="${2:-}"
      shift
      shift
      ;;
    -a|--allow-non-standard-branch)
      allow_non_standard_branches=1
      shift
      ;;
    -z|--skip-aztec-verify)
      skip_aztec_verify=1
      shift
      ;;
    -h|--help)
      usage
      ;;
    -*|--*)
      echo "Unknown option $i"
      exit 1
      ;;
    *)
      ;;
  esac
done

cherry_pick_commits=${cherry_pick_commits%?}
gbm_release_version="${1:-}"

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
                           ❗️ Dry Run Enabled ❗️
                 ~~ Nothing will be pushed to remote repos ~~
${tty_reset}
EOF
}

[[ "${GBM_DRY_RUN}" -gt "0" ]] && show_dry_run_warning


if [[ -z "$gbm_release_version" ]] || ! [[ "$gbm_release_version" =~ [0-9]*\.[0-9]*\.[0-9]* ]]; then
  error "{$tty_red}Error: A valid version is required."
  usage
fi

gbm_release_branch="release/$gbm_release_version"
gb_release_branch="rnmobile/release/$gbm_release_version"

# Validations

## Check to see if the is an open Gutenberg release PR exists
should_create_gb_release_pr=$(gh pr view "$gb_release_branch" --repo "$GBM_GUTENBERG_OWNER/gutenberg" --json "closed" --jq ".closed" 2>/dev/null || echo "true")

if [[ "$should_create_gb_release_pr" = false ]]; then
  echo "Found Gutenberg release PR for $gbm_release_version. Verify the release is ready or close the PR then try again:"
  gh pr view "$gb_release_branch" --repo "$GBM_GUTENBERG_OWNER/gutenberg"
  exit 1
fi

## Check for non standard branches
if [[ "$allow_non_standard_branches"  -eq "0" ]] && [[ ! "$gbm_release_head" =~ ^trunk$|v[0-9]+\.[0-9]+\.[0-9]+ ]]; then
  warn "Looks like you're trying to release from '$gbm_release_head'. This is not recommended."
  abort "Releases should generally only be based on 'trunk', or a release tag."
fi

# Initialize Gutenberg Mobile and Gutenberg for release
git clone "https://github.com/$GBM_WP_MOBILE_OWNER/gutenberg-mobile" \
  --branch "$gbm_release_head" \
  --single-branch \
  --recurse \
  --shallow-submodules \
  --depth 1

cd gutenberg-mobile
git switch "$gbm_release_branch"  2>/dev/null || git checkout -b "$gbm_release_branch"

# Setup Gutenberg for release
cd gutenberg
git remote set-branches origin "$gb_release_head" "$gbm_release_branch"
git switch "$gb_release_head"
git fetch --shallow-since "$gb_shallow_since" origin "$gb_release_head"
git switch "$gb_release_branch" 2>/dev/null || git switch -c "$gb_release_branch"

if [[ -n "$cherry_pick_commits" ]]; then
  git cherry-pick "$cherry_pick_commits"
fi

# hop back up to install the node packages and run the Aztec check.
cd -
npm ci
if [[ "$skip_aztec_verify" -eq "0" ]]; then
  ## TODO change aztec_verification url to 'trunk' after https://github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/pull/98 is merged
  aztec_verification_script="https://raw.githubusercontent.com/wordpress-mobile/release-toolkit-gutenberg-mobile/add/extract-verify-aztec-script/verify_aztec_version.sh"
  aztec_verification=$(curl -sSL "$aztec_verification_script" | bash -s "$gbm_release_head")
  if [[ -n "$aztec_verification" ]] && [[ $(echo "$aztec_verification" | wc -l | xargs) -gt "0" ]]; then
    warn "Aztec version verification failed. Please check the version and try again. Errors:"
    echo "\n$aztec_verification"
    exit 1
  fi
fi
cd -

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

npm run bundle
git add bundle/
git commit -m "Release script: Update bundle for $gbm_release_version"

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
