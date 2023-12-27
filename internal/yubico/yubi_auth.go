package yubico

import (
	"bufio"
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"time"

	"github.com/vitalvas/oneauth/internal/tools"
)

var (
	yubiCloudServers = []string{
		"https://api.yubico.com/wsapi/2.0/verify",
		"https://api2.yubico.com/wsapi/2.0/verify",
		"https://api3.yubico.com/wsapi/2.0/verify",
		"https://api4.yubico.com/wsapi/2.0/verify",
		"https://api5.yubico.com/wsapi/2.0/verify",
	}
	serverErrorCodes = []string{
		"BACKEND_ERROR",
		"NOT_ENOUGH_ANSWERS",
		"OPERATION_NOT_ALLOWED",
		"NO_SUCH_CLIENT",
	}
)

type YubiAuth struct {
	clientID     int
	clientSecret []byte

	httpClient *http.Client
}

type VerifyResponse struct {
	Timestamp      int64
	SessionCounter int64
	SessionUse     int64
	Status         string
}

func NewYubiAuth(clientID int, clientSecret string) (*YubiAuth, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(clientSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to decode client secret: %w", err)
	}

	rand.Shuffle(len(yubiCloudServers), func(i, j int) { yubiCloudServers[i], yubiCloudServers[j] = yubiCloudServers[j], yubiCloudServers[i] })

	return &YubiAuth{
		clientID:     clientID,
		clientSecret: keyBytes,

		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}, nil
}

func (y *YubiAuth) Verify(otp string) (*VerifyResponse, error) {
	nonce, err := tools.GenerateNonce(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	params := url.Values{
		"id":        {fmt.Sprintf("%d", y.clientID)},
		"otp":       {otp},
		"nonce":     {nonce},
		"timestamp": {"1"},
		"sl":        {"secure"},
		"timeout":   {"2"},
	}

	if y.clientSecret != nil {
		signRequest(params, y.clientSecret)
	}

	// TODO: add verify response from yubicloud
	return y.getVerify(params)
}

func (y *YubiAuth) getVerify(params url.Values) (*VerifyResponse, error) {
	for _, server := range yubiCloudServers {
		resp, err := y.makeRequest(server, params)
		if err != nil {
			log.Println(err)
		} else if !slices.Contains(serverErrorCodes, resp.Status) {
			return resp, nil
		}
	}

	return nil, fmt.Errorf("failed to make request to all yubico servers")
}

func (y *YubiAuth) makeRequest(server string, params url.Values) (*VerifyResponse, error) {
	remote, err := url.Parse(server)
	if err != nil {
		return nil, fmt.Errorf("failed to parse server URL: %w", err)
	}

	remote.RawQuery = params.Encode()

	req, err := http.NewRequest(http.MethodGet, remote.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; OneAuthServer/1.0; +https://oneauth.vitalvas.dev)")

	resp, err := y.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return y.responseFromBody(body)
}

func (y *YubiAuth) responseFromBody(body []byte) (*VerifyResponse, error) {
	buf := bytes.NewBuffer(body)

	scanner := bufio.NewScanner(buf)

	m := make(map[string]string)

	for scanner.Scan() {
		l := scanner.Bytes()

		s := bytes.SplitN(l, []byte{'='}, 2)

		if len(s) != 2 {
			continue
		}

		m[string(s[0])] = string(s[1])
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan response body: %w", err)
	}

	resp := &VerifyResponse{}

	if v, ok := m["timestamp"]; ok {
		if val, err := strconv.ParseInt(v, 10, 64); err == nil {
			resp.Timestamp = val
		}
	}

	if v, ok := m["sessioncounter"]; ok {
		if val, err := strconv.ParseInt(v, 10, 64); err == nil {
			resp.SessionCounter = val
		}
	}

	if v, ok := m["sessionuse"]; ok {
		if val, err := strconv.ParseInt(v, 10, 64); err == nil {
			resp.SessionUse = val
		}
	}

	if v, ok := m["status"]; ok {
		resp.Status = v
	}

	return resp, nil
}

func signRequest(params url.Values, secret []byte) {
	h := hmac.New(sha1.New, secret)

	u := params.Encode()
	h.Write([]byte(u))
	sig := h.Sum(nil)

	params.Set("h", base64.StdEncoding.EncodeToString(sig))
}
