Release for Gutenberg Mobile {{ .Version }}

<!-- ## Related PRs
{{ range .RelatedPRs }}
- {{ .Url }}{{ end }}

## Changes

{{ range .Changes }}
### {{ .Title }}
* PR {{ .PrUrl }}
{{ range $i, $issue := .Issues }}
* Issue {{ $i }}{{ $issue }}{{ end }}
{{ end }} -->


<!-- To determine the changes you can check the RELEASE-NOTES.txt and gutenberg/packages/react-native-editor/CHANGELOG.md files and cross check with the list of commits that are part of the PR -->
## Test plan

Once the installable builds of the main apps are ready, perform a quick smoke test of the editor on both iOS and Android to verify it launches without crashing. We will perform additional testing after the main apps cut their releases.

## Release Submission Checklist

- [ ] Verify Items from test plan have been completed
- [ ] Check if `RELEASE-NOTES.txt` is updated with all the changes that made it to the release. Replace `Unreleased` section with the release version and create a new `Unreleased` section.
- [ ] Check if `gutenberg/packages/react-native-editor/CHANGELOG.md` is updated with all the changes that made it to the release. Replace `## Unreleased` with the release version and create a new `## Unreleased`.
- [ ] Bundle package of the release is updated.