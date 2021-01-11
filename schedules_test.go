package securityspy

import (
	"encoding/xml"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetSchedulePreset(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)
	server, _ := GetServer(&Config{Username: "user", Password: "pass", URL: "http://127.0.0.1:5678", VerifySSL: false})
	fake := &fakeAPI{} // create a fake api interface that provides introspection methods.
	server.api = fake  // override our internal api interface with a fake interface.

	assert.Nil(server.SetSchedulePreset(1), "this method must not return an error during testing")
	assert.Equal(1, fake.SimpleReqCallCount(), "this method must call simpleReq() exactly once")

	cmd, params, cameraNum := fake.SimpleReqArgsForCall(0) // check the web request parameters
	assert.Equal(url.Values{"id": []string{"1"}}, params, "the presetID sent to securityspy was incorrect")
	assert.Equal("++ssSetPreset", cmd, "the wrong command was used to invoke a securityspy schedule preset")
	assert.Equal(-1, cameraNum, "this method does not use a camera and must be -1")
}

func TestUnmarshalXMLscheduleContainer(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	var s scheduleContainer

	err := xml.Unmarshal([]byte(testScheduleList), &s)
	assert.Nil(err, "valid data must not produce an error")
	assert.Equal("Armed 24/7", s[1], "the scheduleContainer data did not unmarshal properly")
	assert.Equal(6, len(s))

	err = xml.Unmarshal([]byte("<gotrekt>"), &s)
	assert.NotNil(err, "invalid data must produce an error")

	err = xml.Unmarshal([]byte("<gotrekt><server></gotrekt>"), &s)
	assert.NotNil(err, "invalid data must produce an error")
}

/**/
