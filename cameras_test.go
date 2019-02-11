package securityspy

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalXMLCameraSchedule(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	var s CameraSchedule
	err := xml.Unmarshal([]byte("<tag>3</tag>"), &s)
	assert.Nil(err, "valid data must not produce an error")
	assert.Equal(3, s.ID, "the data was not unmarshalled properly")
}

func TestAll(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	server, _ := GetServer("user", "pass", "http://127.0.0.1:5678", true)
	fake := &fakeAPI{}
	fake.SecReqXMLReturns([]byte(testSystemInfo), nil) // Pass in a test XML payload.
	server.api = fake
	assert.Nil(server.Refresh(), "there must no error when loading fake data") // load the fake testSystemInfo data.
	cams := server.Cameras.All()
	assert.EqualValues(2, len(cams), "the data contains two cameras, two cameras must be returned")
}

func TestByNum(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	server, _ := GetServer("user", "pass", "http://127.0.0.1:5678", true)
	fake := &fakeAPI{}
	fake.SecReqXMLReturns([]byte(testSystemInfo), nil) // Pass in a test XML payload.
	server.api = fake
	assert.Nil(server.Refresh(), "there must no error when loading fake data") // load the fake testSystemInfo data.
	cam := server.Cameras.ByNum(1)
	assert.EqualValues("Porch", cam.Name, "camera 1 is Porch in the test data")
	assert.Nil(server.Cameras.ByNum(99), "a non-existant camera must return nil")
}

func TestByName(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	server, _ := GetServer("user", "pass", "http://127.0.0.1:5678", true)
	fake := &fakeAPI{}
	fake.SecReqXMLReturns([]byte(testSystemInfo), nil) // Pass in a test XML payload.
	server.api = fake
	assert.Nil(server.Refresh(), "there must no error when loading fake data") // load the fake testSystemInfo data.
	cam := server.Cameras.ByName("Porch")
	assert.EqualValues(1, cam.Number, "camera 1 is Porch in the test data")
	assert.Nil(server.Cameras.ByName("not here"), "a non-existant camera must return nil")
}

/*


















/**/
