#!/bin/bash
set -euo pipefail

script_path="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

source ./release_utils.sh

command -v gh >/dev/null || abort "Error: The Github CLI must be installed."

release_type=""
release_date=""
gb_mobile_version=""
main_apps_version=""
issue_note=""
issue_number=""
include_aztec_steps=""
include_incoming_changes=""
debug_template=""
auto_confirm=""

while getopts "t:v:d:g:m:i:axhuy" opt; do
  case ${opt} in
    h )
      echo "options:"
      echo "   -t Release type. accepts scheduled|beta|hotfix"
      echo "   -v Gutenberg Mobile release version"
      echo "   -m Mobile app version"
      echo "   -d Release date in YYYY-MM-DD format. Only used to generate scheduled release checklists"
      echo "   -g Relative path to guteberg mobile"
      echo "   -n Issue note"
      echo "   -i Existing issue to update"
      echo "   -a Include aztec steps"
      echo "   -u Include incomming steps"
      echo "   -x Echo out generated template with out sending to github"
      echo "   -y Auto confirm creating gh calls"
      exit 0
      ;;
    t )
      release_type=$OPTARG
     ;;
    v )
      gb_mobile_version=$OPTARG
      ;;
    d )
      release_date=$OPTARG
      ;;
    g )
      set_gb_mobile_path $OPTARG
      ;;
    m )
      main_apps_version=$OPTARG
      ;;
    i )
      issue_number=$OPTARG
      ;;
    n )
      issue_note=$OPTARG
      ;;
    a )
      include_aztec_steps="true"
      ;;
    u )
      include_incoming_changes="true"
      ;;
    x )
      debug_template="true"
      ;;
    y )
      auto_confirm="true"
      ;;
    \? )
      echo "Invalid Option: -$OPTARG" 1>&2
      exit 1
      ;;
  esac
done
shift $((OPTIND -1))


