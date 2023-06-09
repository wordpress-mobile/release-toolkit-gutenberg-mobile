<!-- Learn how to maintain this file at https://github.com/WordPress/gutenberg/tree/HEAD/packages#maintaining-changelogs. -->

<!--
For each user feature we should also add a importance categorization label  to indicate the relevance of the change for end users of GB Mobile. The format is the following:
[***] → Major new features, significant updates to core flows, or impactful fixes (e.g. a crash that impacts a lot of users) — things our users should be aware of.

[**] → Changes our users will probably notice, but doesn’t impact core flows. Most fixes.

[*] → Minor enhancements and fixes that address annoyances — things our users can miss.
-->

## Unreleased
-   [*] [internal] Upgrade compile and target sdk version to Android API 33 [#50731]

## 1.96.0
-   [**] Tapping on all nested blocks gets focus directly instead of having to tap multiple times depending on the nesting levels. [#50672]
-   [**] Fix undo/redo history when inserting a link configured to open in a new tab [#50460]
-   [*] [List block] Fix an issue when merging a list item into a Paragraph would remove its nested list items. [#50701]

## 1.95.0
-   [**] Fix Android-only issue related to block toolbar not being displayed on some blocks in UBE [#51131]