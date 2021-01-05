# release-toolkit-gutenberg-mobile
Automation Scripts for Releasing Gutenberg-Mobile Updates to the WordPress Mobile Apps.

## Prerequisites

To be able to run the automation script make sure you have installed:

- [Github CLI](https://github.com/cli/cli)
```sh
brew install gh
```
- [jq](https://github.com/stedolan/jq)
```sh
brew install jq
```

## Usage

Prerequisite: Use Xcode 12.1 (not Xcode 12.2). Set it using `sudo xcode-select -s /Applications/Xcode12.1.app` (changing the path if required). This avoids a "Failed to build gem native extension." error during the `rake dependencies` step.

Run the script: `./release_automation.sh`

## Testing

You can test the scripts on forked repos. Please follow the instructions on top of the [release_automation.sh](./release_automation.sh) file.
