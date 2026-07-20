// Package rtspclip captures short H.264 (+ optional AAC) clips from RTSP/RTSPS.
package rtspclip

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/Eyevinn/mp4ff/mp4"
	"github.com/bluenviron/gortsplib/v5"
	"github.com/bluenviron/gortsplib/v5/pkg/base"
	"github.com/bluenviron/gortsplib/v5/pkg/description"
	"github.com/bluenviron/gortsplib/v5/pkg/format"
	"github.com/bluenviron/gortsplib/v5/pkg/format/rtph264"
	"github.com/bluenviron/gortsplib/v5/pkg/format/rtpmpeg4audio"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h264"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/mpeg4audio"
	"github.com/pion/rtp"
)

// Static errors.
var (
	ErrBadDuration = errors.New("rtspclip: Duration must be > 0")
	ErrBadPath     = errors.New("rtspclip: path required")
	ErrNoH264      = errors.New("rtspclip: no H264 media in DESCRIBE")
	ErrNoFrames    = errors.New("rtspclip: no video frames captured")
	ErrNoParameter = errors.New("rtspclip: missing SPS/PPS at IDR")
)

const (
	clockRateH264         = 90000
	defaultFrameTicks     = 3000 // ~30fps in 90kHz
	nalTypeMask           = 0x1F
	avccLengthSize        = 4
	videoTrackID          = 1
	audioTrackID          = 2
	captureTimeoutPadding = 5 * time.Second
	aacBitsPerSample      = 16
)

// Options control a timed RTSP clip capture.
type Options struct {
	Duration           time.Duration // required (>0)
	MaxBytes           int64         // 0 = no limit
	InsecureSkipVerify bool
}

// Result summarizes a successful capture.
type Result struct {
	Frames   int
	Bytes    int64
	Duration time.Duration
	HasAudio bool
}

// SaveMP4 pulls H.264 (and AAC when present) from rtspURL into a fragmented MP4 file.
//
// SecuritySpy RTSP rejects query auth=; use userinfo: rtsps://user:pass@host/stream?...
func SaveMP4(ctx context.Context, rtspURL, path string, opts Options) (*Result, error) {
	if opts.Duration <= 0 {
		return nil, ErrBadDuration
	}

	if path == "" {
		return nil, ErrBadPath
	}

	if _, err := os.Stat(path); err == nil {
		return nil, os.ErrExist
	}

	file, err := os.Create(path) //nolint:gosec // caller-chosen output path
	if err != nil {
		return nil, fmt.Errorf("create output: %w", err)
	}

	defer file.Close()

	res, err := capture(ctx, rtspURL, opts, file)
	if err != nil {
		_ = os.Remove(path)

		return nil, err
	}

	return res, nil
}

// StreamMP4 captures like SaveMP4 but writes to an io.Pipe ReadCloser.
// Close the returned closer when finished; the capture runs in a goroutine.
func StreamMP4(ctx context.Context, rtspURL string, opts Options) (io.ReadCloser, error) {
	if opts.Duration <= 0 {
		return nil, ErrBadDuration
	}

	reader, writer := io.Pipe()

	go func() {
		_, err := capture(ctx, rtspURL, opts, writer)
		if err != nil {
			_ = writer.CloseWithError(err)

			return
		}

		_ = writer.Close()
	}()

	return reader, nil
}

