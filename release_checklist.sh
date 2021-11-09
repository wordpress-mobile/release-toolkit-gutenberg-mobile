#!/bin/bash
set -euo pipefail

script_path="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

source ./release_utils.sh

command -v gh >/dev/null || abort "Error: The Github CLI must be installed."

release_type=""
release_date=""
version_number=""
gb_mobile_path=""
checklist_message=""
issue_number=""
include_aztec_steps=""
include_incoming_changes=""
debug_template=""
auto_confirm=""
silence_output=""

while getopts "t:v:d:g:m:i:axhuy" opt; do
  case ${opt} in
    h )
      echo "options:"
      echo "   -t release type. accepts scheduled|beta|hotfix"
      echo "   -v release version"
      echo "   -d release date in YYYY-MM-DD format. Only used to generate scheduled release checklists"
      echo "   -g relative path to guteberg mobile"
      echo "   -m additional message"
      echo "   -i existing issue to update"
      echo "   -a include aztec steps"
      echo "   -u include incomming steps"
      echo "   -x echo out generated template with out sending to github"
      echo "   -y auto confirm creating gh calls"
      exit 0
      ;;
    t )
      release_type=$OPTARG
     ;;
    v )
      version_number=$OPTARG
      ;;
    d )
      release_date=$OPTARG
      ;;
    g )
      gb_mobile_path=$OPTARG
      ;;
    m )
      checklist_message=$OPTARG
      ;;
    i )
      issue_number=$OPTARG
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

if [[ -z "$gb_mobile_path" ]]; then
  # Ask for path to gutenberg-mobile directory
  # (default is sibling directory of gutenberg-mobile-release-toolkit)
  default_gb_mobile_location="$script_path/../gutenberg-mobile"
  read -r -p "Please enter the path to the gutenberg-mobile directory [$default_gb_mobile_location]:" gb_mobile_path
  gb_mobile_path=${gb_mobile_path:-"$default_gb_mobile_location"}
  echo ""
else
  gb_mobile_path="$script_path/$gb_mobile_path"
fi

if [[ ! "$gb_mobile_path" == *gutenberg-mobile ]]; then
    abort "Error path does not end with gutenberg-mobile"
fi

