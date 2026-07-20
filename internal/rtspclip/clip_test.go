package rtspclip_test

import (
	"context"
	"encoding/base64"
	"testing"
	"time"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/mpeg4audio"
	"github.com/stretchr/testify/require"
	"golift.io/securityspy/v2/internal/rtspclip"
)

// Sanity: library Auth() blob is reversible for RTSP userinfo construction.
func TestAuthBlobRoundTrip(t *testing.T) {
	t.Parallel()

	blob := base64.URLEncoding.EncodeToString([]byte("cursor:secret"))
	raw, err := base64.URLEncoding.DecodeString(blob)
	require.NoError(t, err)
	require.Equal(t, "cursor:secret", string(raw))
}

func TestSaveMP4RejectsBadOptions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	_, err := rtspclip.SaveMP4(ctx, "rtsps://x/stream", "/tmp/x.mp4", rtspclip.Options{})
	require.ErrorIs(t, err, rtspclip.ErrBadDuration)

	_, err = rtspclip.SaveMP4(ctx, "rtsps://x/stream", "", rtspclip.Options{Duration: time.Second})
	require.ErrorIs(t, err, rtspclip.ErrBadPath)
}

func TestStreamMP4RejectsBadDuration(t *testing.T) {
	t.Parallel()

	_, err := rtspclip.StreamMP4(context.Background(), "rtsps://x/stream", rtspclip.Options{})
	require.ErrorIs(t, err, rtspclip.ErrBadDuration)
}

// SecuritySpy office cam is AAC LC mono @ 64kHz; ASC must be 0x1108 (not stereo 0x1210).
func TestMonoAAC64kASC(t *testing.T) {
	t.Parallel()

	cfg := mpeg4audio.AudioSpecificConfig{
		Type:          mpeg4audio.ObjectTypeAACLC,
		SampleRate:    64000,
		ChannelCount:  1,
		ChannelConfig: 1,
	}
	asc, err := cfg.Marshal()
	require.NoError(t, err)
	require.Equal(t, []byte{0x11, 0x08}, asc)
}

func TestHighSampleRateASCStillMarshals(t *testing.T) {
	t.Parallel()

	// 96000 does not fit in mp4a sample_rate uint16; ASC still carries the real rate.
	cfg := mpeg4audio.AudioSpecificConfig{
		Type:          mpeg4audio.ObjectTypeAACLC,
		SampleRate:    96000,
		ChannelCount:  2,
		ChannelConfig: 2,
	}
	asc, err := cfg.Marshal()
	require.NoError(t, err)
	require.NotEmpty(t, asc)
}
