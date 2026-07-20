package securityspy

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golift.io/securityspy/v2/server"
)

func TestGetJPEGSucceeds(t *testing.T) {
	t.Parallel()

	var jpegData bytes.Buffer

	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	require.NoError(t, jpeg.Encode(&jpegData, img, nil))

	fakeServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "image/jpeg")
		_, _ = writer.Write(jpegData.Bytes())
	}))
	defer fakeServer.Close()

	srv := NewMust(&server.Config{
		URL:     fakeServer.URL + "/",
		Timeout: server.Duration{Duration: time.Second},
	})

	camera := &Camera{Number: 2, server: srv}

	got, err := camera.GetJPEG(nil)
	require.NoError(t, err)
	assert.NotNil(t, got)
}

func TestGetJPEGRetriesAndSucceeds(t *testing.T) {
	t.Parallel()

	const retryAttempts = 4

	const badAttempts = retryAttempts - 1

	var (
		requests atomic.Int32
		jpegData bytes.Buffer
	)

	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	require.NoError(t, jpeg.Encode(&jpegData, img, nil))

	fakeServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		n := requests.Add(1)
		if n <= badAttempts {
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
	assert.Equal(t, int32(retryAttempts), requests.Load())
}

func TestGetJPEGRejectsNonJPEG(t *testing.T) {
	t.Parallel()

	fakeServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write([]byte("not-a-jpeg"))
	}))
	defer fakeServer.Close()

	srv := NewMust(&server.Config{
		URL:         fakeServer.URL + "/",
		Timeout:     server.Duration{Duration: time.Second},
		JPEGRetries: 1,
	})
	camera := &Camera{Number: 2, Name: "X", server: srv}

	_, err := camera.GetJPEG(nil)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidJPEG)
}

func TestGetJPEGNoRetryOn404(t *testing.T) {
	t.Parallel()

	var requests atomic.Int32

	fakeServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		writer.WriteHeader(http.StatusNotFound)
		_, _ = writer.Write(bytes.Repeat([]byte("x"), 1024))
	}))
	defer fakeServer.Close()

	srv := NewMust(&server.Config{
		URL:         fakeServer.URL + "/",
		Timeout:     server.Duration{Duration: time.Second},
		JPEGRetries: 5,
	})
	camera := &Camera{Number: 2, Name: "DeadCam", server: srv}

	_, err := camera.GetJPEG(nil)
	require.ErrorIs(t, err, ErrCameraUnavailable)
	assert.Equal(t, int32(1), requests.Load())
}

func TestSaveJPEGNoOverwrite(t *testing.T) {
	t.Parallel()

	var jpegData bytes.Buffer

	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	require.NoError(t, jpeg.Encode(&jpegData, img, nil))

	fakeServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "image/jpeg")
		_, _ = writer.Write(jpegData.Bytes())
	}))
	defer fakeServer.Close()

	srv := NewMust(&server.Config{
		URL:         fakeServer.URL + "/",
		Timeout:     server.Duration{Duration: time.Second},
		JPEGRetries: 1,
	})
	camera := &Camera{Number: 1, server: srv}

	path := t.TempDir() + "/snap.jpg"
	require.NoError(t, os.WriteFile(path, []byte("keep"), 0o600))
	require.ErrorIs(t, camera.SaveJPEG(nil, path), ErrPathExists)

	raw, err := os.ReadFile(path) //nolint:gosec // test temp path
	require.NoError(t, err)
	require.Equal(t, []byte("keep"), raw)
}

