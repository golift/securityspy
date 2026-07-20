package securityspy_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"golift.io/securityspy/v2"
)

func TestToggleContinuousUsesNumericArmValues(t *testing.T) {
	t.Parallel()

	_, recorder, camera := testServerWithCamera(t)

	require.NoError(t, camera.ToggleContinuous(securityspy.CameraDisarm))

	req, found := recorder.findLast("/++ssControlContinuous")
	require.True(t, found)
	require.Equal(t, "0", req.Query.Get("arm"))
	require.Equal(t, "3", req.Query.Get("cameraNum"))

	require.NoError(t, camera.ToggleContinuous(securityspy.CameraArm))

	req, found = recorder.findLast("/++ssControlContinuous")
	require.True(t, found)
	require.Equal(t, "1", req.Query.Get("arm"))
	require.Equal(t, "3", req.Query.Get("cameraNum"))
}

func TestToggleContinuousUnsupported(t *testing.T) {
	t.Parallel()

	recorder := &requestRecorder{}
	serverObj := newTestServer(t, func(resp http.ResponseWriter, req *http.Request) {
		recorder.add(req)

		switch req.URL.Path {
		case systemInfoPath:
			resp.Header().Set("Content-Type", "application/xml")
			_, _ = resp.Write([]byte(testSystemInfoV6))
		case "/++ssControlContinuous":
			http.NotFound(resp, req)
		default:
			http.NotFound(resp, req)
		}
	})

	require.NoError(t, serverObj.Refresh())

	camera := serverObj.Cameras.ByNum(3)
	require.NotNil(t, camera)

	err := camera.ToggleContinuous(securityspy.CameraArm)
	require.ErrorIs(t, err, securityspy.ErrUnsupported)
}

func TestToggleMotionUsesNumericArmValues(t *testing.T) {
	t.Parallel()

	_, recorder, camera := testServerWithCamera(t)

	require.NoError(t, camera.ToggleMotion(securityspy.CameraArm))

	req, found := recorder.findLast("/++ssControlMotionCapture")
	require.True(t, found)
	require.Equal(t, "1", req.Query.Get("arm"))
	require.Equal(t, "3", req.Query.Get("cameraNum"))
}

func TestToggleActionsUsesNumericArmValues(t *testing.T) {
	t.Parallel()

	_, recorder, camera := testServerWithCamera(t)

	require.NoError(t, camera.ToggleActions(securityspy.CameraDisarm))

	req, found := recorder.findLast("/++ssControlActions")
	require.True(t, found)
	require.Equal(t, "0", req.Query.Get("arm"))
	require.Equal(t, "3", req.Query.Get("cameraNum"))
}
