package server_test

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golift.io/securityspy/v2/server"
)

func TestGet(t *testing.T) {
	t.Parallel()

	asert := assert.New(t)
	config := &server.Config{
		Username:  "user",
		Password:  "pass",
		URL:       "http://some.host:5678/",
		VerifySSL: false,
		Timeout:   server.Duration{time.Second},
	}
	handler := http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		asert.Equal("xml", req.URL.Query().Get("format"), "format parameter was not added")
		asert.Equal(config.Password, req.URL.Query().Get("auth"), "auth parameter was not added")
		asert.Equal("application/xml", req.Header.Get("Accept"), "accept header is not correct")

		_, err := resp.Write([]byte("request OK"))
		asert.NoError(err, "the fake server must return an error writing to the client")
		asert.Equal(config.URL, "http://"+req.Host+"/", "the host was not set correctly in the request")
	})

	httpClient, fakeServer := testingHTTPClient(handler)
	defer fakeServer.Close()

	config.Client = httpClient
	resp, err := config.Get("++path", make(url.Values))
	require.NoError(t, err, "the method must not return an error when given a valid server to query")

	if err == nil {
		defer resp.Body.Close()

		asert.Equal(http.StatusOK, resp.StatusCode, "the server must return a 200 response code")
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err, "must not be an error reading the response body")
		asert.Equal("request OK", string(body), "wrong data was returned from the server")
	}
}

func TestPostAudioContentType(t *testing.T) {
	t.Parallel()

	asert := assert.New(t)
	config := &server.Config{
		Username:  "user",
		Password:  "pass",
		URL:       "http://some.host:5678/",
		VerifySSL: false,
		Timeout:   server.Duration{time.Second},
	}

	var (
		audioHeader string
		otherHeader string
	)

	handler := http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/++audio":
			audioHeader = req.Header.Get("Content-Type")
		case "/++other":
			otherHeader = req.Header.Get("Content-Type")
		}

		_, err := resp.Write([]byte("OK"))
		asert.NoError(err)
	})

	httpClient, fakeServer := testingHTTPClient(handler)
	defer fakeServer.Close()

	config.Client = httpClient

	body, err := config.Post("++audio", nil, io.NopCloser(strings.NewReader("audio-data")))
	require.NoError(t, err)
	asert.Equal("OK", string(body))

	body, err = config.Post("++other", nil, io.NopCloser(strings.NewReader("other-data")))
	require.NoError(t, err)
	asert.Equal("OK", string(body))

	asert.Equal("audio/g711-ulaw", audioHeader)
	asert.Empty(otherHeader)
}

func TestGetXMLStatusAndDecodeErrors(t *testing.T) {
	t.Parallel()

	config := &server.Config{
		URL:       "http://some.host:5678/",
		VerifySSL: false,
		Timeout:   server.Duration{time.Second},
	}

	handler := http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/++bad-status":
			http.Error(resp, "forbidden", http.StatusForbidden)
		case "/++bad-xml":
			_, _ = resp.Write([]byte("<not-closed"))
		default:
			_, _ = resp.Write([]byte("<root><value>ok</value></root>"))
		}
	})

	httpClient, fakeServer := testingHTTPClient(handler)
	defer fakeServer.Close()

	config.Client = httpClient

	var val struct {
		Value string `xml:"value"`
	}

	require.NoError(t, config.GetXML("++ok", nil, &val))
	require.Equal(t, "ok", val.Value)

	err := config.GetXML("++bad-status", nil, &val)
	require.Error(t, err)
	require.Contains(t, err.Error(), "request failed")

	err = config.GetXML("++bad-xml", nil, &val)
	require.Error(t, err)
	require.Contains(t, err.Error(), "reading body")
}

func TestSimpleReqOKAndFail(t *testing.T) {
	t.Parallel()

	config := &server.Config{
		URL:       "http://some.host:5678/",
		VerifySSL: false,
		Timeout:   server.Duration{time.Second},
	}

	handler := http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/++ok":
			_, _ = resp.Write([]byte("something OK"))
		default:
			_, _ = resp.Write([]byte("nope"))
		}
	})

	httpClient, fakeServer := testingHTTPClient(handler)
	defer fakeServer.Close()

	config.Client = httpClient

	require.NoError(t, config.SimpleReq("++ok", url.Values{}, 4))
	require.ErrorIs(t, config.SimpleReq("++bad", url.Values{}, 4), server.ErrCmdNotOK)
}

