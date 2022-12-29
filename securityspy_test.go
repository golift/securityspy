package securityspy_test

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golift.io/securityspy"
	"golift.io/securityspy/mocks"
	"golift.io/securityspy/server"
)

var errTest = fmt.Errorf("error goes here")

func TestGetServer(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)
	URL := "http://127.0.0.1:5678"
	user := "user123"
	pass := "pass456"
	server, err := securityspy.New(&server.Config{Username: user, Password: pass, URL: URL, VerifySSL: true})

	assert.NotNil(err, "there is no server at the address provided so an error must exist")
	assert.NotNil(server, "server must not be nil. even wiuth an error it must be returned")
	assert.NotNil(server.API, "api interface pointer must be created by GetServer")

	if !strings.Contains(err.Error(), "target machine actively refused it") &&
		!strings.Contains(err.Error(), "connection refused") {
		t.Error("error does not contain the correct messages.")
	}
}

func TestRefresh(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	assert := assert.New(t)
	server := securityspy.NewMust(
		&server.Config{Username: "user", Password: "pass", URL: "http://127.0.0.1:5678", VerifySSL: false})
	fake := mocks.NewMockAPI(mockCtrl)
	server.API = fake

	fake.EXPECT().GetXML(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Do(
		func(_, _, v interface{}) {
			_ = xml.Unmarshal([]byte(testSystemInfo), &v)
		},
	)
	assert.Nil(server.Refresh(), "an error must not be returned while testing with an overridden api interface")

	// Make sure Refresh() did all the things it is supposed to do.
	assert.EqualValues(server.Info.ServerSchedules, server.Info.ServerSchedules,
		"unexported server schedules must be copied into exported Info struct")
	assert.EqualValues(server.Info.SchedulePresets, server.Info.SchedulePresets,
		"unexported schedule presets must be copied into exported Info struct")
	assert.EqualValues(server.Info.ScheduleOverrides, server.Info.ScheduleOverrides,
		"unexported schedule overrides must be copied into exported Info struct")
	assert.WithinDuration(time.Now(), server.Info.Refreshed, time.Second,
		"Refreshed field must be updated by Refresh() method")

	// Test that the data was unmarshalled properly.
	// These tests assume the test data does not change.
	assert.EqualValues("SecuritySpy", server.Info.Name, "the server's name was not properly unmarshalled")
	assert.Equal("2019-02-10T15:53:23", server.Info.CurrentTime.Format("2006-01-02T15:04:05"),
		"the server's current time was not properly unmarshalled")
	assert.Equal(2304, server.Cameras.ByNum(1).Width, "camera info was not properly unmarshalled")
	assert.Equal("Road", server.Cameras.ByNum(2).Name, "camera info was not properly unmarshalled")
	assert.Equal("Unarmed 24/7", server.Info.ServerSchedules[0], "schedule info was not properly unmarshalled")
	assert.Equal("None", server.Info.ScheduleOverrides[0], "schedule override info was not properly unmarshalled")
	assert.Equal("MyFirstPreset", server.Info.SchedulePresets[1930238093],
		"schedule preset info was not properly unmarshalled")

	// make sure bad xml returns an expected error
	fake.EXPECT().GetXML(gomock.Any(), gomock.Any(), gomock.Any()).Return(errTest)
	assert.ErrorIs(server.Refresh(), errTest)
}

func TestGetSounds(t *testing.T) { //nolint:dupl
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	assert := assert.New(t)
	server := securityspy.NewMust(
		&server.Config{Username: "user", Password: "pass", URL: "http://127.0.0.1:5678", VerifySSL: false})
	fake := mocks.NewMockAPI(mockCtrl) // create a fake api interface that provides introspection methods.
	server.API = fake

	fake.EXPECT().GetXML(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Do(
		func(_, _, v interface{}) {
			_ = xml.Unmarshal([]byte(testSoundsList), &v)
		},
	)

	sounds, err := server.GetSounds()
	assert.Nil(err, "the method must not return an error when given valid XML to unmarshal")
	assert.Equal(20, len(sounds), "all 20 sounds must exist in the slice")
	assert.Equal("Beeps.aif", sounds[0], "the sound files were not properly unmarhsalled")

	// Test error conditions.
	fake.EXPECT().GetXML(gomock.Any(), gomock.Any(), gomock.Any()).Return(errTest)

	_, err = server.GetSounds()
	assert.ErrorIs(err, errTest)
}

func TestGetScripts(t *testing.T) { //nolint:dupl
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	assert := assert.New(t)
	server := securityspy.NewMust(
		&server.Config{Username: "user", Password: "pass", URL: "http://127.0.0.1:5678", VerifySSL: false})
	fake := mocks.NewMockAPI(mockCtrl) // create a fake api interface that provides introspection methods.
	server.API = fake

	fake.EXPECT().GetXML(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Do(
		func(_, _, v interface{}) {
			_ = xml.Unmarshal([]byte(testScriptsList), &v)
		},
	)

	scripts, err := server.GetScripts()
	assert.Nil(err, "the method must not return an error when given valid XML to unmarshal")
	assert.Equal(16, len(scripts), "all 16 scripts must exist in the slice")
	assert.Equal("Web-i Activate Relay 1.scpt", scripts[0], "the script files were not properly unmarhsalled")

	// Test error conditions.
	fake.EXPECT().GetXML(gomock.Any(), gomock.Any(), gomock.Any()).Return(errTest)

	_, err = server.GetScripts()
	assert.ErrorIs(err, errTest)
}

func TestUnmarshalXMLYesNoBool(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)
	good := []string{"true", "yes", "1", "armed", "active", "enabled"}
	fail := []string{"anything", "else", "returns", "false", "including", "no", "0", "disarmed", "inactive", "disabled"}

	var bit securityspy.YesNoBool

	for _, val := range good {
		assert.Nil(xml.Unmarshal([]byte("<tag>"+val+"</tag>"), &bit), "unmarshalling must not produce an error")
		assert.True(bit.Val, "the value must unmarshal to true")
		assert.Equal(val, bit.Txt, "the value was not unmarshalled correctly")
	}

	for _, val := range fail {
		assert.Nil(xml.Unmarshal([]byte("<tag>"+val+"</tag>"), &bit), "unmarshalling must not produce an error")
		assert.False(bit.Val, "the value must unmarshal to false")
		assert.Equal(val, bit.Txt, "the value was not unmarshalled correctly")
	}
}

func TestUnmarshalXMLDuration(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)
	good := []string{"1", "20", "300", "4000", "50000", "666666"}

	var bit securityspy.Duration

	for _, val := range good {
		assert.Nil(xml.Unmarshal([]byte("<tag>"+val+"</tag>"), &bit), "unmarshalling must not produce an error")
		assert.Equal(val, bit.Val, "the value was not unmarshalled correctly")
		num, err := strconv.ParseFloat(val, 64)
		assert.Nil(err, "must not be an error parsing test numbers")
		assert.Equal(num, bit.Seconds(), "the value was not unmarshalled correctly")
		assert.Equal(val, bit.Val, "the value was not unmarshalled correctly")
	}

	// Test empty value.
	assert.Nil(xml.Unmarshal([]byte("<tag></tag>"), &bit), "unmarshalling must not produce an error")
	assert.Equal("", bit.Val, "the value was not unmarshalled correctly")
	assert.Equal(int64(-1), bit.Nanoseconds(), "an empty value must produce -1 nano second.")
}

/**/
