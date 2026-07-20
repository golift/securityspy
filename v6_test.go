package securityspy_test

import (
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"golift.io/securityspy/v2"
)

func TestRefreshV6SystemInfo(t *testing.T) {
	t.Parallel()

	fixture, err := os.ReadFile("testdata/systemInfo-v6.xml")
	require.NoError(t, err)

	serverObj := newTestServer(t, func(resp http.ResponseWriter, _ *http.Request) {
		resp.Header().Set("Content-Type", "application/xml")
		_, _ = resp.Write(fixture)
	})

	require.NoError(t, serverObj.Refresh())
	require.Equal(t, "6.20", serverObj.Info.Version)
	require.Equal(t, 11, serverObj.Info.CameraCount)
	require.Len(t, serverObj.Cameras.All(), 2)
	require.Len(t, serverObj.Groups, 2)
	require.Equal(t, "Base", serverObj.Groups[0].Name)
	require.Equal(t, []int{2, 5, 6, 11, 12}, serverObj.Groups[0].CameraNumbers())

	door := serverObj.Cameras.ByNum(3)
	require.NotNil(t, door)
	require.Equal(t, "Door", door.Name)
	require.Equal(t, 3072, door.Width)
	require.Equal(t, 2048, door.Height)
	require.True(t, door.Connected.Val)
	require.True(t, door.ModeM.Val)
	require.False(t, door.ModeC.Val)
	require.True(t, door.HasAudio.Val)
	require.Equal(t, "/Cameras/Door", door.CapturePath)
	require.Equal(t, "H.264", door.VideoFormat)
	require.NotNil(t, door.PTZ)
	require.True(t, door.PTZ.HasPanTilt)
	require.True(t, door.PTZ.Continuous)
	require.Equal(t, "Armed 24/7", door.ScheduleIDMC.Name)
	require.Equal(t, "Disarmed 24/7", door.ScheduleIDCC.Name)
}

func TestRefreshV5SystemInfoStillWorks(t *testing.T) {
	t.Parallel()

	serverObj := newTestServer(t, func(resp http.ResponseWriter, _ *http.Request) {
		resp.Header().Set("Content-Type", "application/xml")
		_, _ = resp.Write([]byte(testSystemInfo))
	})

	require.NoError(t, serverObj.Refresh())
	require.Equal(t, "4.2.10", serverObj.Info.Version)

	cam := serverObj.Cameras.ByNum(1)
	require.NotNil(t, cam)
	require.Equal(t, 2304, cam.Width)
	require.True(t, cam.ModeC.Val)
	require.Equal(t, "Porch", cam.Name)
}

func TestSettingsGetAndSet(t *testing.T) {
	t.Parallel()

	generalXML, err := os.ReadFile(".archive/settings-general-v6.20.xml")
	require.NoError(t, err)

	var posted url.Values

	serverObj := newTestServer(t, func(resp http.ResponseWriter, req *http.Request) {
		switch {
		case req.URL.Path == systemInfoPath:
			resp.Header().Set("Content-Type", "application/xml")
			_, _ = resp.Write([]byte(testSystemInfoV6))
		case req.Method == http.MethodGet && req.URL.Path == "/++settings-general":
			resp.Header().Set("Content-Type", "application/xml")
			_, _ = resp.Write(generalXML)
		case req.Method == http.MethodPost && req.URL.Path == "/++settings-general":
			_ = req.ParseForm()
			posted = req.PostForm
			_, _ = resp.Write([]byte(`{"result":"OK"}`))
		default:
			http.NotFound(resp, req)
		}
	})

	require.NoError(t, serverObj.Refresh())

	general, err := serverObj.GetGeneralSettings()
	require.NoError(t, err)
	require.Equal(t, "Example Server", general.SysName)
	require.True(t, general.AutoReopen.Val)

	require.NoError(t, serverObj.SetGeneralSettings(url.Values{"sysName": {"Example Server"}}))
	require.Equal(t, "Example Server", posted.Get("sysName"))
}

func TestMediaURLHelpers(t *testing.T) {
	t.Parallel()

	secspyServer, _, camera := testServerWithCamera(t)

	require.Contains(t, camera.HLSMediaPlaylistURL(2), "++hls_mediaplaylist?")
	require.Contains(t, camera.HLSMediaPlaylistURL(2), "quality=2")
	require.Contains(t, camera.LiveURL(), "++live?")

	mux := secspyServer.MultiplexURL(&securityspy.MultiplexOps{
		Cameras:  []int{3, 0},
		CropMode: 1,
		Format:   0,
		FPS:      2,
		CamInfo:  true,
	})
	require.Contains(t, mux, "++multiplex?")
	require.Contains(t, mux, "cameras=3%2C0")
	require.Contains(t, mux, "cropMode=1")
	require.Contains(t, mux, "camInfo=1")
}
