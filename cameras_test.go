package securityspy_test

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golift.io/securityspy/v2"
)

func TestUnmarshalXMLCameraSchedule(t *testing.T) {
	t.Parallel()
	asert := assert.New(t)

	var s securityspy.CameraSchedule

	err := xml.Unmarshal([]byte("<tag>3</tag>"), &s)
	require.NoError(t, err, "valid data must not produce an error")
	asert.Equal(3, s.ID, "the data was not unmarshalled properly")
}

// SS 5.5 keeps v5 width/height tags but also emits video-format (a v6-era field).
func TestUnmarshalXMLCameraSS55WithVideoFormat(t *testing.T) {
	t.Parallel()

	const camXML = `<camera>
		<number>1</number>
		<connected>yes</connected>
		<width>3072</width>
		<height>1728</height>
		<mode-c>armed</mode-c>
		<mode-m>armed</mode-m>
		<mode-a>armed</mode-a>
		<hasaudio>yes</hasaudio>
		<name>Mailbox</name>
		<devicename>Dahua Technology</devicename>
		<video-format>H.265</video-format>
		<audio-format>AAC</audio-format>
	</camera>`

	var cam securityspy.Camera
	require.NoError(t, xml.Unmarshal([]byte(camXML), &cam))
	require.Equal(t, "Mailbox", cam.Name)
	require.Equal(t, 3072, cam.Width)
	require.Equal(t, 1728, cam.Height)
	require.Equal(t, "H.265", cam.VideoFormat)
	require.Equal(t, "AAC", cam.AudioFormat)
	require.True(t, cam.ModeM.Val)
	require.True(t, cam.HasAudio.Val)
	require.Equal(t, "Dahua Technology", cam.DeviceName)
	require.Equal(t, "h265", cam.PreferredVCodec())
}

func TestAll(t *testing.T) {
	t.Parallel()
	asert := assert.New(t)

	secspyServer, _, _ := testServerWithCamera(t)

	cams := secspyServer.Cameras.All()
	asert.Len(cams, 2, "the data contains two cameras, two cameras must be returned")
}

func TestByNum(t *testing.T) {
	t.Parallel()
	asert := assert.New(t)

	secspyServer, _, _ := testServerWithCamera(t)

	cam := secspyServer.Cameras.ByNum(2)
	asert.Equal("Porch", cam.Name, "camera 2 is Porch in the v6 test data")
	require.Nil(t, secspyServer.Cameras.ByNum(99), "a non-existent camera must return nil")
}

func TestByName(t *testing.T) {
	t.Parallel()
	asert := assert.New(t)

	secspyServer, _, _ := testServerWithCamera(t)

	cam := secspyServer.Cameras.ByName("Porch")
	asert.Equal(2, cam.Number, "camera 2 is Porch in the v6 test data")
	require.Nil(t, secspyServer.Cameras.ByName("not here"), "a non-existent camera must return nil")

	cam = secspyServer.Cameras.ByName("porch2")
	require.Nil(t, cam, "there is no camera named porch2")

	cam = secspyServer.Cameras.ByName("porch")
	asert.Equal(2, cam.Number, "camera 2 is Porch in the v6 test data")
	require.Nil(t, secspyServer.Cameras.ByName("not here"), "a non-existent camera must return nil")
}

/* Having a comment at the end of the file like this allows commenting the whole file easily. */
