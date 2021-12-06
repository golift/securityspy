package securityspy_test

import (
	"encoding/xml"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golift.io/securityspy"
	"golift.io/securityspy/mocks"
	"golift.io/securityspy/server"
)

func TestUnmarshalXMLCameraSchedule(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	var s securityspy.CameraSchedule

	err := xml.Unmarshal([]byte("<tag>3</tag>"), &s)
	assert.Nil(err, "valid data must not produce an error")
	assert.Equal(3, s.ID, "the data was not unmarshalled properly")
}

func TestAll(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	server := securityspy.NewMust(
		&server.Config{Username: "user", Password: "pass", URL: "http://127.0.0.1:5678", VerifySSL: true})
	fake := mocks.NewMockAPI(mockCtrl)
	server.API = fake

	fake.EXPECT().GetXML(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Do(
		func(_, _, v interface{}) {
			_ = xml.Unmarshal([]byte(testSystemInfo), &v)
		},
	)
	assert.Nil(server.Refresh(), "there must no error when loading fake data") // load the fake testSystemInfo data.

	cams := server.Cameras.All()
	assert.EqualValues(2, len(cams), "the data contains two cameras, two cameras must be returned")
}

func TestByNum(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	server := securityspy.NewMust(
		&server.Config{Username: "user", Password: "pass", URL: "http://127.0.0.1:5678", VerifySSL: true})
	fake := mocks.NewMockAPI(mockCtrl)
	server.API = fake

	fake.EXPECT().GetXML(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Do(
		func(_, _, v interface{}) {
			_ = xml.Unmarshal([]byte(testSystemInfo), &v)
		},
	)
	assert.Nil(server.Refresh(), "there must no error when loading fake data") // load the fake testSystemInfo data.

	cam := server.Cameras.ByNum(1)
	assert.EqualValues("Porch", cam.Name, "camera 1 is Porch in the test data")
	assert.Nil(server.Cameras.ByNum(99), "a non-existent camera must return nil")
}

func TestByName(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	server := securityspy.NewMust(
		&server.Config{Username: "user", Password: "pass", URL: "http://127.0.0.1:5678", VerifySSL: true})
	fake := mocks.NewMockAPI(mockCtrl)
	server.API = fake

	fake.EXPECT().GetXML(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Do(
		func(_, _, v interface{}) {
			_ = xml.Unmarshal([]byte(testSystemInfo), &v)
		},
	)
	assert.Nil(server.Refresh(), "there must no error when loading fake data") // load the fake testSystemInfo data.

	cam := server.Cameras.ByName("Porch")
	assert.EqualValues(1, cam.Number, "camera 1 is Porch in the test data")
	assert.Nil(server.Cameras.ByName("not here"), "a non-existent camera must return nil")

	cam = server.Cameras.ByName("porch2")
	assert.Nil(cam, "there is no camera named porch2")

	cam = server.Cameras.ByName("porch")
	assert.EqualValues(1, cam.Number, "camera 1 is Porch in the test data")
	assert.Nil(server.Cameras.ByName("not here"), "a non-existent camera must return nil")
}

/* Having a comment at the end of the file like this allows commenting the whole file easily. */
