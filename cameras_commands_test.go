package securityspy_test

import (
	"encoding/xml"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golift.io/securityspy"
	"golift.io/securityspy/mocks"
	"golift.io/securityspy/server"
)

func testServerWithCamera(t *testing.T) (*securityspy.Server, *mocks.MockAPI, *securityspy.Camera) {
	t.Helper()

	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)

	secspyServer := securityspy.NewMust(
		&server.Config{Username: "user", Password: "pass", URL: "http://127.0.0.1:5678", VerifySSL: false})
	fake := mocks.NewMockAPI(mockCtrl)
	secspyServer.API = fake

	fake.EXPECT().GetXML(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Do(
		func(_, _, v any) {
			_ = xml.Unmarshal([]byte(testSystemInfo), &v)
		},
	)
	require.NoError(t, secspyServer.Refresh())

	camera := secspyServer.Cameras.ByNum(1)
	require.NotNil(t, camera)

	return secspyServer, fake, camera
}

func TestToggleContinuousUsesNumericArmValues(t *testing.T) {
	t.Parallel()

	_, fake, camera := testServerWithCamera(t)

	fake.EXPECT().SimpleReq("++ssControlContinuous", url.Values{"arm": []string{"0"}}, camera.Number)
	require.NoError(t, camera.ToggleContinuous(securityspy.CameraDisarm))

	fake.EXPECT().SimpleReq("++ssControlContinuous", url.Values{"arm": []string{"1"}}, camera.Number)
	require.NoError(t, camera.ToggleContinuous(securityspy.CameraArm))
}

func TestToggleMotionUsesNumericArmValues(t *testing.T) {
	t.Parallel()

	_, fake, camera := testServerWithCamera(t)

	fake.EXPECT().SimpleReq("++ssControlMotionCapture", url.Values{"arm": []string{"1"}}, camera.Number)
	require.NoError(t, camera.ToggleMotion(securityspy.CameraArm))
}

func TestToggleActionsUsesNumericArmValues(t *testing.T) {
	t.Parallel()

	_, fake, camera := testServerWithCamera(t)

	fake.EXPECT().SimpleReq("++ssControlActions", url.Values{"arm": []string{"0"}}, camera.Number)
	require.NoError(t, camera.ToggleActions(securityspy.CameraDisarm))
}
