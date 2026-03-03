package securityspy_test

import (
	"encoding/xml"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golift.io/securityspy"
	"golift.io/securityspy/mocks"
	"golift.io/securityspy/server"
)

var errTest = errors.New("error goes here")

func TestGetServer(t *testing.T) {
	t.Parallel()

	asert := assert.New(t)
	URL := "http://127.0.0.1:5678"
	user := "user123"
	pass := "pass456"
	secspyServer, err := securityspy.New(&server.Config{Username: user, Password: pass, URL: URL, VerifySSL: true})

	require.Error(t, err, "there is no server at the address provided so an error must exist")
	asert.NotNil(secspyServer, "server must not be nil. even wiuth an error it must be returned")
	asert.NotNil(secspyServer.API, "api interface pointer must be created by GetServer")

	if !strings.Contains(err.Error(), "target machine actively refused it") &&
		!strings.Contains(err.Error(), "connection refused") {
		t.Error("error does not contain the correct messages.")
	}
}

func TestRefresh(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	asert := assert.New(t)
	secspyServer := securityspy.NewMust(
		&server.Config{Username: "user", Password: "pass", URL: "http://127.0.0.1:5678", VerifySSL: false})
	fake := mocks.NewMockAPI(mockCtrl)
	secspyServer.API = fake

	fake.EXPECT().GetXML(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Do(
		func(_, _, v any) {
			_ = xml.Unmarshal([]byte(testSystemInfo), &v)
		},
	)
	require.NoError(t, secspyServer.Refresh(),
		"an error must not be returned while testing with an overridden api interface")

	// Make sure Refresh() did all the things it is supposed to do.
	asert.WithinDuration(time.Now(), secspyServer.Info.Refreshed, time.Second,
		"Refreshed field must be updated by Refresh() method")

	// Test that the data was unmarshalled properly.
	// These tests assume the test data does not change.
	asert.Equal("SecuritySpy", secspyServer.Info.Name, "the server's name was not properly unmarshalled")
	asert.Equal("2019-02-10T15:53:23", secspyServer.Info.CurrentTime.Format("2006-01-02T15:04:05"),
		"the server's current time was not properly unmarshalled")
	asert.Equal(2304, secspyServer.Cameras.ByNum(1).Width, "camera info was not properly unmarshalled")
	asert.Equal("Road", secspyServer.Cameras.ByNum(2).Name, "camera info was not properly unmarshalled")
	asert.Equal("Unarmed 24/7", secspyServer.Info.ServerSchedules[0], "schedule info was not properly unmarshalled")
	asert.Equal("None", secspyServer.Info.ScheduleOverrides[0], "schedule override info was not properly unmarshalled")
	asert.Equal("MyFirstPreset", secspyServer.Info.SchedulePresets[1930238093],
		"schedule preset info was not properly unmarshalled")

	// make sure bad xml returns an expected error
	fake.EXPECT().GetXML(gomock.Any(), gomock.Any(), gomock.Any()).Return(errTest)
	require.ErrorIs(t, secspyServer.Refresh(), errTest)
}

func TestGetSounds(t *testing.T) { //nolint:dupl // it just looks like a duplicate, but it's not.
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	asert := assert.New(t)
	secspyServer := securityspy.NewMust(
		&server.Config{Username: "user", Password: "pass", URL: "http://127.0.0.1:5678", VerifySSL: false})
	fake := mocks.NewMockAPI(mockCtrl) // create a fake api interface that provides introspection methods.
	secspyServer.API = fake

	fake.EXPECT().GetXML(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Do(
		func(_, _, v any) {
			_ = xml.Unmarshal([]byte(testSoundsList), &v)
		},
	)

	sounds, err := secspyServer.GetSounds()
	require.NoError(t, err, "the method must not return an error when given valid XML to unmarshal")
	asert.Len(sounds, 20, "all 20 sounds must exist in the slice")
	asert.Equal("Beeps.aif", sounds[0], "the sound files were not properly unmarhsalled")

	// Test error conditions.
	fake.EXPECT().GetXML(gomock.Any(), gomock.Any(), gomock.Any()).Return(errTest)

	_, err = secspyServer.GetSounds()
	require.ErrorIs(t, err, errTest)
}

func TestGetScripts(t *testing.T) { //nolint:dupl // it just looks like a duplicate, but it's not.
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	asert := assert.New(t)
	secspyServer := securityspy.NewMust(
		&server.Config{Username: "user", Password: "pass", URL: "http://127.0.0.1:5678", VerifySSL: false})
	fake := mocks.NewMockAPI(mockCtrl) // create a fake api interface that provides introspection methods.
	secspyServer.API = fake

	fake.EXPECT().GetXML(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Do(
		func(_, _, v any) {
			_ = xml.Unmarshal([]byte(testScriptsList), &v)
		},
	)

	scripts, err := secspyServer.GetScripts()
	require.NoError(t, err, "the method must not return an error when given valid XML to unmarshal")
	asert.Len(scripts, 16, "all 16 scripts must exist in the slice")
	asert.Equal("Web-i Activate Relay 1.scpt", scripts[0], "the script files were not properly unmarhsalled")

	// Test error conditions.
	fake.EXPECT().GetXML(gomock.Any(), gomock.Any(), gomock.Any()).Return(errTest)

	_, err = secspyServer.GetScripts()
	require.ErrorIs(t, err, errTest)
}

func TestUnmarshalXMLYesNoBool(t *testing.T) {
	t.Parallel()

	asert := assert.New(t)
	good := []string{"true", "yes", "1", "armed", "active", "enabled"}
	fail := []string{"anything", "else", "returns", "false", "including", "no", "0", "disarmed", "inactive", "disabled"}

	var bit securityspy.YesNoBool

	for _, val := range good {
		require.NoError(t, xml.Unmarshal([]byte("<tag>"+val+"</tag>"), &bit), "unmarshalling must not produce an error")
		asert.True(bit.Val, "the value must unmarshal to true")
		asert.Equal(val, bit.Txt, "the value was not unmarshalled correctly")
	}

	for _, val := range fail {
		require.NoError(t, xml.Unmarshal([]byte("<tag>"+val+"</tag>"), &bit), "unmarshalling must not produce an error")
		asert.False(bit.Val, "the value must unmarshal to false")
		asert.Equal(val, bit.Txt, "the value was not unmarshalled correctly")
	}
}

func TestUnmarshalXMLDuration(t *testing.T) {
	t.Parallel()

	asert := assert.New(t)
	good := []string{"1", "20", "300", "4000", "50000", "666666"}

	var bit securityspy.Duration

	for _, val := range good {
		require.NoError(t, xml.Unmarshal([]byte("<tag>"+val+"</tag>"), &bit), "unmarshalling must not produce an error")
		asert.Equal(val, bit.Val, "the value was not unmarshalled correctly")
		num, err := strconv.ParseFloat(val, 64)
		require.NoError(t, err, "must not be an error parsing test numbers")
		asert.InDelta(num, bit.Seconds(), 0.000001, "the value was not unmarshalled correctly")
		asert.Equal(val, bit.Val, "the value was not unmarshalled correctly")
	}

	// Test empty value.
	require.NoError(t, xml.Unmarshal([]byte("<tag></tag>"), &bit), "unmarshalling must not produce an error")
	asert.Empty(bit.Val, "the value was not unmarshalled correctly")
	asert.Equal(int64(-1), bit.Nanoseconds(), "an empty value must produce -1 nano second.")
}

func TestRefreshHandlesNilPTZAndMissingSchedules(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	secspyServer := securityspy.NewMust(
		&server.Config{Username: "user", Password: "pass", URL: "http://127.0.0.1:5678", VerifySSL: false})
	fake := mocks.NewMockAPI(mockCtrl)
	secspyServer.API = fake

	xmlData := strings.Replace(testSystemInfo, "<ptzcapabilities>0</ptzcapabilities>", "", 1)
	xmlData = strings.Replace(xmlData,
		"<schedule-id-a>3</schedule-id-a>", "<schedule-id-a>99999</schedule-id-a>", 1)
	xmlData = strings.Replace(xmlData,
		"<schedule-override-a>2</schedule-override-a>", "<schedule-override-a>99999</schedule-override-a>", 1)

	fake.EXPECT().GetXML(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Do(
		func(_, _, v any) {
			_ = xml.Unmarshal([]byte(xmlData), &v)
		},
	)

	require.NoError(t, secspyServer.Refresh())
	require.Empty(t, secspyServer.Cameras.ByNum(1).ScheduleIDA.Name)
	require.Empty(t, secspyServer.Cameras.ByNum(1).ScheduleOverrideA.Name)
}

/**/
