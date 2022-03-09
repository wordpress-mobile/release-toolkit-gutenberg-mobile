#!/bin/bash
set -euo pipefail

if [[ -n "${GBM_DEBUG:-}" ]]; then
    set -x
fi

GBM_REPO=${GBM_REPO:="wordpress-mobile/gutenberg-mobile"}
GB_REPO=${GB_REPO:="WordPress/gutenberg"}

gbm_branch=${1:-"trunk"}
gb_sha=${2:-}

fetch_aztec_version() {
  # temporarily remove the pipefail so grep doesn't exit when there is no match
  set +o pipefail
  local source="$1"
  local version_key="$2"
  curl -s "$source" | grep "$version_key" | head -1 | grep -oE "\d+.\d+.\d+" | cat
  set -o pipefail
}

if [[ -z "$gb_sha" ]]; then
  gbm_tree_sha=$(gh api "/repos/${GBM_REPO}/commits/trunk" -q '.commit.tree.sha')
  gb_sha=$(gh api "/repos/${GBM_REPO}/git/trees/${gbm_tree_sha}" -q '.tree | .[] | select(.path == "gutenberg") | .sha')
fi

aztec_android_gradle_url="https://raw.githubusercontent.com/${GB_REPO}/${gb_sha}/packages/react-native-aztec/android/build.gradle"
aztec_android_version=$(fetch_aztec_version "$aztec_android_gradle_url" "aztecVersion")

if [[ -z "$aztec_android_version" ]]; then
  echo "A release version for WordPress-Aztec-Android was not found in $aztec_android_gradle_url"
fi

aztec_ios_podspec_url="https://raw.githubusercontent.com/${GBM_REPO}/${gbm_branch}/RNTAztecView.podspec"
aztec_ios_version=$(fetch_aztec_version "$aztec_ios_podspec_url" "WordPress-Aztec-iOS")
if [[ -z "$aztec_ios_version" ]]; then
  echo "A release version for WordPress-Aztec-iOS was not found in $aztec_ios_podspec_url"
fi
