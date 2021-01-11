package server_test

import (
	"context"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golift.io/securityspy/server"
)

func TestGet(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)
	c := &server.Config{
		Username:  "user",
		Password:  "pass",
		URL:       "http://some.host:5678/",
		VerifySSL: false,
		Timeout:   time.Second,
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal("xml", r.FormValue("format"), "format parameter was not added")
		assert.Equal(c.Password, r.FormValue("auth"), "auth parameter was not added")
		assert.Equal("application/xml", r.Header.Get("Accept"), "accept header is not correct")
		_, err := w.Write([]byte("request OK"))
		assert.Nil(err, "the fake server must return an error writing to the client")
		assert.Equal(c.URL, "http://"+r.Host+"/", "the host was not set correctly in the request")
	})

	httpClient, close := testingHTTPClient(h)
	defer close()

	c.Client = httpClient
	resp, err := c.Get("++path", make(url.Values))
	assert.Nil(err, "the method must not return an error when given a valid server to query")

	if err == nil {
		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode, "the server must return a 200 response code")
		body, err := ioutil.ReadAll(resp.Body)
		assert.Nil(err, "must not be an error reading the response body")
		assert.Equal("request OK", string(body), "wrong data was returned from the server")
	}
}

// testingHTTPClient sets up a fake server for testing secReq().
func testingHTTPClient(handler http.Handler) (*http.Client, func()) {
	fakeServer := httptest.NewServer(handler)
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, network, _ string) (net.Conn, error) {
				return net.Dial(network, fakeServer.Listener.Addr().String())
			},
		},
	}

	return client, fakeServer.Close
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
		Body:       ioutil.NopCloser(bytes.NewBufferString("Hello World")),
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
		Body:       ioutil.NopCloser(bytes.NewBufferString("Hello World")),
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
		Body:       ioutil.NopCloser(bytes.NewBufferString("Hello World")),
		StatusCode: http.StatusOK,
	}
	fake.SecReqReturns(client, nil)

	err := sspy.SimpleReq("++apipath", params, 3)
	assert.Equal(err, server.ErrorCmdNotOK, "hello world must produce an err")
	assert.Equal(1, fake.SecReqCallCount(), "secReq must be called exactly once per invocation")

	// OK response.
	client = &http.Response{
		Body:       ioutil.NopCloser(bytes.NewBufferString("Hello World is OK")),
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
