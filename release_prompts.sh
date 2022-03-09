#!/bin/bash
set -euo pipefail

if [[ -n "${GBM_DEBUG:-}" ]]; then
    set -x
fi

GBM_GUTENBERG_OWNER="${GBM_GUTENBERG_ORG:=WordPress}"
GBM_WP_MOBILE_OWNER="${GBM_WP_MOBILE_ORG:=wordpress-mobile}"

## Output helpers
if [[ -t 1 ]]; then
    tty_escape() { printf "\033[%sm" "$1"; }
else
    tty_escape() { :; }
fi

tty_mkbold() { tty_escape "1;$1"; }
tty_underline="$(tty_escape "4;39")"
tty_blue="$(tty_mkbold 34)"
tty_red="$(tty_mkbold 31)"
tty_reset="$(tty_escape 0)"
tty_cyan="$(tty_mkbold 96)"
tty_bold="$(tty_escape "1;")"


info() { printf "${tty_blue} %s${tty_reset}\n" "$1";}
warn() { printf "${tty_underline}${tty_red}Warning${tty_reset}: %s\n" "$1"; }
error() { printf "${tty_red}Error: %s${tty_reset}\n" "$1"; }
abort() { error "$1" && exit 1; }

confirm_to_proceed() {
    read -r -p "$1 (y/n) " -n 1
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        abort "Aborting release..."
    fi
}

verify_command_version() {
    [[ "$(printf '%s\n' $2 $3 | sort -V | head -n1)" == "$2" ]] && return
    printf "\n${tty_red}%s${tty_reset}\n"  "$1 is unavailable or out of date, please install $1 at or above '$2'"
    false
}
echo ""
info "📡 Checking for required commands...."
## Verify command dependencies: gh and jq
gh_version=$(gh version | tail -1 | xargs basename) 2>/dev/null
! verify_command_version "gh" "v2.2.0" "$gh_version"; gh_verified=$?
! verify_command_version "jq" "jq-1.6" $(jq --version); jq_verified=$?

( [[ $gh_verified -eq "0" ]] || [[ $jq_verified -eq "0" ]] ) && exit 1


tput cuu 1
info "📡 Checking for required commands....✅"
## Prompt for release version if not provided as an argument
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

info "📡 Check for existing release PRs...."

pr_count() {
  local repo=$1
  local branch=$2
  gh pr list --repo "$repo" --head "$branch"  --json number,title,url --jq '. | length'
}

list_prs() {
  local repo=$1
  local branch=$2
  gh pr list --repo "$repo" --head "$branch" --json number,title,url,createdAt --template \
   "{{range .}}{{tablerow (printf \"${repo} #%v\" .number | autocolor \"green\") (printf \"'%s' – created %s –\" .title (timeago .createdAt)) .url}}{{end}}"
}

gbm_release_pr_count=$(pr_count "${GBM_WP_MOBILE_OWNER}/gutenberg-mobile" "release/${gbm_release_version}")
gb_release_pr_count=$(pr_count "${GBM_GUTENBERG_OWNER}/gutenberg" "rnmobile/release/${gbm_release_version}")
wpandroid_integration_pr_count=$(pr_count "${GBM_WP_MOBILE_OWNER}/WordPress-Android"  "gutenberg/integrate_release_${gbm_release_version}" )
wpios_integration_pr_count=$(pr_count "${GBM_WP_MOBILE_OWNER}/WordPress-iOS"  "gutenberg/integrate_release_${gbm_release_version}" )

existing_pr_count=$(($gbm_release_pr_count + $gb_release_pr_count + $wpandroid_integration_pr_count + $wpios_integration_pr_count))

