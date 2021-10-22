Release for Gutenberg Mobile v1.XX.Y

## Related PRs

- Gutenberg: https://github.com/WordPress/gutenberg/pull/
- WPAndroid: https://github.com/wordpress-mobile/WordPress-Android/pull/
- WPiOS: https://github.com/wordpress-mobile/WordPress-iOS/pull/

- Aztec-iOS: https://github.com/wordpress-mobile/AztecEditor-iOS/pull/
- Aztec-Android: https://github.com/wordpress-mobile/AztecEditor-Android/pull

## Extra PRs that Landed After the Release Was Cut

No extra PRs yet. 🎉

## Changes

<!-- To determine the changes you can check the RELEASE-NOTES.txt and gutenberg/packages/react-native-editor/CHANGELOG.md files and cross check with the list of commits that are part of the PR -->

### Change 1
- **PR:** link-to-pr-related-to-change-1
- **Issue:** link-to-issue-related-to-change-1

### Change 2
- **PR:** link-to-pr-related-to-change-2
- **Issue:** link-to-issue-related-to-change-2

## Test plan

- Use the main WP apps to test the changes above.
- Smoke test the main WP apps for [general writing flow](https://github.com/wordpress-mobile/test-cases/tree/master/test-cases/gutenberg/writing-flow).
- Test the Unsupported Block Editor on WP Apps ([see steps](https://github.com/wordpress-mobile/test-cases/blob/trunk/test-cases/gutenberg/unsupported-block-editing.md#unsupported-block-editing---test-cases)).
- Sanity [test suites](https://github.com/wordpress-mobile/test-cases/blob/trunk/test-suites/gutenberg/sanity-test-suites.md) for WP Apps should be completed for each platform.

## Release Submission Checklist

- [ ] Verify Items from test plan have been completed
- [ ] Check if `RELEASE-NOTES.txt` is updated with all the changes that made it to the release. Replace `Unreleased` section with the release version and create a new `Unreleased` section.
- [ ] Check if `gutenberg/packages/react-native-editor/CHANGELOG.md` is updated with all the changes that made it to the release. Replace `## Unreleased` with the release version and create a new `## Unreleased`.
- [ ] Bundle package of the release is updated.
