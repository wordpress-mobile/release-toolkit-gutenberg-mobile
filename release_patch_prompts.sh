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
tty_reset="$(tty_escape 0)"
tty_cyan="$(tty_escape 96)"
tty_bold="$(tty_escape "1;")"
tty_green="$(tty_escape 32)"

abort() { printf "${tty_bold}${tty_red}Error: %s${tty_reset}\n" "$1" && exit 1; }

gbm_version="${1:-}"

shift
passed_commit_shas="${*: }"
passed_commit_shas="${passed_commit_shas// /,}"

if [[ -z "${gbm_version}" ]]; then
  abort "Missing version argument"
fi

if [[ $(cut -d '.' -f 3 <<< "${gbm_version}") -eq 0 ]]; then
  abort "Version must be a patch release, ${gbm_version} is a scheduled release"
fi

major_version=$(cut -d. -f 1 <<< "${gbm_version}")
minor_version=$(cut -d. -f 2 <<< "${gbm_version}")
patch_version=$(cut -d. -f 3 <<< "${gbm_version}")

trunk_version=$(curl -sSL https://raw.githubusercontent.com/${GBM_WP_MOBILE_OWNER}/gutenberg-mobile/trunk/package.json | jq ".version" --raw-output)
trunk_major_version=$(cut -d. -f 1 <<< "${trunk_version}")
trunk_minor_version=$(cut -d. -f 2 <<< "${trunk_version}")
trunk_patch_version=$(cut -d. -f 3 <<< "${trunk_version}")

## If the version in trunk is above the requested verision we are patching the latest stable release.
## We'll also be patching the beta release.
beta_patch_version=''
if [[ "${trunk_minor_version}" -gt "${minor_version}" ]] || [[ "$trunk_major_version" -gt "${major_version}" ]]; then
  echo "Notice: There is a newer version of gutenberg-mobile in trunk: ${trunk_version}"
  beta_patch_version="${trunk_major_version}.${trunk_minor_version}.$((${trunk_patch_version}+1))"
fi

last_patch_version=$(($(cut -d. -f 3 <<< "${gbm_version}")-1))
prior_version="${major_version}.${minor_version}.${last_patch_version}"
cherry_pick_commits=""

commit_message_template='{{tablerow ((slice .sha 0 10) | autocolor "red") (printf "- %-72s" (truncate 72 .commit.message)) ((printf "(%s)" (timeago .commit.author.date)) | autocolor "green") ((printf "<%s>" .commit.author.name) | autocolor "cyan")}}'

if [[ -n "$passed_commit_shas" ]]; then
  echo -e "\n${tty_bold}Validating passed in commits:${tty_reset}${tty_green} $* ${tty_reset}\n"

  IFS=' ' read -r -a commits <<< "$passed_commit_shas"
  for commit in "${commits[@]}"; do
    gh api "repos/$GBM_GUTENBERG_OWNER/gutenberg/commits/$commit" -t "$commit_message_template" 2>/dev/null || { echo -e "Remvoving invalid commit: $commit\n"; continue; }
    read -r -p "${tty_bold}Do you want to cherry-pick this commit? [y/n]: ${tty_reset}"
    if [[ $REPLY =~ ^[Yy]$ ]]; then
      cherry_pick_commits+="$commit "
    else
      echo -e "${tty_red}Removing commit: ${tty_reset}$commit"
    fi
    echo ""
  done

  echo ""
  read -r -p "${tty_bold}That's all the passed in commits, do you want to cherry-pick more commits? [y/n]: ${tty_reset}"
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    interactive_cherry_pick="false"
  fi
  echo ""
else
  interactive_cherry_pick="true"
fi

last_release_commited_at=$(gh api "repos/$GBM_GUTENBERG_OWNER/gutenberg/commits/rnmobile/${prior_version}" -q ".commit.committer.date" --cache="10m")

if [[ "$interactive_cherry_pick" = "true" ]]; then
  echo -e "\n${tty_cyan}Choose some commits to cherry pick off $GBM_GUTENBERG_OWNER/gutenberg/trunk.${tty_reset}
  Try to pick in chronological order, starting with the oldest.\n"

  show_commits() {
    echo "Showing commits since $last_release_commited_at"
    gh api "repos/WordPress/gutenberg/commits?since=$last_release_commited_at" --paginate -t "{{range .}}$commit_message_template{{end}}" | less -R
  }

  while read -r -p "${tty_bold}Enter commit SHA${tty_reset} [ x to exit selection, s to show commits since ${prior_version} ]: " commit_sha; do
    if [[ "$commit_sha" == "x" ]]; then
      echo ""
      break
    fi
    if [[ "$commit_sha" == "s" ]]; then
      show_commits
      continue
    fi
    echo ""
    gh api repos/"$GBM_GUTENBERG_OWNER"/gutenberg/commits/"$commit_sha" -t "$commit_message_template" 2>/dev/null || { echo -e "Invalid commit\n"; continue; }
    read -r -p "${tty_bold}Do you want to cherry-pick this commit? [y/n]: ${tty_reset}"
    if [[ $REPLY =~ ^[Yy]$ ]]; then
      cherry_pick_commits+="$commit_sha "
    fi
    echo ""
  done
fi

echo -e "🎉 ${tty_cyan}Ready to create the patch release ${tty_bold}${tty_underline}${gbm_version}${tty_reset}"
echo -e "${tty_cyan}Continueing will use cherry pick the following commits from gutenberg:${tty_reset}\n"
IFS=' ' read -r -a commits <<< "${cherry_pick_commits%?}"
for commit in "${commits[@]}"; do
  gh api "repos/$GBM_GUTENBERG_OWNER/gutenberg/commits/$commit" -t "$commit_message_template"
done


show_line(){
  local repo=${1:-}
  local cut_from=${2:-}
  local branch=${3:-}
  local merge_to=${4:-}

  local sm_padding="                  "
  local md_padding="                          "
  local lg_padding="                                    "
  printf "${tty_green}%s%s${tty_reset}${tty_bold}%s%s%s%s${tty_reset}${tty_cyan}%s%s${tty_reset}\n"  "$repo" "${sm_padding:${#repo}}" "$cut_from" "${md_padding:${#cut_from}}" "$branch" "${lg_padding:${#branch}}" "$merge_to" "${lg_padding:${#merge_to}}"
}
echo -e "\n${tty_cyan}With the following branches and PRs:\n"
echo "${tty_underline}${tty_cyan}Repo              Cut From                        Branch                        Merge To         ${tty_reset}"
show_line "Gutenberg"         "rnmobile/release_${prior_version}" "rnmobile/release_${gbm_version}"              "trunk"
show_line "Gutenberg Mobile"  "release/${prior_version}"          "release/${gbm_version}"                       "trunk"
show_line "WPAndroid"         "TBD"                               "gutenberg/integrate_release_${gbm_version}"   "TBD"
show_line "WPiOS"             "TBD"                               "gutenberg/integrate_release_${gbm_version}"   "TBD"

echo ""
read -r -p "${tty_reset}${tty_bold}Do you want to continue with the ${gbm_version} release? [y/n]: "
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
  echo -e "\nStoping the release... see you later 👋 "
  exit 0
fi


source ./prepare_gutenberg_release.sh \
--gb-shallow-since "$last_release_commited_at" \
--gb-release-head "rnmobile/release_${prior_version}" \
--cherry-pick "${cherry_pick_commits%?}" \
${gbm_version}


create_beta_release="false"
if [[ -n "${beta_patch_version}" ]]; then
  echo -e "\n${tty_cyan}There is a newer version of gutenberg-mobile in trunk: ${tty_bold}${tty_underline}${trunk_version}${tty_reset}"
  echo -e "\n${tty_cyan}A beta patch release would utilize the following branches and PRs:"
  echo "${tty_underline}${tty_cyan}Repo              Cut From                      Branch                      Merge To          ${tty_reset}"
  show_line "Gutenberg"         "rnmobile/release_${trunk_version}" "rnmobile/release_${beta_patch_version}"              "trunk"
  show_line "Gutenberg Mobile"  "release/${trunk_version}"          "release/${beta_patch_version}"                       "trunk"
  show_line "WPAndroid"         "TBD"                               "gutenberg/integrate_release_${beta_patch_version}"   "TBD"
  show_line "WPiOS"             "TBD"                               "gutenberg/integrate_release_${beta_patch_version}"   "TBD"

  echo ""
  read -r -p "${tty_reset}${tty_bold}Do you want to continue with the ${beta_patch_version} release? [y/n]: "
  if [[ $REPLY =~ ^[Yy]$ ]]; then
    create_beta_release="true"
  fi
fi

if [[ "${create_beta_release}" == "true" ]]; then
  source ./prepare_gutenberg_release.sh \
  --gb-shallow_since "$last_release_commited_at" \
  --gb-release_head "release/${trunk_version}" \
  --cherry-pick "${cherry_pick_commits%?}" \
  ${beta_patch_version}
fi
