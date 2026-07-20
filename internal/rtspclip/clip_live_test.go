//go:build live

package rtspclip_test

import (
	"context"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golift.io/securityspy/v2/internal/rtspclip"
)

// Live spike against a SecuritySpy server.
//
//	SECURITYSPY_URL=https://host:8001 \
//	SECURITYSPY_USER=user SECURITYSPY_PASS=pass \
//	go test -tags=live ./internal/rtspclip/ -run TestLiveSaveMP4 -v -count=1
func TestLiveSaveMP4(t *testing.T) {
	baseHTTP := os.Getenv("SECURITYSPY_URL")
	user := os.Getenv("SECURITYSPY_USER")
	pass := os.Getenv("SECURITYSPY_PASS")
	if baseHTTP == "" || user == "" || pass == "" {
		t.Skip("set SECURITYSPY_URL, SECURITYSPY_USER, SECURITYSPY_PASS")
	}

	cam := os.Getenv("SECURITYSPY_CAMERA")
	if cam == "" {
		cam = "3"
	}

	rtspURL, err := buildRTSPURL(baseHTTP, user, pass, cam)
	require.NoError(t, err)
	t.Logf("rtsp url host/path (user redacted): %s", redactUserinfo(rtspURL))

	out := filepath.Join(t.TempDir(), "spike.mp4")
	ctx := context.Background()

	dur := 4 * time.Second
	if v := os.Getenv("SECURITYSPY_DURATION"); v != "" {
		parsed, err := time.ParseDuration(v)
		require.NoError(t, err)
		dur = parsed
	}

	maxBytes := int64(2_500_000)
	if v := os.Getenv("SECURITYSPY_MAXBYTES"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		require.NoError(t, err)
		maxBytes = n
	}

	res, err := rtspclip.SaveMP4(ctx, rtspURL, out, rtspclip.Options{
		Duration:           dur,
		MaxBytes:           maxBytes,
		InsecureSkipVerify: true,
	})
	require.NoError(t, err)
	require.Greater(t, res.Frames, 0)
	t.Logf("result: frames=%d bytes=%d duration=%s hasAudio=%v", res.Frames, res.Bytes, res.Duration, res.HasAudio)

	st, err := os.Stat(out)
	require.NoError(t, err)
	require.Greater(t, st.Size(), int64(1000))
	t.Logf("wrote %s (%d bytes)", out, st.Size())

	// Keep a copy for manual inspection when SPIKE_OUT is set.
	if dest := os.Getenv("SPIKE_OUT"); dest != "" {
		data, err := os.ReadFile(out)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(dest, data, 0o644))
		t.Logf("copied to %s", dest)
	}

	if _, err := exec.LookPath("ffprobe"); err == nil {
		cmd := exec.Command("ffprobe", "-v", "error", "-show_entries",
			"format=format_name,duration:stream=codec_name,codec_type",
			"-of", "default=noprint_wrappers=1", out)
		b, err := cmd.CombinedOutput()
		t.Logf("ffprobe:\n%s", b)
		require.NoError(t, err, string(b))
	}
}

func buildRTSPURL(httpBase, user, pass, camera string) (string, error) {
	u, err := url.Parse(httpBase)
	if err != nil {
		return "", err
	}
	scheme := "rtsp"
	if u.Scheme == "https" {
		scheme = "rtsps"
	}
	q := url.Values{}
	q.Set("cameraNum", camera)
	q.Set("vcodec", "h264")
	q.Set("acodec", "aac")
	if h := os.Getenv("SECURITYSPY_HEIGHT"); h != "" {
		q.Set("height", h)
	}

	// Query auth= is rejected by SecuritySpy RTSP (401). Use userinfo.
	out := &url.URL{
		Scheme:   scheme,
		User:     url.UserPassword(user, pass),
		Host:     u.Host,
		Path:     "/stream",
		RawQuery: q.Encode(),
	}
	return out.String(), nil
}

func redactUserinfo(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	u.User = url.UserPassword("USER", "REDACTED")
	return u.String()
}