# If including an existing github issue, look for adding aztec or incoming changes checklists
if [[ -n "$issue_number" ]]; then

  if [[ -z "$include_aztec_steps" && -z "$include_incoming_changes" ]]; then
    abort "Nothing to update, please set the -a or -u flags to include Aztec or Incomming Changes steps."
  fi
  pushd_gb_mobile
    issue_json=$(gh issue view "$issue_number" --json 'title,url')
    issue_title=$(jq '.title' <<< "$issue_json")
    issue_url=$(jq '.url' <<< "$issue_json")

    confirm_message="Ready to update $issue_title $issue_url ?"

    # Pulling the body from the gh call above i.e. with --json 'body,title,url'
    # strips the newlines from the body. The call below does not.
    issue_body=$(gh issue view "$issue_number" --json 'body' --jq '.body')

    issue_comment="Issue updated:"

    if [[ -n "$include_aztec_steps" ]]; then
      aztec_checklist_template=$(<"$script_path/templates/release_checklist_update_aztec.md")
      aztec_checklist_template=${aztec_checklist_template//$'\n'/\\n}
      issue_body=$(sed -e "s/.*optional_aztec_release.*/$aztec_checklist_template/" <<< "$issue_body")
      issue_comment="$issue_comment"$'\n\n- Added Aztec Release Checklist'
      confirm_message="$confirm_message"$'\n\n- Add Aztec Release Checklist'
    fi

    if [[ -n "$include_incoming_changes" ]]; then
      incoming_change_note=$issue_note
      if [[ -n "$incoming_change_note" ]]; then
        incoming_change_note="**Update:** $issue_note"
      fi
      incoming_changes_template=$(sed -e "s/{{incoming_change_note}}/${incoming_change_note}/g" "$script_path/templates/release_checklist_update_incoming.md")
      incoming_changes_template=${incoming_changes_template//$'\n'/\\n}

      issue_body=$(sed -e "s/.*optional_incoming_changes.*/${incoming_changes_template}/" <<< "$issue_body")
      issue_comment="$issue_comment"$'\n\n- Added Incoming Changes Checklist'
      confirm_message="$confirm_message"$'\n\n- Add Incoming Changes Checklist'
    fi

    if [[ -n "$issue_note" ]]; then
      issue_comment="$issue_comment"$'\n\n'"$issue_note"
    fi

    if [[ -n "$debug_template" ]]; then
      echo "$issue_body"
      exit 0
    fi

    if [[ -z "$auto_confirm" ]]; then
      echo ""
      confirm_to_proceed "$confirm_message"$'\n\n'
    fi

    gh issue comment "$issue_number" --body "$issue_comment">/dev/null
    gh issue edit "$issue_number" --body "$issue_body"
  popd_gb_mobile

  exit 0;
fi

if [[ -z "$release_type" ]]; then
  default_release_type="scheduled"

  read -r -p "Please enter release type: scheduled|beta|hotfix [$default_release_type]: " release_type
  release_type=${release_type:-$default_release_type}
  echo ""
fi

if [[ ! "$release_type" == "scheduled" && ! "$release_type" == "beta" && ! "$release_type" == "hotfix" ]]; then
    abort "Error release type must be scheduled|beta|hotfix, you entered '$release_type'"
fi

if [[ -z "$gb_mobile_version" ]]; then
  pushd_gb_mobile

  # Ask for new version number
  current_gb_mobile_version=$(jq '.version' package.json --raw-output)
  IFS='.' read -r -a version_array <<< "$current_gb_mobile_version"

  if [[ "$release_type" == "scheduled" ]]; then
    default_gb_mobile_version="${version_array[0]}.$((version_array[1] + 1)).${version_array[2]}"
  else
    default_gb_mobile_version="${version_array[0]}.${version_array[1]}.$((version_array[2] + 1))"
  fi

  read -r -p "Enter the new version number [$default_gb_mobile_version]: " gb_mobile_version
  echo ""
  gb_mobile_version=${gb_mobile_version:-$default_gb_mobile_version}

  if [[ -z "$gb_mobile_version" ]]; then
      abort "Version number cannot be empty."
  fi

  popd_gb_mobile
fi

# Propmt for release date if scheduled release
if [[ -z "$release_date" && "$release_type" == "scheduled" ]]; then
    if command -v dateround &> /dev/null; then
      default_release_date=$(dateround today Thurs)

      read -r -p "Enter the release date (YYYY-MM-DD) [$default_release_date]: " release_date
      release_date=${release_date:-$default_release_date}
    else
      read -r -p "Enter the release date (YYYY-MM-DD): " release_date
    fi
    echo ""
fi

main_apps_branch="develop"

# Prompt for the main apps version if this is a non-scheduled release
if [[  "$release_type" != "scheduled" ]]; then
  if [[ -z "$main_apps_version" ]]; then
    read -r -p "Enter the main apps version: " main_apps_version

    if [[ -z "$main_apps_version" ]]; then
      abort "Version number cannot be empty."
    fi

    echo ""
  fi
  main_apps_branch="release\/$main_apps_version"
fi

checklist_template_path="$script_path/templates/release_checklist.md"

pushd_gb_mobile
  milestone_url=$(gh api  --method GET repos/:owner/:repo/milestones --jq ".[0].html_url")
  default_milestone_url="https://wordpress-mobile/gutenber-mobile/milestones"

  milestone_url=${milestone_url:-$default_milestone_url}
popd_gb_mobile

checklist_template=$(sed \
-e "s/{{gb_mobile_version}}/${gb_mobile_version}/g" \
-e "s/{{release_date}}/${release_date}/g" \
-e "s/{{release_type}}/${release_type}/g" \
-e "s/{{main_apps_branch}}/${main_apps_branch}/g" \
-e "s/{{before_release_date}}/$(date '+%Y-%m-%d')/g" \
-e "s/{{milestone_url}}/${milestone_url//\//\\/}/g" "$checklist_template_path")

if [[ $release_type == "beta" || $release_type == "hotfix" ]]; then
  release_checklist_template=$(sed "/<!-- scheduled_release_only -->/,/<!-- \/scheduled_release_only -->/d" <<< "$checklist_template")
else
  release_checklist_template=$(sed "/<!-- non_scheduled_release_only -->/,/<!-- \/non_scheduled_release_only -->/d" <<< "$checklist_template")
fi

if [[ -n "$include_aztec_steps" ]]; then
  aztec_checklist_template=$(<"$script_path/templates/release_checklist_update_aztec.md")
  aztec_checklist_template=${aztec_checklist_template//$'\n'/\\n}
  release_checklist_template=$(sed -e "s/.*optional_aztec_release.*/$aztec_checklist_template/" <<< "$release_checklist_template")
fi

issue_title="Release checklist for v$gb_mobile_version ($release_type)"
issue_label="release checklist,$release_type release"
issue_assignee="@me"
issue_body=$'# Release Checklist\n'"This checklist is for the $release_type release v$gb_mobile_version."$'\n\n **Release date:** '"$release_date"$'\n'"$issue_note"$'\n'"$release_checklist_template"

if [[ -n "$debug_template" ]]; then
  echo "$issue_body"
  exit 0;
fi

if [[ -z "$auto_confirm" ]]; then
  echo ""
  confirm_to_proceed "Ready to create '$issue_title' issue ?"
fi

pushd_gb_mobile

  gh issue create --title "$issue_title" --body "$issue_body" --assignee "$issue_assignee" --label "$issue_label"

popd_gb_mobile
