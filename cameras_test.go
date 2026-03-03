package securityspy_test

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golift.io/securityspy"
	"golift.io/securityspy/mocks"
	"golift.io/securityspy/server"
)

func TestUnmarshalXMLCameraSchedule(t *testing.T) {
	t.Parallel()
	asert := assert.New(t)

	var s securityspy.CameraSchedule

	err := xml.Unmarshal([]byte("<tag>3</tag>"), &s)
	require.NoError(t, err, "valid data must not produce an error")
	asert.Equal(3, s.ID, "the data was not unmarshalled properly")
}

func TestAll(t *testing.T) {
	t.Parallel()
	asert := assert.New(t)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	secspyServer := securityspy.NewMust(
		&server.Config{Username: "user", Password: "pass", URL: "http://127.0.0.1:5678", VerifySSL: true})
	fake := mocks.NewMockAPI(mockCtrl)
	secspyServer.API = fake

	fake.EXPECT().GetXML(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Do(
		func(_, _, v any) {
			_ = xml.Unmarshal([]byte(testSystemInfo), &v)
		},
	)
	require.NoError(t, secspyServer.Refresh(),
		"there must no error when loading fake data") // load the fake testSystemInfo data.

	cams := secspyServer.Cameras.All()
	asert.Len(cams, 2, "the data contains two cameras, two cameras must be returned")
}

func TestByNum(t *testing.T) {
	t.Parallel()
	asert := assert.New(t)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	secspyServer := securityspy.NewMust(
		&server.Config{Username: "user", Password: "pass", URL: "http://127.0.0.1:5678", VerifySSL: true})
	fake := mocks.NewMockAPI(mockCtrl)
	secspyServer.API = fake

	fake.EXPECT().GetXML(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Do(
		func(_, _, v any) {
			_ = xml.Unmarshal([]byte(testSystemInfo), &v)
		},
	)
	require.NoError(t, secspyServer.Refresh(),
		"there must no error when loading fake data") // load the fake testSystemInfo data.

	cam := secspyServer.Cameras.ByNum(1)
	asert.Equal("Porch", cam.Name, "camera 1 is Porch in the test data")
	require.Nil(t, secspyServer.Cameras.ByNum(99), "a non-existent camera must return nil")
}

func TestByName(t *testing.T) {
	t.Parallel()
	asert := assert.New(t)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	secspyServer := securityspy.NewMust(
		&server.Config{Username: "user", Password: "pass", URL: "http://127.0.0.1:5678", VerifySSL: true})
	fake := mocks.NewMockAPI(mockCtrl)
	secspyServer.API = fake

	fake.EXPECT().GetXML(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Do(
		func(_, _, v any) {
			_ = xml.Unmarshal([]byte(testSystemInfo), &v)
		},
	)
	require.NoError(t, secspyServer.Refresh(),
		"there must no error when loading fake data") // load the fake testSystemInfo data.

	cam := secspyServer.Cameras.ByName("Porch")
	asert.Equal(1, cam.Number, "camera 1 is Porch in the test data")
	require.Nil(t, secspyServer.Cameras.ByName("not here"), "a non-existent camera must return nil")

	cam = secspyServer.Cameras.ByName("porch2")
	require.Nil(t, cam, "there is no camera named porch2")

	cam = secspyServer.Cameras.ByName("porch")
	asert.Equal(1, cam.Number, "camera 1 is Porch in the test data")
	require.Nil(t, secspyServer.Cameras.ByName("not here"), "a non-existent camera must return nil")
}

/* Having a comment at the end of the file like this allows commenting the whole file easily. */
