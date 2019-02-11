package securityspy

import (
	"encoding/base64"
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

/**/