func capture(ctx context.Context, rtspURL string, opts Options, writer io.Writer) (*Result, error) {
	parsed, err := base.ParseURL(rtspURL)
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}

	client := newRTSPClient(parsed, opts.InsecureSkipVerify)

	if err := client.Start(); err != nil {
		return nil, fmt.Errorf("rtsp start: %w", err)
	}

	defer client.Close()

	session, _, err := client.Describe(parsed)
	if err != nil {
		return nil, fmt.Errorf("rtsp describe: %w", err)
	}

	tracks, err := openTracks(session)
	if err != nil {
		return nil, err
	}

	mux := newFMP4Muxer(writer, tracks.h264.SPS, tracks.h264.PPS, tracks.aac)

	captureCtx, cancel := context.WithTimeout(ctx, opts.Duration+captureTimeoutPadding)
	defer cancel()

	state := &captureState{}
	stop := makeStopper(state, cancel, client)

	if err := client.SetupAll(session.BaseURL, session.Medias); err != nil {
		return nil, fmt.Errorf("rtsp setup: %w", err)
	}

	attachVideo(client, tracks, mux, state, opts, stop)
	attachAudio(client, tracks, mux, state, opts, stop)

	if _, err := client.Play(nil); err != nil {
		return nil, fmt.Errorf("rtsp play: %w", err)
	}

	if err := waitCapture(captureCtx, client, state, stop); err != nil {
		return nil, err
	}

	if err := mux.close(); err != nil {
		return nil, err
	}

	return state.result(mux.hasAudio())
}

func newRTSPClient(parsed *base.URL, insecure bool) *gortsplib.Client {
	proto := gortsplib.ProtocolTCP
	client := &gortsplib.Client{
		Scheme:   parsed.Scheme,
		Host:     parsed.Host,
		Protocol: &proto,
	}

	if insecure {
		client.TLSConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // LAN self-signed
	}

	return client
}

type mediaTracks struct {
	h264Media *description.Media
	h264      *format.H264
	h264Dec   *rtph264.Decoder
	aacMedia  *description.Media
	aac       *format.MPEG4Audio
	aacDec    *rtpmpeg4audio.Decoder
}

func openTracks(session *description.Session) (*mediaTracks, error) {
	tracks := &mediaTracks{}

	tracks.h264Media = session.FindFormat(&tracks.h264)
	if tracks.h264Media == nil {
		return nil, ErrNoH264
	}

	dec, err := tracks.h264.CreateDecoder()
	if err != nil {
		return nil, fmt.Errorf("h264 rtp decoder: %w", err)
	}

	tracks.h264Dec = dec

	tracks.aacMedia = session.FindFormat(&tracks.aac)
	if tracks.aacMedia == nil {
		return tracks, nil
	}

	aacDec, err := tracks.aac.CreateDecoder()
	if err != nil {
		return nil, fmt.Errorf("aac rtp decoder: %w", err)
	}

	tracks.aacDec = aacDec

	return tracks, nil
}

func makeStopper(state *captureState, cancel context.CancelFunc, client *gortsplib.Client) func(error) {
	return func(err error) {
		state.mu.Lock()
		defer state.mu.Unlock()

		if state.done {
			return
		}

		state.done = true
		if err != nil && state.writeErr == nil {
			state.writeErr = err
		}

		cancel()

		go client.Close() // unblock Wait
	}
}

func attachVideo(
	client *gortsplib.Client,
	tracks *mediaTracks,
	mux *fmp4Muxer,
	state *captureState,
	opts Options,
	stop func(error),
) {
	client.OnPacketRTP(tracks.h264Media, tracks.h264, func(pkt *rtp.Packet) {
		if state.isDone() {
			return
		}

		pts, ok := client.PacketPTS(tracks.h264Media, pkt)
		if !ok {
			return
		}

		accessUnit, err := tracks.h264Dec.Decode(pkt)
		if err != nil {
			if !errors.Is(err, rtph264.ErrNonStartingPacketAndNoPrevious) &&
				!errors.Is(err, rtph264.ErrMorePacketsNeeded) {
				stop(fmt.Errorf("rtp h264 decode: %w", err))
			}

			return
		}

		nbytes, err := mux.writeVideo(accessUnit, pts)
		if err != nil {
			stop(err)

			return
		}

		if nbytes == 0 {
			return
		}

		if state.noteVideo(nbytes, opts) {
			stop(nil)
		}
	})
}

