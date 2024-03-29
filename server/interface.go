package server

//go:generate mockgen -destination ../mocks/api.go -package mocks golift.io/securityspy/server API

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"time"
)

// API interface is provided only to allow overriding local methods during local testing.
// The methods in this interface connect to SecuritySpy so they become
// blockers when testing without a SecuritySpy server available. Overriding
// them with fakes makes testing (for most methods in this library) possible.
// One day this will be ripped out and you can test with a mock http server.
type API interface { //nolint:interfacebloat
	Get(apiPath string, params url.Values) (resp *http.Response, err error)
	GetContext(ctx context.Context, apiPath string, params url.Values) (resp *http.Response, err error)
	GetContextClient(context.Context, string, url.Values, *http.Client) (resp *http.Response, err error)
	GetClient(string, url.Values, *http.Client) (resp *http.Response, err error)
	Post(apiPath string, params url.Values, post io.ReadCloser) (body []byte, err error)
	GetXML(apiPath string, params url.Values, v interface{}) (err error)
	SimpleReq(apiURI string, params url.Values, cameraNum int) error
	TimeoutDur() time.Duration
	BaseURL() string
	Auth() string
	HTTPClient() *http.Client
}

var _ = API(&Config{})
