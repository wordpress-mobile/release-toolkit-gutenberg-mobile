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

### Release Script
- Run the script: `./release_automation.sh`
- See [Releasing.md](./Releasing.md) for more details

### WordPress Gutenberg Reference Update Automation
- Run the script: `./wp_gutenberg_ref_update_prs.sh`
- See [WordPressApps_Gutenberg_Reference_Update.md](./WordPressApps_Gutenberg_Reference_Update.md) for more details

## Testing

You can test the scripts on forked repos. Please follow the instructions on top of the [release_automation.sh](./release_automation.sh) file.
