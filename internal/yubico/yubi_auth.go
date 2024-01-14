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

const (
	cloudUserAgent = "Mozilla/5.0 (compatible; OneAuthServer/1.0; +https://oneauth.vitalvas.dev)"
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
		StatusBackendError,
		StatusNotEnoughAnswers,
		StatusOperationNotAllowed,
		StatusNoSuchClient,
	}
	httpClient = &http.Client{
		Timeout: 3 * time.Second,
	}
)

type YubiAuth struct {
	clientID     int
	clientSecret []byte
}

type VerifyResponse struct {
	Serial         int64
	Timestamp      int64
	SessionCounter int64
	SessionUse     int64
	Status         string
}

func getVerifyServers(locals ...string) []string {
	servers := yubiCloudServers

	if len(locals) > 0 {
		servers = locals
	}

	// allways shuffle servers for load balancing
	rand.Shuffle(len(servers), func(i, j int) {
		servers[i], servers[j] = servers[j], servers[i]
	})

	return servers
}

func NewYubiAuth(clientID int, clientSecret string) (*YubiAuth, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(clientSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to decode client secret: %w", err)
	}

	return &YubiAuth{
		clientID:     clientID,
		clientSecret: keyBytes,
	}, nil
}

func (y *YubiAuth) Verify(otp string) (*VerifyResponse, error) {
	serial, err := ValidateOTP(otp)
	if err != nil {
		return nil, fmt.Errorf("failed to validate otp: %w", err)
	}

	params, err := getRequestParams(y.clientID, otp)
	if err != nil {
		return nil, fmt.Errorf("failed to get request params: %w", err)
	}

	if y.clientSecret != nil {
		signRequest(params, y.clientSecret)
	}

	// TODO: add verify response from yubicloud
	resp, err := y.getVerify(params)
	if err != nil {
		return nil, err
	}

	resp.Serial = serial

	return resp, nil
}

func (y *YubiAuth) getVerify(params url.Values) (*VerifyResponse, error) {
	for _, server := range getVerifyServers() {
		resp, err := makeRequest(server, params)
		if err != nil {
			log.Println(err)
		} else if !slices.Contains(serverErrorCodes, resp.Status) {
			return resp, nil
		}
	}

	return nil, fmt.Errorf("failed to make request to all yubico servers")
}

func makeRequest(server string, params url.Values) (*VerifyResponse, error) {
	remote, err := url.Parse(server)
	if err != nil {
		return nil, fmt.Errorf("failed to parse server URL: %w", err)
	}

	remote.RawQuery = params.Encode()

	req, err := http.NewRequest(http.MethodGet, remote.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", cloudUserAgent)

	resp, err := httpClient.Do(req)
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

	return responseFromBody(body)
}

func getRequestParams(clientID int, otp string) (url.Values, error) {
	nonce, err := tools.GenerateNonce(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	return url.Values{
		"id":        {fmt.Sprintf("%d", clientID)},
		"otp":       {otp},
		"nonce":     {nonce},
		"timestamp": {"1"},
		"sl":        {"secure"},
		"timeout":   {"2"},
	}, nil
}

func responseFromBody(body []byte) (*VerifyResponse, error) {
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
