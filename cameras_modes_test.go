package securityspy_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCameraModes(t *testing.T) {
	t.Parallel()

	serverObj := newTestServer(t, func(resp http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case systemInfoPath:
			resp.Header().Set("Content-Type", "application/xml")
			_, _ = resp.Write([]byte(testSystemInfo))
		case "/++cameramodes":
			_, _ = resp.Write([]byte("C:DISARMED\rM:ARMED\rA:ARMED\r"))
		default:
			http.NotFound(resp, req)
		}
	})

	require.NoError(t, serverObj.Refresh())

	camera := serverObj.Cameras.ByNum(1)
	require.NotNil(t, camera)

	modes, err := camera.Modes()
	require.NoError(t, err)
	require.Equal(t, "DISARMED", modes.Continuous)
	require.Equal(t, "ARMED", modes.Motion)
	require.Equal(t, "ARMED", modes.Actions)
}

func TestCameraHLSURL(t *testing.T) {
	t.Parallel()

	_, _, camera := testServerWithCamera(t)

	url := camera.HLSURL()
	require.Contains(t, url, "++hls?")
	require.Contains(t, url, "cameraNum=1")
	require.Contains(t, url, "auth=")
}
