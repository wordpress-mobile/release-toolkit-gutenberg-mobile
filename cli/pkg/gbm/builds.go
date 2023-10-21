package gbm

import (
	"github.com/wordpress-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/gbm-cli/pkg/gh"
)

func AndroidGbmBuildPublished(version string) (bool, error) {
	// Forcing the search to look in the wordpress-mobile org
	// Since forks will not have the status checks
	pr, err := FindGbmReleasePr(version)
	console.Info("Checking status for %s %s", pr.Title, pr.Url)
	if err != nil {
		return false, err
	}
	sha := pr.Head.Sha
	status, err := gh.GetStatusCheck("gutenberg-mobile", sha, "build-android-rn-bridge-and-publish-to-s3")
	if err != nil {
		return false, err
	}
	return status.State == "success", nil
}

func IosGbmBuildPublished(version string) (bool, error) {
	// Forcing the search to look in the wordpress-mobile org
	// Since forks will not have the status checks
	pr, err := FindGbmReleasePr(version)
	console.Info("Checking status for %s %s", pr.Title, pr.Url)
	if err != nil {
		return false, err
	}
	sha := pr.Head.Sha
	status, err := gh.GetStatusCheck("gutenberg-mobile", sha, "build-ios-rn-xcframework-and-publish-to-s3")
	if err != nil {
		return false, err
	}
	return status.State == "success", nil

}
