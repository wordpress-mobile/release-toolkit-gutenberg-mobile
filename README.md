# release-toolkit-gutenberg-mobile
Automation Scripts for Releasing Gutenberg-Mobile Updates to the WordPress Mobile Apps.

## GBM-CLI

The preferred automation flow now uses the `gbm-cli` tool. See the tool [README](./gbm-cli/README.md) for more information. Follow the [Installing](./gbm-cli/README.md#installing) section to get started.

### Prerequisites

There are no prerequisites to using the `gbm-cli` tool beyond what is used to develop on Gutenberg Mobile.

It does use the same Github authentication as `gh` so it is recommended to run `gh auth` before using the tool. See [Authentication](./gbm-cli/README.md#authentication) for alternatives.

#### NVM

`nvm` is the recommended node manager for running the `gbm-cli` release commands.

When preparing Gutenberg for a release it is possible to set the global node version to the current version required by Gutenberg.

At the moment this is not possible when preparing a Gutenberg Mobile PR locally since there are multiple node versions during the preparation.

## Legacy Automation Script

### Prerequisites

To be able to run the legacy automation script make sure you have installed:

- [Github CLI](https://github.com/cli/cli)
```sh
brew install gh
```
- [jq](https://github.com/stedolan/jq)
```sh
brew install jq
```

### Usage

Run the script: `./release_automation.sh`

### Testing

You can test the scripts on forked repos. Please follow the instructions on top of the [release_automation.sh](./release_automation.sh) file.

### Troubleshooting

Occasionally, the script may encounter an error while running. Because the project makes use of a `node_modules` directory within the Gutenberg submodule, it may be useful in some cases to manually install npm dependencies there. By default, running `npm install` (or `npm i`) from within the `gutenberg-mobile` directory will trigger `npm ci` to run within the `gutenberg` directory, but sometimes it may also be helpful to try:

```
cd ./gutenberg
npm i
```

It can also be useful to run `cd ./gutenberg && npm run distclean`
