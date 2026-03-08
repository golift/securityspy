package securityspy_test

import (
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
	require.Equal(t, "1", req.Query.Get("cameraNum"))

	require.NoError(t, camera.ToggleContinuous(securityspy.CameraArm))

	req, found = recorder.findLast("/++ssControlContinuous")
	require.True(t, found)
	require.Equal(t, "1", req.Query.Get("arm"))
	require.Equal(t, "1", req.Query.Get("cameraNum"))
}

func TestToggleMotionUsesNumericArmValues(t *testing.T) {
	t.Parallel()

	_, recorder, camera := testServerWithCamera(t)

	require.NoError(t, camera.ToggleMotion(securityspy.CameraArm))

	req, found := recorder.findLast("/++ssControlMotionCapture")
	require.True(t, found)
	require.Equal(t, "1", req.Query.Get("arm"))
	require.Equal(t, "1", req.Query.Get("cameraNum"))
}

func TestToggleActionsUsesNumericArmValues(t *testing.T) {
	t.Parallel()

	_, recorder, camera := testServerWithCamera(t)

	require.NoError(t, camera.ToggleActions(securityspy.CameraDisarm))

	req, found := recorder.findLast("/++ssControlActions")
	require.True(t, found)
	require.Equal(t, "0", req.Query.Get("arm"))
	require.Equal(t, "1", req.Query.Get("cameraNum"))
}
