package updates

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/vitalvas/oneauth/internal/buildinfo"
)

var (
	httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}
)

func getJSON(appName, remote string, v interface{}) error {
	req, err := http.NewRequest(http.MethodGet, remote, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", fmt.Sprintf(
		"Mozilla/5.0 (compatible; %s/%s; os/%s; arch/%s)",
		appName, buildinfo.Version, buildinfo.OS, buildinfo.ARCH,
	))

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(&v)
}
