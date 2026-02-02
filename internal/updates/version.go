package updates

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/Masterminds/semver/v3"
)

var (
	ErrInvalidVersion = errors.New("invalid version")
	versionRegex      = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)
)

func checkVersion(version string) (*semver.Constraints, error) {
	if !versionRegex.MatchString(version) {
		return nil, ErrInvalidVersion
	}

	return semver.NewConstraint(fmt.Sprintf(">=%s", version))
}

func CheckNewVersion(currentVersion, newVersion string) (bool, error) {
	if _, err := checkVersion(currentVersion); err != nil {
		return false, fmt.Errorf("failed to check current version: %w", err)
	}

	if _, err := checkVersion(newVersion); err != nil {
		return false, fmt.Errorf("failed to check new version: %w", err)
	}

	c, err := semver.NewConstraint(fmt.Sprintf(">%s", currentVersion))
	if err != nil {
		return false, fmt.Errorf("failed to create constraint: %w", err)
	}

	v, err := semver.NewVersion(newVersion)
	if err != nil {
		return false, fmt.Errorf("failed to create version: %w", err)
	}

	return c.Check(v), nil
}
