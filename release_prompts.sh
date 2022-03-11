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

confirm_to_proceed() {
  read -r -p "${tty_bold}$1 (y/n) ${tty_reset}" -n 1
  echo ""
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
      echo -e "\nStoping the release... see you later 👋 "
      exit 0
  fi
}

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

[[ "${GBM_DRY_RUN}" -gt "0" ]] && show_dry_run_warning

# Prompt for release version if not provided as an argument
gbm_release_version="${1:-}"

if [[ -z "$gbm_release_version" ]]; then
  gbm_current_version=$(curl -s https://raw.githubusercontent.com/${GBM_WP_MOBILE_OWNER}/gutenberg-mobile/trunk/package.json | jq ".version" --raw-output)
  echo "Please specify a version. The current version is ${tty_underline}${gbm_current_version}"
  read -r -p "Enter a version number to release: " gbm_release_version
  if [[ -z "$gbm_release_version" ]]; then
    abort "Version number cannot be empty."
  fi
fi

if ! [[ "$gbm_release_version" =~ [0-9]*\.[0-9]*\.[0-9]* ]]; then
  abort "A valid version is required."
fi

# Preflight checks

check_command_message="◦ Checking for required commands.......📡 "
check_aztec_message="◦ Checking Aztec version...............📡 "
check_prs_message="◦ Checking for existing release PRs....📡 "
check_milestone_message="◦ Checking for open milestone PRs......📡 "

update_check_message() {
  local warning_count=${1:-}
  local check_message=${2:-}

  result=$([[ "$warning_count" -gt "0" ]] && echo ⚠️ || echo ✅)
  tput cuu 1
  echo "${check_message%??}${result}"
}

echo -e "${tty_bold}🔍 Running preflight checks:\n${tty_reset}"

## Verify command dependencies: gh and jq
verify_command_version() {
    [[ "$(printf '%s\n' "$2" "$3" | sort -V | head -n1)" == "$2" ]] && return
    printf "\n${tty_red}%s${tty_reset}\n"  "$1 is unavailable or out of date, please install $1 at or above '$2'"
}

gh_version=$(gh version | tail -1 | xargs basename) 2>/dev/null
! verify_command_version "gh" "v2.2.0" "$gh_version" && exit 1
! verify_command_version "jq" "jq-1.6" "$(jq --version)" && exit 1

echo "${check_command_message%??}✅"

## Verify Aztec version
echo "$check_aztec_message"
aztec_verification=$(curl -sSL https://raw.githubusercontent.com/wordpress-mobile/release-toolkit-gutenberg-mobile/add/extract-verify-aztec-script/verify_aztec_version.sh | bash)
aztec_version_warning=0
if [[ -n "$aztec_verification" ]]; then
  aztec_version_warning=1
fi
update_check_message "$aztec_version_warning" "$check_aztec_message"

## Check for open PRs for the release version
echo "$check_prs_message"
pr_count() {
  local repo=$1
  local branch=$2
  gh pr list --repo "$repo" --head "$branch"  --json number,title,url --jq '. | length'
}

existing_release_warning=0
gbm_release_pr_count=$(pr_count "${GBM_WP_MOBILE_OWNER}/gutenberg-mobile" "release/${gbm_release_version}")
gb_release_pr_count=$(pr_count "${GBM_GUTENBERG_OWNER}/gutenberg" "rnmobile/release/${gbm_release_version}")
wpandroid_integration_pr_count=$(pr_count "${GBM_WP_MOBILE_OWNER}/WordPress-Android"  "gutenberg/integrate_release_${gbm_release_version}" )
wpios_integration_pr_count=$(pr_count "${GBM_WP_MOBILE_OWNER}/WordPress-iOS"  "gutenberg/integrate_release_${gbm_release_version}" )

existing_release_warning=$((gbm_release_pr_count + gb_release_pr_count + wpandroid_integration_pr_count + wpios_integration_pr_count))
update_check_message "$existing_release_warning" "$check_prs_message"


## Verify milestone prs
echo "$check_milestone_message"
# Look for both full and short version milestones, that is both "1.0.0" and "1.0"
release_milestone_pr_count=$(gh api -X GET search/issues -f q="repo:${GBM_WP_MOBILE_OWNER}/gutenberg-mobile milestone:${gbm_release_version} is:open" -q '.total_count')
gbm_short_version=$(cut -d '.' -f 1-2 <<< "$gbm_release_version")
short_release_milestone_pr_count=$(gh api -X GET search/issues -f q="repo:${GBM_WP_MOBILE_OWNER}/gutenberg-mobile milestone:${gbm_short_version} is:open" -q '.total_count')

release_milestone_pr_count=$((release_milestone_pr_count + short_release_milestone_pr_count))
update_check_message "$release_milestone_pr_count" "$check_milestone_message"


# Prompt for any warning overides before proceeding

## Continue with Aztec release warnings
if [[ "$aztec_version_warning" -gt "0" ]]; then
  echo -e "\n⚠️  ${tty_bold}Aztec verification failed. Please check the version of Aztec:${tty_reset}\n"
  echo "$aztec_verification"
  echo ""
  confirm_to_proceed "Do you want to continue?"
fi

## Continue with existing release PRs
list_prs() {
  local repo=$1
  local branch=$2
  gh pr list --repo "$repo" --head "$branch" --json number,title,url,createdAt --template \
   "{{range .}}{{tablerow (printf \"◦ ${repo} #%v\" .number | autocolor \"green\") (printf \"'%s' – created %s –\" .title (timeago .createdAt)) .url}}{{end}}"
}
if [[ "$existing_release_warning" -gt "0" ]] ; then

  echo -e "\n⚠️  ${tty_bold}There are some existing PRs for the Gutenberg Mobile release ${gbm_release_version}:${tty_reset}\n"
  list_prs "${GBM_WP_MOBILE_OWNER}/gutenberg-mobile" "release/${gbm_release_version}"
  list_prs  "${GBM_GUTENBERG_OWNER}/gutenberg" "rnmobile/release/${gbm_release_version}"
  list_prs "${GBM_WP_MOBILE_OWNER}/WordPress-Android"  "gutenberg/integrate_release_${gbm_release_version}"
  list_prs "${GBM_WP_MOBILE_OWNER}/WordPress-iOS"  "gutenberg/integrate_release_${gbm_release_version}"
  echo ""
  confirm_to_proceed "Do you want to continue?"
fi

## Continue with milestone prs warnings
if [[ "$release_milestone_pr_count" -gt "0" ]]; then

  echo -e "\n⚠️  ${tty_bold}There are currently $release_milestone_pr_count PR(s) with a milestone matching $gbm_release_version:${tty_reset} (or $gbm_short_version)\n"
  gh api -X GET search/issues -f q="repo:${GBM_WP_MOBILE_OWNER}/gutenberg-mobile milestone:${gbm_release_version} is:open" --template \
  "{{range .items}}{{tablerow (printf \"◦ ${GBM_WP_MOBILE_OWNER}/gutenberg-mobile #%v\" .number | autocolor \"green\") (printf \"'%s' – created %s –\" .title (timeago .created_at)) .url}}{{end}}"

  gh api -X GET search/issues -f q="repo:${GBM_WP_MOBILE_OWNER}/gutenberg-mobile milestone:${gbm_short_version} is:open" --template \
  "{{range .items}}{{tablerow (printf \"◦ ${GBM_WP_MOBILE_OWNER}/gutenberg-mobile #%v\" .number | autocolor \"green\") (printf \"'%s' – created %s –\" .title (timeago .created_at)) .url}}{{end}}"
  echo ""
  confirm_to_proceed "Do you want to continue?"
fi

echo -e "\n🎉 ${tty_bold}${tty_cyan}Ready to set up release PRs for ${tty_underline}v${gbm_release_version}${tty_reset}.\n"
echo -e "${tty_bold}${tty_cyan}Continuing will use or create the following branches and PRs:${tty_reset}"
echo -e "${tty_underline}${tty_cyan}Head → Base                                                                                            ${tty_reset}"

echo_create_or_use() { [[ "${1:-}" -gt "0" ]] && echo -n "${tty_green}◦ (use)   " || echo -n "${tty_cyan}◦ (create)"; echo -e " ${2:-}"; }
echo_create_or_use "$gb_release_pr_count" "${GBM_GUTENBERG_OWNER}/gutenberg/rnmobile/release_${gbm_release_version}                    → ${GBM_GUTENBERG_OWNER}gutenberg/master${tty_reset}"
echo_create_or_use "$gbm_release_pr_count" "${GBM_WP_MOBILE_OWNER}/gutenberg-mobile/release/${gbm_release_version}                      → ${GBM_WP_MOBILE_OWNER}gutenberg-mobile/master${tty_reset}"
echo_create_or_use "$wpandroid_integration_pr_count" "${GBM_WP_MOBILE_OWNER}/WordPress-Android/gutenberg/integrate_release_${gbm_release_version} → ${GBM_WP_MOBILE_OWNER}WordPress-Android/master${tty_reset}"
echo_create_or_use "$wpandroid_integration_pr_count" "${GBM_WP_MOBILE_OWNER}/WordPress-iOS/gutenberg/integrate_release_${gbm_release_version}     → ${GBM_WP_MOBILE_OWNER}WordPress-iOS/master${tty_reset}\n"

# If is tty ?
if [[ -t 0 ]]; then
  confirm_to_proceed "Do you want to continue?"
fi

[[ "${GBM_DRY_RUN}" -gt "0" ]] && show_dry_run_warning

echo "🔧 STUB Creating Gutenberg and Gutenberg Mobile release branches..."
sleep 1
tput cuu 1
echo "🔧 STUB Creating Gutenberg and Gutenberg Mobile release branches...✅"


echo "🔧 STUB Creating iOS and Android release integration branches......"
sleep 1
tput cuu 1
echo "🔧 STUB Creating iOS and Android release integration branches......✅"

cat << EOF

${tty_underline}${tty_cyan}Resulting PRs (STUB):                                                                  ${tty_reset}${tty_cyan}
◦ Gutenberg:        https://github.com/WordPress/gutenberg/pull/STUB
◦ Gutenberg Mobile: https://github.com/wordpress-mobile/gutenberg-mobile/pull/STUB
◦ WPAndriod:        https://github.com/wordpress-mobile/WordPress-Android/pull/STUB
◦ WPiOS:            https://github.com/wordpress-mobile/WordPress-iOS/pull/STUB
EOF
echo -e "\n${tty_reset}${tty_bold}🚀  All set and have fun! 🚀 ${tty_reset}"
