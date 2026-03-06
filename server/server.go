package server

import (
	"context"
	"crypto/tls"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ErrCmdNotOK is returned for any command that has a successful web request,
// but the reply does not end with the word OK.
var ErrCmdNotOK = errors.New("command unsuccessful")

// DefaultTimeout it used for almost every request to SecuritySpy. Adjust as needed.
const DefaultTimeout = 10 * time.Second

// Config is the input data for this library. Only set VerifySSL to true if your server
// has a valid SSL certificate. The password is auto-repalced with a base64 encoded string.
type Config struct {
	URL       string
	Password  string //nolint:gosec // User provided.
	Username  string
	Client    *http.Client // Provide an HTTP client, or:
	Timeout   Duration     // Only used if you do not provide an HTTP client.
	VerifySSL bool         // Also only used if you do not provide an HTTP client.
}

// HTTPClient returns an http.Client with the configured timeout and SSL verification.
func (s *Config) HTTPClient() *http.Client {
	return &http.Client{
		Timeout: s.TimeoutDur(),
		Transport: &http.Transport{
			DisableKeepAlives: true, // SecuritySpy has a Keep-Alive Bug.
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: !s.VerifySSL, //nolint:gosec // User selected.
			},
		},
	}
}

// Duration allows you to pass the server Config struct in from a json file.
type Duration struct {
	time.Duration
}

// UnmarshalText parses a duration type from a config file. This method works
// with the Duration type to allow unmarshaling of durations from files and
// env variables in the same struct. You won't generally call this directly.
func (d *Duration) UnmarshalText(b []byte) error {
	var err error

	d.Duration, err = time.ParseDuration(string(b))
	if err != nil {
		return fmt.Errorf("parsing Go duration '%s': %w", b, err)
	}

	return nil
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
	if s.Timeout.Duration <= 0 {
		return DefaultTimeout
	}

	return s.Timeout.Duration
}

// GetContextClient is the same as Get except you can pass in your own context and http Client.
func (s *Config) GetContextClient( //nolint:cyclop // might make it less complicated later.
	ctx context.Context,
	api string,
	params url.Values,
	client *http.Client,
) (*http.Response, error) {
	if params == nil {
		params = make(url.Values)
	}

	if s.Password != "" {
		params.Set("auth", s.Password)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL+api, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest(): %w", err)
	}

	if !strings.HasPrefix(api, "++getfile") && !strings.HasPrefix(api, "++event") &&
		!strings.HasPrefix(api, "++image") && !strings.HasPrefix(api, "++audio") &&
		!strings.HasPrefix(api, "++stream") && !strings.HasPrefix(api, "++video") {
		params.Set("format", "xml")
		req.Header.Add("Accept", "application/xml")
	}

	req.URL.RawQuery = params.Encode()

	resp, err := client.Do(req) //nolint:gosec // the taint comes from the operator.
	if err != nil {
		return resp, fmt.Errorf("http request: %w", err)
	}

	return resp, nil
}

// GetContext is the same as Get except you can pass in your own context.
func (s *Config) GetContext(ctx context.Context, apiPath string, params url.Values) (*http.Response, error) {
	if s.Client == nil {
		s.Client = s.HTTPClient()
	}

	return s.GetContextClient(ctx, apiPath, params, s.Client)
}

// GetClient is the same as Get except you can pass in your own http Client.
func (s *Config) GetClient(apiPath string, params url.Values, client *http.Client) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutDur())
	defer cancel()

	return s.GetContextClient(ctx, apiPath, params, client)
}

// Get is a helper function that formats the http request to SecuritySpy.
func (s *Config) Get(apiPath string, params url.Values) (*http.Response, error) {
	if s.Client == nil {
		s.Client = s.HTTPClient()
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutDur())
	defer cancel()

	return s.GetContextClient(ctx, apiPath, params, s.Client)
}

// Post is a helper function that formats the http request to SecuritySpy.
func (s *Config) Post(apiPath string, params url.Values, body io.ReadCloser) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutDur())
	defer cancel()

	return s.PostContext(ctx, apiPath, params, body)
}

// PostContext is a helper function that formats the http request to SecuritySpy.
func (s *Config) PostContext(
	ctx context.Context, apiPath string, params url.Values, body io.ReadCloser,
) ([]byte, error) {
	if params == nil {
		params = make(url.Values)
	}

	if s.Password != "" {
		params.Set("auth", s.Password)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.URL+apiPath, body)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest(): %w", err)
	}

	if strings.HasPrefix(apiPath, "++audio") {
		req.Header.Add("Content-Type", "audio/g711-ulaw")
	}

	req.URL.RawQuery = params.Encode()

	resp, err := s.Client.Do(req) //nolint:gosec // the taint comes from the operator.
	if err != nil {
		return nil, fmt.Errorf("getting body: %w", err)
	}
	defer resp.Body.Close()

	reply, err := io.ReadAll(resp.Body)
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
func (s *Config) GetXML(apiPath string, params url.Values, val any) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutDur())
	defer cancel()

	return s.GetXMLContext(ctx, apiPath, params, val)
}

// GetXMLContext returns raw http body, so it can be unmarshaled into an xml struct.
func (s *Config) GetXMLContext(ctx context.Context, apiPath string, params url.Values, val any) error {
	resp, err := s.GetContext(ctx, apiPath, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		return fmt.Errorf("request failed (%v): %v (status: %v/%v): %w: %s",
			s.Username, s.URL+apiPath, resp.StatusCode, resp.Status, err, string(body))
	}

	if err = xml.NewDecoder(resp.Body).Decode(val); err != nil {
		return fmt.Errorf("reading body: %w", err)
	}

	return nil
}

// SimpleReq performes HTTP req, checks for OK at end of output.
func (s *Config) SimpleReq(apiURI string, params url.Values, cameraNum int) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutDur())
	defer cancel()

	return s.SimpleReqContext(ctx, apiURI, params, cameraNum)
}

// SimpleReqContext performes HTTP req, checks for OK at end of output.
func (s *Config) SimpleReqContext(ctx context.Context, apiURI string, params url.Values, cameraNum int) error {
	if cameraNum >= 0 {
		params.Set("cameraNum", strconv.Itoa(cameraNum))
	}

	resp, err := s.GetContext(ctx, apiURI, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil || !strings.HasSuffix(string(body), "OK") {
		return ErrCmdNotOK
	}

	return nil
}
