package securityspy_test

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golift.io/securityspy"
	"golift.io/securityspy/server"
)

func TestGetServer(t *testing.T) {
	t.Parallel()

	asert := assert.New(t)
	URL := "http://127.0.0.1:5678"
	user := "user123"
	pass := "pass456"
	secspyServer, err := securityspy.New(&server.Config{Username: user, Password: pass, URL: URL, VerifySSL: true})

	require.Error(t, err, "there is no server at the address provided so an error must exist")
	asert.NotNil(secspyServer, "server must not be nil. even wiuth an error it must be returned")

	if !strings.Contains(err.Error(), "target machine actively refused it") &&
		!strings.Contains(err.Error(), "connection refused") {
		t.Error("error does not contain the correct messages.")
	}
}

func TestRefresh(t *testing.T) {
	t.Parallel()

	var requestCount int

	fakeServer := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/++systemInfo" {
			http.NotFound(resp, req)

			return
		}

		requestCount++
		if requestCount == 1 {
			resp.Header().Set("Content-Type", "application/xml")
			_, _ = resp.Write([]byte(testSystemInfo))

			return
		}

		http.Error(resp, "bad xml", http.StatusInternalServerError)
	}))
	defer fakeServer.Close()

	asert := assert.New(t)
	secspyServer := securityspy.NewMust(
		&server.Config{Username: "user", Password: "pass", URL: fakeServer.URL + "/", VerifySSL: false})
	require.NoError(t, secspyServer.Refresh(),
		"an error must not be returned while testing with valid XML")

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

	// second call returns an HTTP error response.
	require.Error(t, secspyServer.Refresh())
}

func TestGetSounds(t *testing.T) { //nolint:dupl // it just looks like a duplicate, but it's not.
	t.Parallel()

	var requestCount int

	fakeServer := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/++sounds" {
			http.NotFound(resp, req)
			return
		}

		requestCount++
		if requestCount == 1 {
			resp.Header().Set("Content-Type", "application/xml")
			_, _ = resp.Write([]byte(testSoundsList))

			return
		}

		http.Error(resp, "fail", http.StatusInternalServerError)
	}))
	defer fakeServer.Close()

	asert := assert.New(t)
	secspyServer := securityspy.NewMust(
		&server.Config{Username: "user", Password: "pass", URL: fakeServer.URL + "/", VerifySSL: false})

	sounds, err := secspyServer.GetSounds()
	require.NoError(t, err, "the method must not return an error when given valid XML to unmarshal")
	asert.Len(sounds, 20, "all 20 sounds must exist in the slice")
	asert.Equal("Beeps.aif", sounds[0], "the sound files were not properly unmarhsalled")

	_, err = secspyServer.GetSounds()
	require.Error(t, err)
}

func TestGetScripts(t *testing.T) { //nolint:dupl // it just looks like a duplicate, but it's not.
	t.Parallel()

	var requestCount int

	fakeServer := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/++scripts" {
			http.NotFound(resp, req)
			return
		}

		requestCount++
		if requestCount == 1 {
			resp.Header().Set("Content-Type", "application/xml")
			_, _ = resp.Write([]byte(testScriptsList))

			return
		}

		http.Error(resp, "fail", http.StatusInternalServerError)
	}))
	defer fakeServer.Close()

	asert := assert.New(t)
	secspyServer := securityspy.NewMust(
		&server.Config{Username: "user", Password: "pass", URL: fakeServer.URL + "/", VerifySSL: false})

	scripts, err := secspyServer.GetScripts()
	require.NoError(t, err, "the method must not return an error when given valid XML to unmarshal")
	asert.Len(scripts, 16, "all 16 scripts must exist in the slice")
	asert.Equal("Web-i Activate Relay 1.scpt", scripts[0], "the script files were not properly unmarhsalled")

	_, err = secspyServer.GetScripts()
	require.Error(t, err)
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

	xmlData := strings.Replace(testSystemInfo, "<ptzcapabilities>0</ptzcapabilities>", "", 1)
	xmlData = strings.Replace(xmlData,
		"<schedule-id-a>3</schedule-id-a>", "<schedule-id-a>99999</schedule-id-a>", 1)
	xmlData = strings.Replace(xmlData,
		"<schedule-override-a>2</schedule-override-a>", "<schedule-override-a>99999</schedule-override-a>", 1)

	fakeServer := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/++systemInfo" {
			http.NotFound(resp, req)
			return
		}

		resp.Header().Set("Content-Type", "application/xml")
		_, _ = resp.Write([]byte(xmlData))
	}))
	defer fakeServer.Close()

	secspyServer := securityspy.NewMust(
		&server.Config{Username: "user", Password: "pass", URL: fakeServer.URL + "/", VerifySSL: false})

	require.NoError(t, secspyServer.Refresh())
	require.Empty(t, secspyServer.Cameras.ByNum(1).ScheduleIDA.Name)
	require.Empty(t, secspyServer.Cameras.ByNum(1).ScheduleOverrideA.Name)
}

/**/
