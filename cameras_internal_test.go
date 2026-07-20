package securityspy

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golift.io/securityspy/v2/server"
)

func TestGetJPEGRetriesAndSucceeds(t *testing.T) {
	t.Parallel()

	const retryAttempts = 4

	const badAttempts = retryAttempts - 1

	var (
		requests int
		jpegData bytes.Buffer
	)

	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	require.NoError(t, jpeg.Encode(&jpegData, img, nil))

	fakeServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		requests++
		if requests <= badAttempts {
			_, _ = writer.Write([]byte("not-a-jpeg"))

			return
		}

		writer.Header().Set("Content-Type", "image/jpeg")
		_, _ = writer.Write(jpegData.Bytes())
	}))
	defer fakeServer.Close()

	srv := NewMust(&server.Config{
		URL:         fakeServer.URL + "/",
		Timeout:     server.Duration{Duration: time.Second},
		JPEGRetries: retryAttempts,
	})

	camera := &Camera{Number: 2, server: srv}

	got, err := camera.GetJPEG(nil)
	require.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, retryAttempts, requests)
}

func TestMakeVideoURLUserinfoAndCodecs(t *testing.T) {
	t.Parallel()

	srv := NewMust(&server.Config{
		URL:      "https://ss.example:8001/",
		Username: "admin",
		Password: "s3cret",
		Timeout:  server.Duration{Duration: time.Second},
	})
	cam := &Camera{Number: 3, Name: "Office", server: srv}

	raw, err := cam.makeVideoURL(&VidOps{Height: 720, ACodec: "aac", Quality: 20},
		cam.makeRequestParams(&VidOps{Height: 720, Quality: 20}))
	require.NoError(t, err)
	require.Contains(t, raw, "rtsps://admin:s3cret@ss.example:8001/stream?")
	require.Contains(t, raw, "cameraNum=3")
	require.Contains(t, raw, "vcodec=h264")
	require.Contains(t, raw, "acodec=aac")
	require.NotContains(t, raw, "auth=")
	require.NotContains(t, raw, "quality=")
}

func TestMakeVideoURLRejectsUseHTTP(t *testing.T) {
	t.Parallel()

	srv := NewMust(&server.Config{
		URL:     "http://ss.example:8000/",
		Timeout: server.Duration{Duration: time.Second},
	})
	cam := &Camera{Number: 1, server: srv}

	_, err := cam.makeVideoURL(&VidOps{UseHTTP: true}, cam.makeRequestParams(nil))
	require.ErrorIs(t, err, ErrHTTPVideoUnsupported)

	_, err = cam.StreamVideo(&VidOps{UseHTTP: true}, time.Second, 0)
	require.ErrorIs(t, err, ErrHTTPVideoUnsupported)

	err = cam.SaveVideo(&VidOps{UseHTTP: true}, time.Second, 0, t.TempDir()+"/x.mp4")
	require.ErrorIs(t, err, ErrHTTPVideoUnsupported)
}
