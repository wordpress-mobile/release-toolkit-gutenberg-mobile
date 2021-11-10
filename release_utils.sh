#!/bin/bash

# Utils adapted from https://github.com/Homebrew/install/blob/master/install.sh
# string formatters
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


# Takes multiple arguments and prints them joined with single spaces in-between
# while escaping any spaces in arguments themselves
shell_join() {
    local arg
    printf "%s" "$1"
    shift
    for arg in "$@"; do
        printf " "
        printf "%s" "${arg// /\ }"
    done
}

# Takes multiple arguments, joins them and prints them in a colored format
ohai() {
  printf "${tty_blue}==> %s${tty_reset}\n" "$(shell_join "$@")"
}

# Takes a single argument and prints it in a colored format
warn() {
  printf "${tty_underline}${tty_red}Warning${tty_reset}: %s\n" "$1"
}

# Takes a single argument, prints it in a colored format and aborts the script
abort() {
    printf "\n${tty_red}%s${tty_reset}\n" "$1"
    exit 1
}

# Takes multiple arguments consisting a command and executes it. If the command
# is not successful aborts the script, printing the failed command and its
# arguments in a colored format.
#
# Returns the executed command's result if it's successful.
execute() {
    if ! "$@"; then
        abort "$(printf "Failed during: %s" "$(shell_join "$@")")"
    fi
}


#####
# Confirm to Proceed Prompt
#####

# Accepts a single argument: a yes/no question (ending with a ? most likely) to ask the user
function confirm_to_proceed() {
    read -r -p "$1 (y/n) " -n 1
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        abort "Aborting..."
    fi
}

GB_MOBILE_PATH=""

verify_gb_mobile_path() {
     # Verify the path is valid
    gb_package_name=$(jq '.name' "$1/package.json" --raw-output)
    if [[ "$gb_package_name" != "gutenberg-mobile" ]]; then
        GB_MOBILE_PATH=""
        abort "Could not find a valid gutenberg-mobile package.json at $1"
    fi
}

set_gb_mobile_path() {

    if [[ -n "${1-}" ]]; then
        verify_gb_mobile_path "$1"

        GB_MOBILE_PATH="$1"
        return
    fi

    if [[ -n "$GB_MOBILE_PATH" ]]; then
        return
    fi


    # Ask for path to gutenberg-mobile directory
    default_gb_mobile_location="$(pwd)/../gutenberg-mobile"
    read -r -p "Please enter the path to the gutenberg-mobile directory [$default_gb_mobile_location]:" gb_mobile_path
    gb_mobile_path=${gb_mobile_path:-"$default_gb_mobile_location"}
    echo ""

    verify_gb_mobile_path "$gb_mobile_path"

    GB_MOBILE_PATH="$gb_mobile_path"
}

pushd_gb_mobile() {
    set_gb_mobile_path

    if [[ -z "$GB_MOBILE_PATH" ]]; then
        abort "Unable to move to gutenberg mobile path"
    fi
    pushd $GB_MOBILE_PATH > /dev/null
}

popd_gb_mobile() {
    popd > /dev/null
}