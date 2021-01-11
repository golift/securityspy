package server

import (
	"context"
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ErrorCmdNotOK is returned for any command that has a successful web request,
// but the reply does not end with the word OK.
var ErrorCmdNotOK = fmt.Errorf("command unsuccessful")

// DefaultTimeout it used for almost every request to SecuritySpy. Adjust as needed.
const DefaultTimeout = 10 * time.Second

// Config is the input data for this library. Only set VerifySSL to true if your server
// has a valid SSL certificate. The password is auto-repalced with a base64 encoded string.
type Config struct {
	URL       string
	Password  string
	Username  string
	Client    *http.Client  // Provide an HTTP client, or:
	Timeout   time.Duration // Only used if you do not provide an HTTP client.
	VerifySSL bool          // Also only used if you do not provide an HTTP client.
}

func (s *Config) getClient() {
	if s.Client != nil {
		return
	}

	if s.Timeout == 0 {
		s.Timeout = DefaultTimeout
	}

	s.Client = &http.Client{
		Timeout: s.Timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: !s.VerifySSL}, //nolint:gosec
		},
	}
}

// BaseURL returns the URL.
func (s *Config) BaseURL() string {
	return s.URL
}

// Auth returns the base64'd auth parameter.
func (s *Config) Auth() string {
	return s.Password
}

// TimeoutDur returns the configured timeout.
func (s *Config) TimeoutDur() time.Duration {
	return s.Timeout
}

func (s *Config) GetContext(ctx context.Context, apiPath string, params url.Values) (*http.Response, error) {
	s.getClient()

	if params == nil {
		params = make(url.Values)
	}

	if s.Password != "" {
		params.Set("auth", s.Password)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL+apiPath, nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest(): %w", err)
	}

	if a := apiPath; !strings.HasPrefix(a, "++getfile") && !strings.HasPrefix(a, "++event") &&
		!strings.HasPrefix(a, "++image") && !strings.HasPrefix(a, "++audio") &&
		!strings.HasPrefix(a, "++stream") && !strings.HasPrefix(a, "++video") {
		params.Set("format", "xml")
		req.Header.Add("Accept", "application/xml")
	}

	req.URL.RawQuery = params.Encode()

	return s.Client.Do(req)
}

// Get is a helper function that formats the http request to SecuritySpy.
func (s *Config) Get(apiPath string, params url.Values) (*http.Response, error) {
	return s.GetContext(context.TODO(), apiPath, params)
}

// Post is a helper function that formats the http request to SecuritySpy.
func (s *Config) Post(apiPath string, params url.Values, body io.ReadCloser) ([]byte, error) {
	s.getClient()

	if params == nil {
		params = make(url.Values)
	}

	if s.Password != "" {
		params.Set("auth", s.Password)
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.URL+apiPath, body)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest(): %w", err)
	}

	if a := apiPath; !strings.HasPrefix(a, "++audio") {
		req.Header.Add("Content-Type", "audio/g711-ulaw")
	}

	req.URL.RawQuery = params.Encode()

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("getting body: %w", err)
	}
	defer resp.Body.Close()

	reply, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return reply, fmt.Errorf("request failed (%v): %v (status: %v/%v): %w",
			s.Username, s.URL+apiPath, resp.StatusCode, resp.Status, err)
	}

	return reply, nil
}

// GetXML returns raw http body, so it can be unmarshaled into an xml struct.
func (s *Config) GetXML(apiPath string, params url.Values, v interface{}) error {
	resp, err := s.Get(apiPath, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)

		return fmt.Errorf("request failed (%v): %v (status: %v/%v): %w: %s",
			s.Username, s.URL+apiPath, resp.StatusCode, resp.Status, err, string(body))
	}

	if err = xml.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("reading body: %w", err)
	}

	return nil
}

// SimpleReq performes HTTP req, checks for OK at end of output.
func (s *Config) SimpleReq(apiURI string, params url.Values, cameraNum int) error {
	if cameraNum >= 0 {
		params.Set("cameraNum", strconv.Itoa(cameraNum))
	}

	resp, err := s.Get(apiURI, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil || !strings.HasSuffix(string(body), "OK") {
		return ErrorCmdNotOK
	}

	return nil
}
