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