// Slow bodies used to fail with "missing SOI marker" / "context canceled" because
// GetJPEG canceled the request context before reading the response body.
func TestGetJPEGReadsBodyBeforeCancel(t *testing.T) {
	t.Parallel()

	var jpegData bytes.Buffer

	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 1, G: 2, B: 3, A: 255})
	require.NoError(t, jpeg.Encode(&jpegData, img, &jpeg.Options{Quality: 90}))

	payload := jpegData.Bytes()

	fakeServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "image/jpeg")
		writer.WriteHeader(http.StatusOK)

		flusher, _ := writer.(http.Flusher)
		// trickle bytes so a premature context cancel would truncate the body
		for i := 0; i < len(payload); i += 8 {
			end := min(i+8, len(payload))

			_, _ = writer.Write(payload[i:end])

			if flusher != nil {
				flusher.Flush()
			}

			time.Sleep(5 * time.Millisecond)
		}
	}))
	defer fakeServer.Close()

	srv := NewMust(&server.Config{
		URL:         fakeServer.URL + "/",
		Timeout:     server.Duration{Duration: 5 * time.Second},
		JPEGRetries: 1,
	})
	camera := &Camera{Number: 1, server: srv}

	got, err := camera.GetJPEG(nil)
	require.NoError(t, err)
	require.NotNil(t, got)

	path := t.TempDir() + "/snap.jpg"
	require.NoError(t, camera.SaveJPEG(nil, path))

	raw, err := os.ReadFile(path) //nolint:gosec // test temp path
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(raw), 2)
	require.Equal(t, byte(0xFF), raw[0])
	require.Equal(t, byte(0xD8), raw[1])
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

	raw, err := cam.makeVideoURL(&VidOps{Height: 720, ACodec: aacStr, Quality: 20},
		cam.makeRequestParams(&VidOps{Height: 720, Quality: 20}))
	require.NoError(t, err)
	require.Contains(t, raw, "rtsps://admin:s3cret@ss.example:8001/stream?")
	// cameraNum/vcodec before acodec — SS ignores vcodec=h265 if acodec leads the query.
	require.Contains(t, raw, "stream?cameraNum=3&vcodec=h264&acodec=aac&height=720")
	require.NotContains(t, raw, "auth=")
	require.NotContains(t, raw, "quality=")

	raw, err = cam.makeVideoURL(&VidOps{Height: 720, VCodec: "h265", ACodec: aacStr},
		cam.makeRequestParams(&VidOps{Height: 720}))
	require.NoError(t, err)
	require.Contains(t, raw, "stream?cameraNum=3&vcodec=h265&acodec=aac&height=720")
}

func TestMakeVideoURLPreservesBasePath(t *testing.T) {
	t.Parallel()

	srv := NewMust(&server.Config{
		URL:      "https://ss.example:8001/securityspy/",
		Username: "admin",
		Password: "s3cret",
		Timeout:  server.Duration{Duration: time.Second},
	})
	cam := &Camera{Number: 3, server: srv}

	raw, err := cam.makeVideoURL(nil, cam.makeRequestParams(nil))
	require.NoError(t, err)
	require.Contains(t, raw, "rtsps://admin:s3cret@ss.example:8001/securityspy/stream?")
}

func TestMakeVideoURLOmitsUserinfoWithoutAuth(t *testing.T) {
	t.Parallel()

	srv := NewMust(&server.Config{
		URL:     "https://ss.example:8001/",
		Timeout: server.Duration{Duration: time.Second},
	})
	cam := &Camera{Number: 1, server: srv}

	raw, err := cam.makeVideoURL(nil, cam.makeRequestParams(nil))
	require.NoError(t, err)
	require.Contains(t, raw, "rtsps://ss.example:8001/stream?")
	require.NotContains(t, raw, "@")
}

func TestPreferredVCodec(t *testing.T) {
	t.Parallel()

	require.Equal(t, "h264", (&Camera{VideoFormat: "H.264"}).PreferredVCodec())
	require.Equal(t, "h265", (&Camera{VideoFormat: "H.265"}).PreferredVCodec())
	require.Equal(t, "h265", (&Camera{VideoFormat: "HEVC"}).PreferredVCodec())
	require.Equal(t, "h264", (&Camera{}).PreferredVCodec())
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
