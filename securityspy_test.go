package securityspy

import (
	"bytes"
	"context"
	"encoding/base64"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestGetServer(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	URL := "http://127.0.0.1:5678"
	user := "user123"
	pass := "pass456"
	b64 := base64.URLEncoding.EncodeToString([]byte(user + ":" + pass))
	server, err := GetServer(user, pass, URL, true)
	assert.NotNil(err, "there is no server at the address provided so an error must exist")
	assert.NotNil(server, "server must not be nil. even wiuth an error it must be returned")
	assert.NotNil(server.systemInfo, "systemInfo pointer must be created by GetServer")
	assert.NotNil(server.systemInfo.Server, "systemInfo.Server pointer must be created by GetServer")
	assert.NotNil(server.api, "api interface pointer must be created by GetServer")
	assert.NotNil(server.Info, "ServerInfo pointer must be created by GetServer")
	assert.NotNil(server.Files.server, "Files and Files.server pointers must be created by GetServer")
	assert.NotNil(server.Events.server, "Events and Events.server pointers must be created by GetServer")
	assert.NotNil(server.Events.eventBinds, "eventBinds map must be created by GetServer")
	assert.NotNil(server.Events.eventChans, "eventChans map must be created by GetServer")
	assert.Contains(err.Error(), "http.Do(req)", "the wrong error was returned")
	assert.Equal(user, server.username, "the username must be saved by GetServer")
	assert.Equal(URL+"/", server.baseURL, "the url must be saved by GetServer after adding a / suffix")
	assert.Equal(b64, server.authB64, "the base64 encoding of user/pass must be saved by GetServer")
	assert.True(server.verifySSL, "SSL certificate checking was requested so it must be true")
}

func TestRefresh(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	server, _ := GetServer("user", "pass", "http://127.0.0.1:5678", false)
	fake := &fakeAPI{}                                 // create a fake api interface that provides introspection methods.
	server.api = fake                                  // override our internal api interface with a fake interface.
	fake.SecReqXMLReturns([]byte(testSystemInfo), nil) // Pass in a test XML payload.
	assert.Nil(server.Refresh(), "an error must not be returned while testing with an overridden api interface")

	// Make sure Refresh() did all the things it is supposed to do.
	assert.EqualValues(server.systemInfo.Schedules, server.Info.ServerSchedules, "unexported server schedules must be copied into exported Info struct")
	assert.EqualValues(server.systemInfo.SchedulePresets, server.Info.SchedulePresets, "unexported schedule presets must be copied into exported Info struct")
	assert.EqualValues(server.systemInfo.ScheduleOverrides, server.Info.ScheduleOverrides, "unexported schedule overrides must be copied into exported Info struct")
	assert.Equal(server, server.Cameras.server, "server struct pointer must be copied into Cameras struct")
	assert.Equal(2, len(server.Cameras.Names), "both test data camera names must be saved into a convenience slice")
	assert.Equal(2, len(server.Cameras.Numbers), "both test data camera numbers must be saved into a convenience slice")
	assert.WithinDuration(time.Now(), server.Info.Refreshed, time.Second, "Refreshed field must be updated by Refresh() method")

	// Test that the data was unmarshalled properly.
	// These tests assume the test data does not change.
	assert.EqualValues("SecuritySpy", server.Info.Name, "the server's name was not properly unmarshalled")
	assert.Equal("2019-02-10T15:53:23", server.Info.CurrentTime.Format("2006-01-02T15:04:05"), "the server's current time was not properly unmarshalled")
	assert.Equal(2304, server.systemInfo.CameraList.Cameras[0].Width, "camera info was not properly unmarshalled")
	assert.Equal("Road", server.systemInfo.CameraList.Cameras[1].Name, "camera info was not properly unmarshalled")
	assert.Equal("Unarmed 24/7", server.Info.ServerSchedules[0], "schedule info was not properly unmarshalled")
	assert.Equal("None", server.Info.ScheduleOverrides[0], "schedule override info was not properly unmarshalled")
	assert.Equal("MyFirstPreset", server.Info.SchedulePresets[1930238093], "schedule preset info was not properly unmarshalled")

	// make sure bad xml returns an expected error
	fake.SecReqXMLReturns([]byte("<xml>broken<xml/>"), nil) // Pass in a broken XML payload.
	assert.Contains(server.Refresh().Error(), "xml.Unmarshal(++systemInfo)", "xml unmarhsalling must fail and produce this error")
}

func TestGetSounds(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	server, _ := GetServer("user", "pass", "http://127.0.0.1:5678", false)
	fake := &fakeAPI{}                                 // create a fake api interface that provides introspection methods.
	server.api = fake                                  // override our internal api interface with a fake interface.
	fake.SecReqXMLReturns([]byte(testSoundsList), nil) // Pass in a test XML payload.
	sounds, err := server.GetSounds()
	assert.Nil(err, "the method must not return an error when given valid XML to unmarshal")
	assert.Equal(20, len(sounds), "all 20 sounds must exist in the slice")
	assert.Equal("Beeps.aif", sounds[0], "the sound files were not properly unmarhsalled")

	// Test error conditions.
	fake.SecReqXMLReturns([]byte(testSoundsList), errors.New("error goes here")) // Pass in a test XML payload.
	_, err = server.GetSounds()
	assert.EqualError(err, "error goes here", "the error from secReqXML must be returned")
	fake.SecReqXMLReturns([]byte("bad xml goes here"), nil) // Pass in a bad XML payload.
	_, err = server.GetSounds()
	assert.Contains(err.Error(), "xml.Unmarshal(++sounds)", "the error from xml.Unmarshal must be returned")
}

func TestGetScripts(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	server, _ := GetServer("user", "pass", "http://127.0.0.1:5678", false)
	fake := &fakeAPI{}                                  // create a fake api interface that provides introspection methods.
	server.api = fake                                   // override our internal api interface with a fake interface.
	fake.SecReqXMLReturns([]byte(testScriptsList), nil) // Pass in a test XML payload.
	scripts, err := server.GetScripts()
	assert.Nil(err, "the method must not return an error when given valid XML to unmarshal")
	assert.Equal(16, len(scripts), "all 16 scripts must exist in the slice")
	assert.Equal("Web-i Activate Relay 1.scpt", scripts[0], "the script files were not properly unmarhsalled")

	// Test error conditions.
	fake.SecReqXMLReturns([]byte(testScriptsList), errors.New("error goes here")) // Pass in a test XML payload.
	_, err = server.GetScripts()
	assert.EqualError(err, "error goes here", "the error from secReqXML must be returned")
	fake.SecReqXMLReturns([]byte("bad xml goes here"), nil) // Pass in a bad XML payload.
	_, err = server.GetScripts()
	assert.Contains(err.Error(), "xml.Unmarshal(++scripts)", "the error from xml.Unmarshal must be returned")
}

func TestGetClient(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	server := &Server{verifySSL: true}
	client := server.getClient(DefaultTimeout + 7*time.Second)
	assert.Equal(DefaultTimeout+7*time.Second, client.Timeout, "timeout was not applied to the client")
	// no way to check the verifySSL parameter?
}

func TestSecReq(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	server, _ := GetServer("user", "pass", "http://some.host:5678", false)
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal("xml", r.FormValue("format"), "format parameter was not added")
		assert.Equal(server.authB64, r.FormValue("auth"), "auth parameter was not added")
		assert.Equal("application/xml", r.Header.Get("Accept"), "accept header is not correct")
		_, err := w.Write([]byte("request OK"))
		assert.Nil(err, "the fake server must return an error writing to the client")
		assert.Equal(server.baseURL, "http://"+r.Host+"/", "the host was not set correctly in the request")
	})
	httpClient, close := testingHTTPClient(h)
	defer close()
	resp, err := server.secReq("++path", make(url.Values), httpClient)
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

