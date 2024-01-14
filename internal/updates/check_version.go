package updates

import (
	"fmt"
	"net/url"

	"github.com/vitalvas/oneauth/internal/buildinfo"
)

type UpdateVersionManifest struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Sha256  string `json:"sha256"`
}

func getUpdateVersionManifestURL(appName, remotePrefix string) (string, error) {
	parsedBaseURL, err := url.Parse(remotePrefix)
	if err != nil {
		return "", err
	}

	if parsedBaseURL.Scheme != "https" {
		return "", ErrSchemeNotHTTPS
	}

	endpointURL, err := url.Parse(fmt.Sprintf("%s_%s_%s_manifest.json", appName, buildinfo.OS, buildinfo.ARCH))
	if err != nil {
		return "", err
	}

	fullURL := parsedBaseURL.ResolveReference(endpointURL)

	return fullURL.String(), nil
}

func getRemoteVersionManifest(appName, remote string) (*UpdateVersionManifest, error) {
	manifest := &UpdateVersionManifest{}

	err := getJSON(appName, remote, &manifest)
	if err != nil {
		return nil, err
	}

	return manifest, nil
}

func CheckVersion(appName, remotePrefix string) (*UpdateVersionManifest, error) {
	manifestURL, err := getUpdateVersionManifestURL(appName, remotePrefix)
	if err != nil {
		return nil, err
	}

	manifest, err := getRemoteVersionManifest(appName, manifestURL)
	if err != nil {
		return nil, err
	}

	return manifest, nil
}