# If including an existing github issue, look for adding aztec or incoming changes checklists
if [[ -n "$issue_number" ]]; then

  if [[ -z "$include_aztec_steps" && -z "$include_incoming_changes" ]]; then
    abort "Nothing to update, please set the -a or -u flags to include Aztec or Incomming Changes steps."
  fi
  pushd "$gb_mobile_path" >/dev/null
    issue_json=$(gh issue view $issue_number --json 'title,url')
    issue_title=$(jq '.title' <<< "$issue_json")
    issue_url=$(jq '.url' <<< "$issue_json")


    confirm_message="Ready to update $issue_title $issue_url ?"

    # Pulling the body from the gh call above i.e. with --json 'body,title,url'
    # strips the newlines from the body. The call below does not.
    issue_body=$(gh issue view $issue_number --json 'body' --jq '.body')

    issue_comment="Issue updated:"

    if [[ -n "$include_aztec_steps" ]]; then
      aztec_checklist_template=$(<"$script_path/templates/release_checklist_update_aztec.md")
      aztec_checklist_template=${aztec_checklist_template//$'\n'/\\n}
      issue_body=$(sed -e "s/.*optional_aztec_release.*/$aztec_checklist_template/" <<< "$issue_body")
      issue_comment="$issue_comment"$'\n\n- Added Aztec Release Checklist'
      confirm_message="$confirm_message"$'\n\n- Add Aztec Release Checklist'
    fi

    if [[ -n "$include_incoming_changes" ]]; then
      incoming_change_note=$checklist_message
      if [[ -n "$incoming_change_note" ]]; then
        incoming_change_note="**Update:** $checklist_message"
      fi
      incoming_changes_template=$(sed -e "s/{{incoming_change_note}}/${incoming_change_note}/g" "$script_path/templates/release_checklist_update_incoming.md")
      incoming_changes_template=${incoming_changes_template//$'\n'/\\n}

      issue_body=$(sed -e "s/.*optional_incoming_changes.*/${incoming_changes_template}/" <<< "$issue_body")
      issue_comment="$issue_comment"$'\n\n- Added Incoming Changes Checklist'
      confirm_message="$confirm_message"$'\n\n- Add Incoming Changes Checklist'
    fi

    if [[ -n "$checklist_message" ]]; then
      issue_comment="$issue_comment"$'\n\n'"$checklist_message"
    fi

    if [[ -n "$debug_template" ]]; then
      echo "$issue_body"
      exit 0
    fi

    if [[ -z "$auto_confirm" ]]; then
      echo ""
      confirm_to_proceed "$confirm_message"$'\n\n'
    fi

    gh issue comment $issue_number --body "$issue_comment">/dev/null
    gh issue edit $issue_number --body "$issue_body"
  popd >/dev/null

  exit 0;
fi



if [[ -z "$release_type" ]]; then
  default_release_TYPE="scheduled"

  read -r -p "Please enter release type: scheduled|beta|hotfix [$default_release_type]:" release_type
  release_type=${release_type:-$default_release_type}
  echo ""
fi

if [[ ! "$release_type" == "scheduled" && ! "$release_type" == "beta" && ! "$release_type" == "hotfix" ]]; then
    abort "Error release type must be scheduled|beta|hotfix, you entered '$release_type'"
fi

if [[ -z "$version_number" ]]; then
  pushd "$gb_mobile_path" >/dev/null

  # Ask for new version number
  current_version_number=$(jq '.version' package.json --raw-output)
  version_array=($(awk -F. '{$1=$1} 1' <<<"${current_version_number}"))
  default_version_number="${version_array[0]}.$((version_array[1] + 1)).${version_array[2]}"

  read -r -p "Enter the new version number [$default_version_number]: " version_number
  echo ""
  version_number=${version_number:-$default_version_number}

  if [[ -z "$version_number" ]]; then
      abort "Version number cannot be empty."
  fi

  popd >/dev/null
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

checklist_template_path="$script_path/templates/release_checklist.md"

pushd "$gb_mobile_path" >/dev/null
  milestone_url=$(gh api  --method GET repos/:owner/:repo/milestones --jq ".[0].html_url")
  default_milestone_url="https://wordpress-mobile/gutenber-mobile/milestones"

  milestone_url=${milestone_url:-$default_milestone_url}
popd >/dev/null

checklist_template=$(sed -e "s/{{version_number}}/${version_number}/g" -e "s/{{release_date}}/${release_date}/g" -e "s/{{milestone_url}}/${milestone_url//\//\\/}/g" $checklist_template_path)

if [[ $release_type == "beta" || $release_type == "hotfix" ]]; then
  release_checklist_template=$(sed "/<!-- scheduled_release_only -->/,/<!-- \/scheduled_release_only -->/d" <<< "$checklist_template")
else
  release_checklist_template=$(sed "/<!-- non_scheduled_release_only-->/,/<!-- \/non_scheduled_release_only -->/d" <<< "$checklist_template")
fi

if [[ -n "$include_aztec_steps" ]]; then
  aztec_checklist_template=$(<"$script_path/templates/release_checklist_update_aztec.md")
  aztec_checklist_template=${aztec_checklist_template//$'\n'/\\n}
  release_checklist_template=$(sed -e "s/.*optional_aztec_release.*/$aztec_checklist_template/" <<< "$release_checklist_template")
fi

issue_title="Release checklist for v$version_number ($release_type)"
issue_label="release checklist,$release_type release"
issue_assignee="@me"
issue_body=$'# Release Checklist\n'"This checklist is for the $release_type release v$version_number."$'\n\n **Release date:** '"$release_date"$'\n'"$checklist_message"$'\n'"$release_checklist_template"

if [[ -n "$debug_template" ]]; then
  echo $issue_body
  exit 0;
fi

if [[ -z "$auto_confirm" ]]; then
  echo ""
  confirm_to_proceed "Ready to create '$issue_title' issue ?"
fi

pushd "$gb_mobile_path" >/dev/null

  gh issue create --title "$issue_title" --body "$issue_body" --assignee "$issue_assignee" --label "$issue_label"

popd >/dev/null