func attachAudio(
	client *gortsplib.Client,
	tracks *mediaTracks,
	mux *fmp4Muxer,
	state *captureState,
	opts Options,
	stop func(error),
) {
	if tracks.aacMedia == nil || tracks.aacDec == nil {
		return
	}

	client.OnPacketRTP(tracks.aacMedia, tracks.aac, func(pkt *rtp.Packet) {
		if state.isDone() {
			return
		}

		pts, ok := client.PacketPTS(tracks.aacMedia, pkt)
		if !ok {
			return
		}

		frames, err := tracks.aacDec.Decode(pkt)
		if err != nil {
			if !errors.Is(err, rtpmpeg4audio.ErrMorePacketsNeeded) {
				stop(fmt.Errorf("rtp aac decode: %w", err))
			}

			return
		}

		for i, frame := range frames {
			framePTS := pts + int64(i)*int64(mpeg4audio.SamplesPerAccessUnit)

			nbytes, err := mux.writeAudio(frame, framePTS)
			if err != nil {
				stop(err)

				return
			}

			if state.noteAudio(nbytes, opts) {
				stop(nil)

				return
			}
		}
	})
}

func waitCapture(
	captureCtx context.Context,
	client *gortsplib.Client,
	state *captureState,
	stop func(error),
) error {
	waitErr := make(chan error, 1)

	go func() {
		waitErr <- client.Wait()
	}()

	select {
	case <-captureCtx.Done():
		stop(nil)
	case err := <-waitErr:
		if err != nil && !errors.Is(err, context.Canceled) && !state.isDone() {
			return fmt.Errorf("rtsp wait: %w", err)
		}
	}

	return nil
}

type captureState struct {
	mu       sync.Mutex
	frames   int
	written  int64
	started  time.Time
	done     bool
	writeErr error
}

func (s *captureState) isDone() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.done
}

func (s *captureState) noteVideo(nbytes int, opts Options) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started.IsZero() {
		s.started = time.Now()
	}

	s.frames++
	s.written += int64(nbytes)

	return s.hitLimit(opts)
}

func (s *captureState) noteAudio(nbytes int, opts Options) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started.IsZero() {
		return false
	}

	s.written += int64(nbytes)

	return s.hitLimit(opts)
}

func (s *captureState) hitLimit(opts Options) bool {
	elapsed := time.Since(s.started)
	if elapsed >= opts.Duration {
		return true
	}

	return opts.MaxBytes > 0 && s.written >= opts.MaxBytes
}

func (s *captureState) result(hasAudio bool) (*Result, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.writeErr != nil {
		return nil, s.writeErr
	}

	if s.frames == 0 {
		return nil, ErrNoFrames
	}

	var dur time.Duration
	if !s.started.IsZero() {
		dur = time.Since(s.started)
	}

	return &Result{
		Frames:   s.frames,
		Bytes:    s.written,
		Duration: dur,
		HasAudio: hasAudio,
	}, nil
}

type fmp4Muxer struct {
	w         io.Writer
	sps, pps  []byte
	aacFormat *format.MPEG4Audio

	mu           sync.Mutex
	dtsExtractor *h264.DTSExtractor
	started      bool
	videoDecode  uint64
	audioDecode  uint64
	prevVideoPTS int64
	prevAudioPTS int64
	videoSamples []mp4.FullSample
	audioSamples []mp4.FullSample
	audioRate    uint32
}

func newFMP4Muxer(writer io.Writer, sps, pps []byte, aacFmt *format.MPEG4Audio) *fmp4Muxer {
	mux := &fmp4Muxer{w: writer, sps: sps, pps: pps, aacFormat: aacFmt}
	if aacFmt != nil {
		mux.audioRate = uint32(aacFmt.ClockRate()) //nolint:gosec // sample rates fit uint32
	}

	return mux
}

func (m *fmp4Muxer) hasAudio() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	return len(m.audioSamples) > 0
}

