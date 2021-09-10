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

Run the script: `./release_automation.sh`

## Testing

You can test the scripts on forked repos. Please follow the instructions on top of the [release_automation.sh](./release_automation.sh) file.

## Troubleshooting

Occasionally, the script may encounter an error while running. Because the project makes use of a `node_modules` directory within the Gutenberg submodule, it may be useful in some cases to manually install npm dependencies there. By default, running `npm install` (or `npm i`) from within the `gutenberg-mobile` directory will trigger `npm ci` to run within the `gutenberg` directory, but sometimes it may also be helpful to try:

```
cd ./gutenberg
npm i
```

It can also be useful to run `cd ./gutenberg && npm run distclean`
