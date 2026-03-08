package securityspy_test

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/require"
	"golift.io/securityspy/v2"
)

func TestPTZPresetCommandMapping(t *testing.T) {
	t.Parallel()

	_, recorder, camera := testServerWithCamera(t)
	require.NotNil(t, camera.PTZ)

	require.NoError(t, camera.PTZ.Preset(securityspy.PTZpreset1))

	req, found := recorder.findLast("/++ptz/command")
	require.True(t, found)
	require.Equal(t, "12", req.Query.Get("command"))
	require.Equal(t, "1", req.Query.Get("cameraNum"))

	require.NoError(t, camera.PTZ.PresetSave(securityspy.PTZpreset1))

	req, found = recorder.findLast("/++ptz/command")
	require.True(t, found)
	require.Equal(t, "112", req.Query.Get("command"))
	require.Equal(t, "1", req.Query.Get("cameraNum"))
}

func TestPTZCapabilitiesSpeedAndContinuous(t *testing.T) {
	t.Parallel()

	var withSpeed struct {
		PTZ securityspy.PTZ `xml:"ptzcapabilities"`
	}

	require.NoError(t, xml.Unmarshal([]byte("<root><ptzcapabilities>16</ptzcapabilities></root>"), &withSpeed))
	require.True(t, withSpeed.PTZ.HasSpeed)
	require.False(t, withSpeed.PTZ.Continuous)

	var withContinuous struct {
		PTZ securityspy.PTZ `xml:"ptzcapabilities"`
	}

	require.NoError(t, xml.Unmarshal([]byte("<root><ptzcapabilities>32</ptzcapabilities></root>"), &withContinuous))
	require.False(t, withContinuous.PTZ.HasSpeed)
	require.True(t, withContinuous.PTZ.Continuous)
}