func TestSecReqXML(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	server, _ := GetServer("user", "pass", "http://some.host:5678", false)
	fake := &fakeAPI{}
	server.api = fake
	params := make(url.Values)
	params.Add("myKey", "theValue")
	client := &http.Response{
		Body:       ioutil.NopCloser(bytes.NewBufferString("Hello World")),
		StatusCode: http.StatusOK,
	}
	fake.SecReqReturns(client, nil)
	body, err := server.secReqXML("++foo", params)
	assert.Nil(err, "there must not be an error when input data is valid")
	assert.Equal("Hello World", string(body), "the wrong request response was provided")
	assert.Equal(1, fake.SecReqCallCount(), "secReq must be called exactly once per invocation")
	calledWithPath, calledWithParams, calledWithClient := fake.SecReqArgsForCall(0)
	assert.Equal("++foo", calledWithPath, "the api path was not correct in the request")
	assert.Equal("theValue", calledWithParams.Get("myKey"), "the custom parameter was not set")
	assert.Equal(DefaultTimeout, calledWithClient.Timeout, "default timeout must be applied to the request")
	// TODO: check that the correct data was passed INTO secReq().

	// try again with a bad status.
	client = &http.Response{
		Body:       ioutil.NopCloser(bytes.NewBufferString("Hello World")),
		StatusCode: http.StatusForbidden,
	}
	fake.SecReqReturns(client, nil)
	_, err = server.secReqXML("++foo", params)
	assert.Contains(err.Error(), "request failed", "the wrong error was returned")
	assert.Equal(2, fake.SecReqCallCount(), "secReq must be called exactly once per invocation")

}

func TestSimpleReq(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	server, _ := GetServer("user", "pass", "http://some.host:5678", false)
	fake := &fakeAPI{}
	server.api = fake
	params := make(url.Values)
	params.Add("myKey", "theValue")
	client := &http.Response{
		Body:       ioutil.NopCloser(bytes.NewBufferString("Hello World")),
		StatusCode: http.StatusOK,
	}
	fake.SecReqReturns(client, nil)
	err := server.simpleReq("++apipath", params, 3)
	assert.Equal(err, ErrorCmdNotOK, "hello world must produce an err")
	assert.Equal(1, fake.SecReqCallCount(), "secReq must be called exactly once per invocation")

	// OK response.
	client = &http.Response{
		Body:       ioutil.NopCloser(bytes.NewBufferString("Hello World is OK")),
		StatusCode: http.StatusOK,
	}
	fake.SecReqReturns(client, nil)
	err = server.simpleReq("++apipath", params, 3)
	assert.Nil(err, "the responds ends with OK so we must have no error")
	assert.Equal(2, fake.SecReqCallCount(), "secReq must be called exactly once per invocation")
	calledWithPath, calledWithParams, calledWithClient := fake.SecReqArgsForCall(1)
	assert.Equal("++apipath", calledWithPath, "the api path was not correct in the request")
	assert.Equal("3", calledWithParams.Get("cameraNum"), "the camera number was not in the parameters")
	assert.Equal(DefaultTimeout, calledWithClient.Timeout, "default timeout must be applied to the request")

	// test another error
	fake.SecReqReturns(client, ErrorCmdNotOK)
	err = server.simpleReq("++apipath", params, 3)
	assert.Equal(ErrorCmdNotOK, err, "the error from secreq must be returned")
	assert.Equal(3, fake.SecReqCallCount(), "secReq must be called exactly once per invocation")
}
