package securityspy_test

import (
	"encoding/xml"
	"net/url"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golift.io/securityspy"
	"golift.io/securityspy/mocks"
	"golift.io/securityspy/server"
)

func TestSetSchedulePreset(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	assert := assert.New(t)
	server := securityspy.NewMust(
		&server.Config{Username: "user", Password: "pass", URL: "http://127.0.0.1:5678", VerifySSL: false})
	fake := mocks.NewMockAPI(mockCtrl) // create a fake api interface that provides introspection methods.
	server.API = fake                  // override our internal api interface with a fake interface.

	fake.EXPECT().SimpleReq("++ssSetPreset", url.Values{"id": []string{"1"}}, -1)
	assert.Nil(server.SetSchedulePreset(1), "this method must not return an error during testing")
}

func TestUnmarshalXMLscheduleContainer(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	var schedule securityspy.ScheduleContainer

	err := xml.Unmarshal([]byte(testScheduleList), &schedule)
	assert.Nil(err, "valid data must not produce an error")
	assert.Equal("Armed 24/7", schedule[1], "the scheduleContainer data did not unmarshal properly")
	assert.Equal(6, len(schedule))

	err = xml.Unmarshal([]byte("<gotrekt>"), &schedule)
	assert.NotNil(err, "invalid data must produce an error")

	err = xml.Unmarshal([]byte("<gotrekt><server></gotrekt>"), &schedule)
	assert.NotNil(err, "invalid data must produce an error")
}

/**/
