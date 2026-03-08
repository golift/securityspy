package securityspy_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golift.io/securityspy/v2"
)

func TestGetFileValidName(t *testing.T) {
	t.Parallel()

	secspyServer, _, _ := testServerWithCamera(t)
	file, err := secspyServer.Files.GetFile("01-18-2019 10-17-53 M Porch.m4v")
	require.NoError(t, err)
	require.Equal(t, "video/quicktime", file.Link.Type)
	require.Equal(t, 1, file.CameraNum)
	require.Contains(t, file.Link.HREF, "++getfile/1/2019-01-18/")
}

func TestGetFileInvalidNames(t *testing.T) {
	t.Parallel()

	secspyServer, _, _ := testServerWithCamera(t)

	_, err := secspyServer.Files.GetFile("not-a-real-file-name")
	require.ErrorIs(t, err, securityspy.ErrInvalidName)

	_, err = secspyServer.Files.GetFile("01-18-2019 10-17-53 M MissingCamera.m4v")
	require.ErrorIs(t, err, securityspy.ErrCAMMissing)
}

func TestGetFileJPGType(t *testing.T) {
	t.Parallel()

	secspyServer, _, _ := testServerWithCamera(t)
	file, err := secspyServer.Files.GetFile("01-18-2019 10-17-53 M Porch.jpg")
	require.NoError(t, err)
	require.Equal(t, "image/jpeg", file.Link.Type)
}

func TestFileDateFormatParse(t *testing.T) {
	t.Parallel()

	_, err := time.Parse(securityspy.FileDateFormat, "01-18-2019")
	require.NoError(t, err)
}
