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

	ErrUpdateNotFound  = fmt.Errorf("update not found")
	ErrUpdateForbidden = fmt.Errorf("update forbidden")
)

func getUserAget(appName string) string {
	return fmt.Sprintf(
		"Mozilla/5.0 (compatible; %s/%s; os/%s; arch/%s)",
		appName, buildinfo.Version, buildinfo.OS, buildinfo.ARCH,
	)
}

func getJSON(appName, remote string, v interface{}) error {
	req, err := http.NewRequest(http.MethodGet, remote, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", getUserAget(appName))

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return json.NewDecoder(resp.Body).Decode(&v)

	case http.StatusNotFound:
		return ErrUpdateNotFound

	case http.StatusForbidden:
		return ErrUpdateForbidden

	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}