func (m *fmp4Muxer) writeVideo(accessUnit [][]byte, pts int64) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	filtered, idr, sps, pps := filterAU(accessUnit)
	if sps != nil {
		m.sps = sps
	}

	if pps != nil {
		m.pps = pps
	}

	if len(filtered) == 0 {
		return 0, nil
	}

	if m.dtsExtractor == nil {
		if !idr {
			return 0, nil
		}

		if len(m.sps) == 0 || len(m.pps) == 0 {
			return 0, ErrNoParameter
		}

		m.dtsExtractor = &h264.DTSExtractor{}
		m.dtsExtractor.Initialize()
		m.started = true
		m.prevVideoPTS = pts
	}

	return m.appendVideo(filtered, pts, idr)
}

func (m *fmp4Muxer) appendVideo(filtered [][]byte, pts int64, idr bool) (int, error) {
	if idr {
		filtered = append([][]byte{m.sps, m.pps}, filtered...)
	}

	dts, err := m.dtsExtractor.Extract(filtered, pts)
	if err != nil {
		return 0, fmt.Errorf("dts extract: %w", err)
	}

	sampleData := avccFromNALUs(filtered)
	dur := sampleDuration(pts, m.prevVideoPTS, defaultFrameTicks, len(m.videoSamples) > 0)
	m.prevVideoPTS = pts

	flags := mp4.NonSyncSampleFlags
	if idr {
		flags = mp4.SyncSampleFlags
	}

	m.videoSamples = append(m.videoSamples, mp4.FullSample{
		Sample: mp4.Sample{
			Flags:                 flags,
			Dur:                   dur,
			Size:                  uint32(len(sampleData)), //nolint:gosec // sample sizes are small
			CompositionTimeOffset: compositionOffset(pts, dts),
		},
		DecodeTime: m.videoDecode,
		Data:       sampleData,
	})
	m.videoDecode += uint64(dur)

	return len(sampleData), nil
}

func (m *fmp4Muxer) writeAudio(frame []byte, pts int64) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Defer init until close so we only declare AAC when samples exist.
	// Drop early audio until the first video IDR (short clips; negligible).
	if !m.started || m.aacFormat == nil || len(frame) == 0 {
		return 0, nil
	}

	fallback := uint32(mpeg4audio.SamplesPerAccessUnit)
	dur := sampleDuration(pts, m.prevAudioPTS, fallback, len(m.audioSamples) > 0)
	m.prevAudioPTS = pts

	data := append([]byte(nil), frame...)
	m.audioSamples = append(m.audioSamples, mp4.FullSample{
		Sample: mp4.Sample{
			Flags: mp4.SyncSampleFlags,
			Dur:   dur,
			Size:  uint32(len(data)), //nolint:gosec // frame sizes are small
		},
		DecodeTime: m.audioDecode,
		Data:       data,
	})
	m.audioDecode += uint64(dur)

	return len(data), nil
}

func sampleDuration(pts, prevPTS int64, fallback uint32, havePrev bool) uint32 {
	if !havePrev || pts <= prevPTS {
		return fallback
	}

	delta := pts - prevPTS
	if delta > 0 && delta < int64(^uint32(0)) {
		return uint32(delta)
	}

	return fallback
}

func compositionOffset(pts, dts int64) int32 {
	cto := pts - dts
	if cto > int64(^uint32(0)>>1) {
		return 0
	}

	return int32(cto) //nolint:gosec // clamped above
}

func (m *fmp4Muxer) writeInit(includeAudio bool) error {
	init := mp4.CreateEmptyInit()
	init.AddEmptyTrack(clockRateH264, "video", "und")

	if err := init.Moov.Trak.SetAVCDescriptor("avc1", [][]byte{m.sps}, [][]byte{m.pps}, true); err != nil {
		return fmt.Errorf("avc descriptor: %w", err)
	}

	if includeAudio && m.aacFormat != nil && m.aacFormat.Config != nil && m.audioRate > 0 {
		init.AddEmptyTrack(m.audioRate, "audio", "und")

		traks := init.Moov.Traks
		audioTrak := traks[len(traks)-1]

		if err := setMPEG4AudioDescriptor(audioTrak, m.aacFormat.Config); err != nil {
			return err
		}
	}

	if err := init.Encode(m.w); err != nil {
		return fmt.Errorf("write init: %w", err)
	}

	return nil
}

