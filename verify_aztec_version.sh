#!/bin/bash
set -euo pipefail

if [[ -n "${GBM_DEBUG:-}" ]]; then
    set -x
fi
# Aztec Version Check
#
# Checks for the latest version of Aztec for Android and iOS
# If ran inside of the Gutenberg Mobile directory, it will check local files.
# Otherwise it will fetch the versions from Gutenberg and Gutenberg Mobile origin repos.
#
# A warning message will be displayed if the version on either platform can not be verified.
#
# Usage: ./verify_aztec_version.sh [GBM branch] [GB commit sha]
#
# Options:
#   [GBM branch] The GB mobile branch to check. Defaults to trunk.
#   [GB commit sha] The GB commit sha to check. Defaults to revision of Gutenberg that is in the GBM branch.
#
# Environment Variables:
#  GBM_GUTENBERG_OWNER: The owner of the Gutenberg repo. Defaults to WordPress.
#  GBM_WP_MOBILE_OWNER: The owner of the Gutenberg Mobile repo. Defaults to WordPress.


GBM_REPO_OWNER=${GBM_REPO:="wordpress-mobile"}
GB_REPO_OWNER=${GB_REPO:="WordPress"}

gbm_branch=${1:-"trunk"}
gb_sha=${2:-}

fetch_aztec_version() {
  # temporarily remove the pipefail so grep doesn't exit when there is no match
  set +o pipefail
  local source="$1"
  local version_key="$2"
  curl -sSL "$source" 2>/dev/null || cat "$source" | grep "$version_key" | head -1 | grep -oE "\d+.\d+.\d+" | cat
  set -o pipefail
}


## Check for Android Aztec version.
aztec_android_gradle_source="$(pwd)/gutenberg/packages/react-native-aztec/build.gradle"
if [[ -z "$gb_sha" ]] || [[ ! -f "$aztec_android_gradle_source" ]]; then
  gbm_tree_sha=$(gh api "/repos/${GBM_REPO_OWNER}/gutenberg-mobile/commits/${gbm_branch}" -q '.commit.tree.sha')
  gb_sha=$(gh api "/repos/${GBM_REPO_OWNER}/gutenberg-mobile/git/trees/${gbm_tree_sha}" -q '.tree | .[] | select(.path == "gutenberg") | .sha')
fi

if [[ ! -f "$aztec_android_gradle_source" ]]; then
  aztec_android_gradle_source="https://raw.githubusercontent.com/${GB_REPO_OWNER}/gutenberg/${gb_sha}/packages/react-native-aztec/android/build.gradle"
fi

aztec_android_version=$(fetch_aztec_version "$aztec_android_gradle_source" "aztecVersion")

if [[ -z "$aztec_android_version" ]]; then
  echo "A release version for WordPress-Aztec-Android was not found in $aztec_android_gradle_source"
fi

## Chekc for iOS Aztex version.
aztec_ios_podspec_source="$(pwd)/RNTAztecView.podspec"
if [[ ! -f "$aztec_ios_podspec_source" ]]; then
  aztec_ios_podspec_source="https://raw.githubusercontent.com/${GBM_REPO_OWNER}/gutenberg-mobile/${gbm_branch}/RNTAztecView.podspec"
fi
aztec_ios_version=$(fetch_aztec_version "$aztec_ios_podspec_source" "WordPress-Aztec-iOS")
if [[ -z "$aztec_ios_version" ]]; then
  echo "A release version for WordPress-Aztec-iOS was not found in $aztec_ios_podspec_source"
fi
