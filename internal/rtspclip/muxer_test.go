package rtspclip //nolint:testpackage // white-box muxer flush behavior

import (
	"bytes"
	"encoding/base64"
	"testing"

	"github.com/Eyevinn/mp4ff/mp4"
	"github.com/stretchr/testify/require"
)

func TestProgressiveMuxMultipleFragments(t *testing.T) {
	t.Parallel()

	sps, err := base64.StdEncoding.DecodeString("Z2QAH6wTFsBQBbsBbdgYAC7gAAu4L3vg+EQjcA==")
	require.NoError(t, err)

	pps, err := base64.StdEncoding.DecodeString("aO4fLA==")
	require.NoError(t, err)

	var buf bytes.Buffer

	mux := newFMP4Muxer(&buf, codecH264, nil, sps, pps, nil, true)

	mux.mu.Lock()
	mux.started = true

	const totalFrames = fragmentFrames*2 + 5

	for range totalFrames {
		data := []byte{0x00, 0x00, 0x00, 0x01, 0x65, 0x00, 0x01, 0x02}
		mux.videoSamples = append(mux.videoSamples, mp4.FullSample{
			Sample: mp4.Sample{
				Flags: mp4.SyncSampleFlags,
				Dur:   defaultFrameTicks,
				Size:  uint32(len(data)), //nolint:gosec // test fixture
			},
			DecodeTime: mux.videoDecode,
			Data:       data,
		})
		mux.videoDecode += uint64(defaultFrameTicks)

		require.NoError(t, mux.maybeInitProgressive())
		require.True(t, mux.initWritten)
		require.NoError(t, mux.maybeFlush(false))
	}

	require.NoError(t, mux.maybeFlush(true))
	mux.mu.Unlock()

	out := buf.Bytes()
	require.Contains(t, string(out), "moov")

	moofCount := bytes.Count(out, []byte("moof"))
	require.GreaterOrEqual(t, moofCount, 2, "expected multiple media fragments, got %d", moofCount)
}

func TestSaveMuxSingleFragmentAtClose(t *testing.T) {
	t.Parallel()

	sps, err := base64.StdEncoding.DecodeString("Z2QAH6wTFsBQBbsBbdgYAC7gAAu4L3vg+EQjcA==")
	require.NoError(t, err)

	pps, err := base64.StdEncoding.DecodeString("aO4fLA==")
	require.NoError(t, err)

	var buf bytes.Buffer

	mux := newFMP4Muxer(&buf, codecH264, nil, sps, pps, nil, false)

	mux.mu.Lock()
	mux.started = true

	for range 20 {
		data := []byte{0x00, 0x00, 0x00, 0x01, 0x65}
		mux.videoSamples = append(mux.videoSamples, mp4.FullSample{
			Sample: mp4.Sample{
				Flags: mp4.SyncSampleFlags,
				Dur:   defaultFrameTicks,
				Size:  uint32(len(data)), //nolint:gosec // test fixture
			},
			DecodeTime: mux.videoDecode,
			Data:       data,
		})
		mux.videoDecode += uint64(defaultFrameTicks)
	}
	mux.mu.Unlock()

	require.Empty(t, buf.Bytes(), "save mux must not write before close")
	require.NoError(t, mux.close())

	out := buf.Bytes()
	require.Contains(t, string(out), "moov")
	require.Equal(t, 1, bytes.Count(out, []byte("moof")))
}