// testingHTTPClient sets up a fake server for testing secReq().
func testingHTTPClient(handler http.Handler) (*http.Client, *httptest.Server) {
	fakeServer := httptest.NewServer(handler)
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
				return (&net.Dialer{Timeout: server.DefaultTimeout}).DialContext(ctx, network, fakeServer.Listener.Addr().String())
			},
		},
	}

	return client, fakeServer
}

/*
func TestGetXML(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	assert := assert.New(t)
	sspy, _ := GetServer(&server.Config{
    Username: "user", Password: "pass", URL: "http://some.host:5678", VerifySSL: false})
	fake := &fakeAPI{}

	sspy.API = mocks.NewMockAPI(mockCtrl)
	params := make(url.Values)

	params.Add("myKey", "theValue")

	client := &http.Response{
		Body:       io.NopCloser(bytes.NewBufferString("Hello World")),
		StatusCode: http.StatusOK,
	}

	fake.SecReqReturns(client, nil)

	body, err := sspy.GetXML("++foo", params)
	assert.Nil(err, "there must not be an error when input data is valid")
	assert.Equal("Hello World", string(body), "the wrong request response was provided")
	assert.Equal(1, fake.SecReqCallCount(), "secReq must be called exactly once per invocation")

	calledWithPath, calledWithParams, calledWithClient := fake.SecReqArgsForCall(0)
	assert.Equal("++foo", calledWithPath, "the api path was not correct in the request")
	assert.Equal("theValue", calledWithParams.Get("myKey"), "the custom parameter was not set")
	assert.Equal(server.DefaultTimeout, calledWithClient.Timeout, "default timeout must be applied to the request")

	// try again with a bad status.
	client = &http.Response{
		Body:       io.NopCloser(bytes.NewBufferString("Hello World")),
		StatusCode: http.StatusForbidden,
	}
	fake.SecReqReturns(client, nil)

	_, err = sspy.GetXML("++foo", params)
	assert.Contains(err.Error(), "request failed", "the wrong error was returned")
	assert.Equal(2, fake.SecReqCallCount(), "secReq must be called exactly once per invocation")
}

func TestSimpleReq(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	assert := assert.New(t)
	sspy, _ := GetServer(&server.Config{
    Username: "user", Password: "pass", URL: "http://some.host:5678", VerifySSL: false})
	fake := &fakeAPI{}
	sspy.API = mocks.NewMockAPI(mockCtrl)
	params := make(url.Values)

	params.Add("myKey", "theValue")

	client := &http.Response{
		Body:       io.NopCloser(bytes.NewBufferString("Hello World")),
		StatusCode: http.StatusOK,
	}
	fake.SecReqReturns(client, nil)

	err := sspy.SimpleReq("++apipath", params, 3)
	assert.Equal(err, server.ErrorCmdNotOK, "hello world must produce an err")
	assert.Equal(1, fake.SecReqCallCount(), "secReq must be called exactly once per invocation")

	// OK response.
	client = &http.Response{
		Body:       io.NopCloser(bytes.NewBufferString("Hello World is OK")),
		StatusCode: http.StatusOK,
	}

	fake.SecReqReturns(client, nil)

	err = sspy.SimpleReq("++apipath", params, 3)
	assert.Nil(err, "the responds ends with OK so we must have no error")
	assert.Equal(2, fake.SecReqCallCount(), "secReq must be called exactly once per invocation")

	calledWithPath, calledWithParams, calledWithClient := fake.SecReqArgsForCall(1)
	assert.Equal("++apipath", calledWithPath, "the api path was not correct in the request")
	assert.Equal("3", calledWithParams.Get("cameraNum"), "the camera number was not in the parameters")
	assert.Equal(server.DefaultTimeout, calledWithClient.Timeout, "default timeout must be applied to the request")

	// test another error
	fake.SecReqReturns(client, server.ErrorCmdNotOK)

	err = sspy.SimpleReq("++apipath", params, 3)
	assert.Equal(server.ErrorCmdNotOK, err, "the error from secreq must be returned")
	assert.Equal(3, fake.SecReqCallCount(), "secReq must be called exactly once per invocation")
}

/**/
