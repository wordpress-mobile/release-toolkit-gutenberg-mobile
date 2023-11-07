# Testing

If developing with VSCode, tests can be run inline within the test files themselves, and is the recommended method for running and debugging tests. Tests can also be run from the command line with the `go test` command.

## Testing Environment for Development
The CLI tool can be run against forked repos for testing. To configure your forked repos:

 1. Fork the following repos to your github user repo:
    - [Gutenberg-Mobile](https://github.com/wordpress-mobile/gutenberg-mobile)    
    - [Gutenberg](https://github.com/WordPress/gutenberg)
    - [WordPress-Android](https://github.com/wordpress-mobile/WordPress-Android)
    - [WordPress-iOS](https://github.com/wordpress-mobile/WordPress-iOS)

3. Ensure that your forked repos contains the PR labels specified below:
    a) Gutenberg Mobile: "release-process"
    b) Gutenberg: "Mobile App - i.e. Android or iOS"

4. Ensure that each of your repos contains the target branch `trunk`.

5. Ensure that [.gitmodules](https://github.com/wordpress-mobile/gutenberg-mobile/blob/trunk/.gitmodules) references your Gutenberg fork.


To run commands against the forked repos, set `GBM_WORDPRESS_ORG` and `GBM_WPMOBILE_ORG` as environment variables to user GitHub username. By default, these values with be WordPress and WordPress-Mobile, respectively.

Example command:

```
GBM_WPMOBILE_ORG=yourusername GBM_WORDPRESS_ORG=yourusername go run main.go release prepare gb 1.109.0 
```