// setMPEG4AudioDescriptor writes the SDP AudioSpecificConfig into esds/mp4a.
// Do not use SetAACDescriptor: it hardcodes stereo ASC and buzzes on mono SecuritySpy AAC.
// ChannelCount must match the ASC; lying (e.g. 2ch entry + mono ASC) makes Telegram drop audio.
func setMPEG4AudioDescriptor(trak *mp4.TrakBox, cfg *mpeg4audio.AudioSpecificConfig) error {
	ascBytes, err := cfg.Marshal()
	if err != nil {
		return fmt.Errorf("marshal aac config: %w", err)
	}

	channels := uint16(cfg.ChannelConfig)
	if channels == 0 {
		channels = 1
	}

	sampleRate := uint16(cfg.SampleRate) //nolint:gosec // AAC rates fit uint16
	esds := mp4.CreateEsdsBox(ascBytes)
	mp4a := mp4.CreateAudioSampleEntryBox("mp4a", channels, aacBitsPerSample, sampleRate, esds)
	trak.Mdia.Minf.Stbl.Stsd.AddChild(mp4a)

	return nil
}

func (m *fmp4Muxer) close() error { //nolint:cyclop // init + multi-track fragment
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.started || len(m.videoSamples) == 0 {
		return nil
	}

	includeAudio := len(m.audioSamples) > 0
	if err := m.writeInit(includeAudio); err != nil {
		return err
	}

	seg := mp4.NewMediaSegment()

	frag, err := m.newFragment(includeAudio)
	if err != nil {
		return err
	}

	seg.AddFragment(frag)

	for _, sample := range m.videoSamples {
		if err := frag.AddFullSampleToTrack(sample, videoTrackID); err != nil {
			return fmt.Errorf("add video sample: %w", err)
		}
	}

	if includeAudio {
		for _, sample := range m.audioSamples {
			if err := frag.AddFullSampleToTrack(sample, audioTrackID); err != nil {
				return fmt.Errorf("add audio sample: %w", err)
			}
		}
	}

	if err := seg.Encode(m.w); err != nil {
		return fmt.Errorf("encode segment: %w", err)
	}

	return nil
}

func (m *fmp4Muxer) newFragment(includeAudio bool) (*mp4.Fragment, error) {
	if includeAudio {
		frag, err := mp4.CreateMultiTrackFragment(1, []uint32{videoTrackID, audioTrackID})
		if err != nil {
			return nil, fmt.Errorf("create fragment: %w", err)
		}

		return frag, nil
	}

	frag, err := mp4.CreateFragment(1, videoTrackID)
	if err != nil {
		return nil, fmt.Errorf("create fragment: %w", err)
	}

	return frag, nil
}

func filterAU(accessUnit [][]byte) ([][]byte, bool, []byte, []byte) {
	var (
		nalus [][]byte
		idr   bool
		sps   []byte
		pps   []byte
	)

	for _, nalu := range accessUnit {
		if len(nalu) == 0 {
			continue
		}

		switch h264.NALUType(nalu[0] & nalTypeMask) {
		case h264.NALUTypeSPS:
			sps = nalu
		case h264.NALUTypePPS:
			pps = nalu
		case h264.NALUTypeAccessUnitDelimiter:
			continue
		case h264.NALUTypeIDR:
			idr = true

			nalus = append(nalus, nalu)
		default:
			nalus = append(nalus, nalu)
		}
	}

	return nalus, idr, sps, pps
}

func avccFromNALUs(nalus [][]byte) []byte {
	size := 0
	for _, nalu := range nalus {
		size += avccLengthSize + len(nalu)
	}

	buf := make([]byte, size)
	off := 0

	for _, nalu := range nalus {
		binary.BigEndian.PutUint32(buf[off:], uint32(len(nalu))) //nolint:gosec // NALU sizes fit
		off += avccLengthSize
		off += copy(buf[off:], nalu)
	}

	return buf
}
