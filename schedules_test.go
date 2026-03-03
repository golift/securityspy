package securityspy_test

import (
	"encoding/xml"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golift.io/securityspy"
	"golift.io/securityspy/mocks"
	"golift.io/securityspy/server"
)

func TestSetSchedulePreset(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	secspyServer := securityspy.NewMust(
		&server.Config{Username: "user", Password: "pass", URL: "http://127.0.0.1:5678", VerifySSL: false})
	fake := mocks.NewMockAPI(mockCtrl) // create a fake api interface that provides introspection methods.
	secspyServer.API = fake            // override our internal api interface with a fake interface.

	fake.EXPECT().SimpleReq("++ssSetPreset", url.Values{"id": []string{"1"}}, -1)
	require.NoError(t, secspyServer.SetSchedulePreset(1), "this method must not return an error during testing")
}

func TestUnmarshalXMLscheduleContainer(t *testing.T) {
	t.Parallel()

	asert := assert.New(t)

	var schedule securityspy.ScheduleContainer

	err := xml.Unmarshal([]byte(testScheduleList), &schedule)
	require.NoError(t, err, "valid data must not produce an error")
	asert.Equal("Armed 24/7", schedule[1], "the scheduleContainer data did not unmarshal properly")
	asert.Len(schedule, 6)

	err = xml.Unmarshal([]byte("<gotrekt>"), &schedule)
	require.Error(t, err, "invalid data must produce an error")

	err = xml.Unmarshal([]byte("<gotrekt><server></gotrekt>"), &schedule)
	require.Error(t, err, "invalid data must produce an error")
}

/**/
