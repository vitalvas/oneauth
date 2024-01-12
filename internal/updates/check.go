package updates

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/vitalvas/oneauth/internal/buildinfo"
)

type UpdateManifest struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	RemotePrefix string `json:"remote_prefix"`
}

var (
	ErrSchemeNotHTTPS    = errors.New("scheme is not HTTPS")
	ErrNoUpdateAvailable = errors.New("no update available")

	httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}
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
	req, err := http.NewRequest(http.MethodGet, remote, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", fmt.Sprintf(
		"Mozilla/5.0 (compatible; %s/%s; os/%s; arch/%s)",
		appName, buildinfo.Version, runtime.GOOS, runtime.GOARCH,
	))

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data := &UpdateManifest{}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}

func Check(appName string, version string) (*UpdateManifest, error) {
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
