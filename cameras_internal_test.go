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
