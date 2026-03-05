package securityspy_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golift.io/securityspy"
)

func TestUnmarshalEventWithoutInfo(t *testing.T) {
	t.Parallel()

	secspyServer, _, _ := testServerWithCamera(t)
	event := secspyServer.Events.UnmarshalEvent("20190927091955 1 1 ARM_C")

	require.Equal(t, 1, event.ID)
	require.Equal(t, securityspy.EventArmContinuous, event.Type)
	require.Equal(t, "ARM_C", event.Msg)
	require.Empty(t, event.Errors)
	require.NotNil(t, event.Camera)
}

func TestUnmarshalEventShortPayload(t *testing.T) {
	t.Parallel()

	secspyServer, _, _ := testServerWithCamera(t)
	event := secspyServer.Events.UnmarshalEvent("bad event")

	require.Equal(t, securityspy.EventUnknownEvent, event.Type)
	require.NotEmpty(t, event.Errors)
}

func TestUnmarshalEventTriggerReasons(t *testing.T) {
	t.Parallel()

	secspyServer, _, _ := testServerWithCamera(t)
	event := secspyServer.Events.UnmarshalEvent("20190927092026 5 1 TRIGGER_M 3")

	require.Equal(t, securityspy.EventTriggerMotion, event.Type)
	require.Contains(t, event.Msg, "Motion Detected")
	require.Contains(t, event.Msg, "Audio Detected")
	require.Len(t, event.Reasons, 2)
}

func TestUnmarshalEventMotionEnd(t *testing.T) {
	t.Parallel()

	secspyServer, _, _ := testServerWithCamera(t)
	event := secspyServer.Events.UnmarshalEvent("20190927092040 9 1 MOTION_END")

	require.Equal(t, 9, event.ID)
	require.Equal(t, securityspy.EventMotionEnd, event.Type)
	require.Equal(t, "MOTION_END", event.Msg)
	require.Empty(t, event.Errors)
	require.NotNil(t, event.Camera)
}

func TestUnmarshalEventTriggerReasonsNewFlags(t *testing.T) {
	t.Parallel()

	secspyServer, _, _ := testServerWithCamera(t)
	event := secspyServer.Events.UnmarshalEvent("20190927092026 6 1 TRIGGER_A 12288")

	require.Equal(t, securityspy.EventTriggerAction, event.Type)
	require.Contains(t, event.Msg, "Human Departure")
	require.Contains(t, event.Msg, "Vehicle Arrival")
	require.Len(t, event.Reasons, 2)
}