if [[ $existing_pr_count -gt 0 ]] ; then
  warn "There are some existing PRs for the Gutenberg Mobile release ${gbm_release_version}."
  list_prs "${GBM_WP_MOBILE_OWNER}/gutenberg-mobile" "release/${gbm_release_version}"
  list_prs  "${GBM_GUTENBERG_OWNER}/gutenberg" "rnmobile/release/${gbm_release_version}"
  list_prs "${GBM_WP_MOBILE_OWNER}/WordPress-Android"  "gutenberg/integrate_release_${gbm_release_version}"
  list_prs "${GBM_WP_MOBILE_OWNER}/WordPress-iOS"  "gutenberg/integrate_release_${gbm_release_version}"
  confirm_to_proceed "Do you want to continue?"
else
  tput cuu 1
  info "📡 Check for existing release PRs....✅"
fi

info "📡 Verifing Aztec...................."
aztec_verification=$(curl -sSL https://raw.githubusercontent.com/wordpress-mobile/release-toolkit-gutenberg-mobile/add/extract-verify-aztec-script/verify_aztec_version.sh | bash)
if [[ $(echo "$aztec_verification" | tr -d '\n' | wc -l | xargs) -gt "0" ]]; then
  warn "Aztec verification failed. Please check the version of Aztec:"
  echo "$aztec_verification"
  confirm_to_proceed "Do you want to continue?"
else
  tput cuu 1
  info "📡 Verifing Aztec....................✅"

fi

## Verify milestone prs
info "📡 Checking for open milestone PRs..."
milestone_pr_count=$(gh api -X GET search/issues -f q="repo:${GBM_WP_MOBILE_OWNER}/gutenberg-mobile milestone:${gbm_release_version} is:open" -q '.total_count')

if [[ "$milestone_pr_count" -eq 0 ]]; then
  # Now check with the short version. for example, if the version is 1.2.0, check for 1.2
  gbm_short_version=$(cut -d '.' -f 1-2 <<< "$gbm_release_version")
  milestone_pr_count=$(gh api -X GET search/issues -f q="repo:${GBM_WP_MOBILE_OWNER}/gutenberg-mobile milestone:${gbm_short_version} is:open" -q '.total_count')
fi

if [[ $milestone_pr_count -gt 0 ]] ; then
  warn "There are currently $milestone_pr_count PRs with a milestone matching $gbm_release_version."
  confirm_to_proceed "Do you want to continue?"
else
  tput cuu 1
  info "📡 Checking for open milestone PRs...✅"
fi

echo ""
echo "🎉 ${tty_bold}Ready to set up release PRs for ${tty_underline}v${gbm_release_version}${tty_reset}."

echo -e "${tty_cyan}"
cat << EOF
Continuing will create the following branches and PRs: ${tty_underline}${tty_cyan}

Head → Base                                                                      ${tty_reset}${tty_cyan}

◦ gutenberg/rnmobile/release_${gbm_release_version}                    → gutenberg/master
◦ gutenberg-mobile/release/${gbm_release_version}                      → gutenberg-mobile/master
◦ WordPress-Android/gutenberg/integrate_release_${gbm_release_version} → WordPress-Android/master
◦ WordPress-iOS/gutenberg/integrate_release_${gbm_release_version}     → WordPress-iOS/master${tty_reset}
EOF

confirm_to_proceed "Do you want to continue?"

info "🔧 Creating Gutenberg and Gutenberg Mobile release branches..."
sleep 1
tput cuu 1
info "🔧 Creating Gutenberg and Gutenberg Mobile release branches...✅"

info "🔧 Creating iOS and Android release integration branches......"
sleep 1
tput cuu 1
info "🔧 Creating iOS and Android release integration branches......✅"

cat << EOF

${tty_underline}${tty_cyan}Resulting Prs:                                               ${tty_reset}${tty_cyan}

◦ Gutenberg:        https://github.com/WordPress/gutenberg/pull/1
◦ Gutenberg Mobile: https://github.com/wordpress-mobile/gutenberg-mobile/pull/1
◦ WPAndriod:        https://github.com/wordpress-mobile/WordPress-Android/pull/1
◦ WPiOS:            https://github.com/wordpress-mobile/WordPress-iOS/pull/1
EOF
echo -e "\n${tty_cyan}🎉 All set!${tty_reset}"
