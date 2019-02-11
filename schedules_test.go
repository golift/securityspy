package securityspy

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetSchedulePreset(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	server, _ := GetServer("user", "pass", "http://127.0.0.1:5678", false)
	fake := &fakeAPI{} // create a fake api interface that provides introspection methods.
	server.api = fake  // override our internal api interface with a fake interface.
	a.Nil(server.SetSchedulePreset(1), "this method must not return an error during testing")
	a.Equal(1, fake.SimpleReqCallCount(), "this method must call simpleReq() exactly once")
	cmd, params, cameraNum := fake.SimpleReqArgsForCall(0) // check the web request parameters
	a.Equal(url.Values{"id": []string{"1"}}, params, "the presetID sent to securityspy was incorrect")
	a.Equal("++ssSetPreset", cmd, "the wrong command was used to invoke a securityspy schedule preset")
	a.Equal(-1, cameraNum, "this method does not use a camera and must be -1")
}
