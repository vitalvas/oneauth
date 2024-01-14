package updates

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/Masterminds/semver/v3"
)

type UpdateManifest struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	RemotePrefix string `json:"remote_prefix"`
}

var (
	ErrSchemeNotHTTPS    = errors.New("scheme is not HTTPS")
	ErrNoUpdateAvailable = errors.New("no update available")
)

func getUpdateManifestURL(appName string, channel Channel) (string, error) {
	parsedBaseURL, err := url.Parse(channel.String())
	if err != nil {
		return "", err
	}

	if parsedBaseURL.Scheme != "https" {
		return "", ErrSchemeNotHTTPS
	}

	endpointURL, err := url.Parse(fmt.Sprintf("%s_update_manifest.json", appName))
	if err != nil {
		return "", err
	}

	fullURL := parsedBaseURL.ResolveReference(endpointURL)

	return fullURL.String(), nil
}

func getRemoteManifest(appName, remote string) (*UpdateManifest, error) {
	manifest := &UpdateManifest{}

	if err := getJSON(appName, remote, &manifest); err != nil {
		return nil, err
	}

	return manifest, nil
}

func Check(appName, version string) (*UpdateManifest, error) {
	channel := getChannel(version)

	localVersion, err := checkVersion(version)
	if err != nil {
		return nil, err
	}

	manifestURL, err := getUpdateManifestURL(appName, channel)
	if err != nil {
		return nil, err
	}

	manifest, err := getRemoteManifest(appName, manifestURL)
	if err != nil {
		return nil, err
	}

	if _, err := checkVersion(manifest.Version); err != nil {
		return nil, fmt.Errorf("invalid version in manifest: %w", err)
	}

	remoteVersion, err := semver.NewVersion(manifest.Version)
	if err != nil {
		return nil, err
	}

	if !localVersion.Check(remoteVersion) {
		return nil, ErrNoUpdateAvailable
	}

	return manifest, nil
}
