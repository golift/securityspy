package securityspy

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"golift.io/securityspy/v2/internal/rtspclip"
	"golift.io/securityspy/v2/server"
)

// All returns interfaces for every camera.
func (c *Cameras) All() []*Camera {
	return c.cameras
}

// ByNum returns an interface for a single camera.
func (c *Cameras) ByNum(number int) *Camera {
	for _, cam := range c.cameras {
		if cam.Number == number {
			return cam
		}
	}

	return nil
}

// ByName returns an interface for a single camera, using the name.
func (c *Cameras) ByName(name string) *Camera {
	for _, cam := range c.cameras {
		if cam.Name == name {
			return cam
		}
	}

	// Try again, case-insensitive.
	for _, cam := range c.cameras {
		if strings.EqualFold(cam.Name, name) {
			return cam
		}
	}

	return nil
}

// StreamVideo returns a ReadCloser that yields a progressive fragmented MP4
// (H.264 or H.265 + AAC when present at init time). Init is written after the first IDR;
// media fragments flush while capture continues. Close() cancels an in-progress capture.
// UseHTTP is not supported (returns ErrHTTPVideoUnsupported).
func (c *Camera) StreamVideo(ops *VidOps, length time.Duration, maxsize int64) (io.ReadCloser, error) {
	rtspURL, err := c.makeVideoURL(ops, c.makeRequestParams(ops))
	if err != nil {
		return nil, err
	}

	video, err := rtspclip.StreamMP4(context.Background(), rtspURL, c.rtspclipOptions(length, maxsize))
	if err != nil {
		return nil, fmt.Errorf("capturing stream for %s: %w", c.Name, err)
	}

	return video, nil
}

// SaveVideo saves a short RTSP remux (H.264 or H.265 + AAC when present) to an MP4/MOV path.
// UseHTTP is not supported (returns ErrHTTPVideoUnsupported).
// Set VidOps.VCodec to "h265" for HEVC cameras (see PreferredVCodec).
func (c *Camera) SaveVideo(ops *VidOps, length time.Duration, maxsize int64, outputFile string) error {
	if _, err := os.Stat(outputFile); !os.IsNotExist(err) {
		return ErrPathExists
	}

	rtspURL, err := c.makeVideoURL(ops, c.makeRequestParams(ops))
	if err != nil {
		return err
	}

	_, err = rtspclip.SaveMP4(context.Background(), rtspURL, outputFile, c.rtspclipOptions(length, maxsize))
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return ErrPathExists
		}

		return fmt.Errorf("capturing video for %s: %w", c.Name, err)
	}

	return nil
}

func (c *Camera) rtspclipOptions(length time.Duration, maxsize int64) rtspclip.Options {
	return rtspclip.Options{
		Duration:           length,
		MaxBytes:           maxsize,
		InsecureSkipVerify: !c.server.VerifySSL,
	}
}

// StreamMJPG makes a web request to retrieve a motion JPEG stream.
// Returns an io.ReadCloser that will (hopefully) never end.
func (c *Camera) StreamMJPG(ops *VidOps) (io.ReadCloser, error) {
	return c.StreamMJPGContext(context.Background(), ops)
}

// StreamMJPGContext makes a web request to retrieve a motion JPEG stream.
// Returns an io.ReadCloser that will (hopefully) never end.
func (c *Camera) StreamMJPGContext(ctx context.Context, ops *VidOps) (io.ReadCloser, error) {
	resp, err := c.server.GetContextClient(ctx, "++video", c.makeRequestParams(ops), c.streamHTTPClient())
	if err != nil {
		return nil, fmt.Errorf("getting video: %w", err)
	}

	return resp.Body, nil
}

// StreamH264 makes a web request to retrieve an H264 stream.
// Returns an io.ReadCloser that will (hopefully) never end.
func (c *Camera) StreamH264(ops *VidOps) (io.ReadCloser, error) {
	return c.StreamH264Context(context.Background(), ops)
}

// StreamH264Context makes a web request to retrieve an H264 stream.
// Returns an io.ReadCloser that will (hopefully) never end.
func (c *Camera) StreamH264Context(ctx context.Context, ops *VidOps) (io.ReadCloser, error) {
	resp, err := c.server.GetContextClient(ctx, "++stream", c.makeRequestParams(ops), c.streamHTTPClient())
	if err != nil {
		return nil, fmt.Errorf("getting stream: %w", err)
	}

	return resp.Body, nil
}

// StreamG711 makes a web request to retrieve an G711 audio stream.
// Returns an io.ReadCloser that will (hopefully) never end.
func (c *Camera) StreamG711() (io.ReadCloser, error) {
	return c.StreamG711Context(context.Background())
}

// StreamG711Context makes a web request to retrieve an G711 audio stream.
// Returns an io.ReadCloser that will (hopefully) never end.
func (c *Camera) StreamG711Context(ctx context.Context) (io.ReadCloser, error) {
	resp, err := c.server.GetContextClient(ctx, "++audio", c.makeRequestParams(nil), c.streamHTTPClient())
	if err != nil {
		return nil, fmt.Errorf("getting audio: %w", err)
	}

	return resp.Body, nil
}

// PostG711 makes a POST request to send audio to a camera with a speaker.
// Accepts an io.ReadCloser that will be closed. Probably an open file.
// This is untested. Report your success or failure!
func (c *Camera) PostG711(audio io.ReadCloser) ([]byte, error) {
	if audio == nil {
		return nil, nil
	}

	body, err := c.server.Post("++audio", c.makeRequestParams(nil), audio)
	if err != nil {
		return nil, fmt.Errorf("posting audio: %w", err)
	}

	return body, nil
}

// GetJPEG returns an images from a camera.
// VidOps defines the image size. ops.FPS is ignored.
// Makes several attempts in case of an error or time out.
func (c *Camera) GetJPEG(ops *VidOps) (image.Image, error) {
	data, err := c.fetchJPEGBytes(ops)
	if err != nil {
		return nil, err
	}

	jpgImage, err := jpeg.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decoding jpeg: %w", err)
	}

	return jpgImage, nil
}

// SaveJPEG gets a picture from a camera and puts it in a file (path).
// Fails if the path already exists. VidOps defines the image size; ops.FPS is ignored.
// Writes the server JPEG bytes directly (no decode/re-encode).
func (c *Camera) SaveJPEG(ops *VidOps, path string) error {
	data, err := c.fetchJPEGBytes(ops)
	if err != nil {
		return fmt.Errorf("getting jpeg: %w", err)
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, jpegFilePerm) //nolint:gosec // caller-chosen path
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return ErrPathExists
		}

		return fmt.Errorf("creating jpeg: %w", err)
	}

	_, writeErr := file.Write(data)
	closeErr := file.Close()

	if writeErr != nil {
		_ = os.Remove(path)

		return fmt.Errorf("writing jpeg: %w", writeErr)
	}

	if closeErr != nil {
		_ = os.Remove(path)

		return fmt.Errorf("closing jpeg: %w", closeErr)
	}

	return nil
}

// fetchJPEGBytes downloads a still from ++image. The request context stays alive
// until the body is fully read — canceling earlier truncates the JPEG.
// Retries use server.Config.JPEGRetries; each attempt is bounded by Timeout.
func (c *Camera) fetchJPEGBytes(ops *VidOps) ([]byte, error) {
	if ops == nil {
		ops = &VidOps{}
	}

	ops.FPS = -1 // not used for single image

	client := c.streamHTTPClient() // Timeout=0; context bounds the whole fetch

	var lastErr error

	for range c.server.JPEGTries() {
		data, err := c.fetchJPEGBytesOnce(ops, client)
		if err == nil {
			return data, nil
		}

		lastErr = err
		// Offline / missing cameras won't recover within this call.
		if errors.Is(err, ErrCameraUnavailable) {
			return nil, err
		}
	}

	return nil, lastErr
}

func (c *Camera) fetchJPEGBytesOnce(ops *VidOps, client *http.Client) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.server.TimeoutDur())

	resp, err := c.server.GetContextClient(ctx, "++image", c.makeRequestParams(ops), client)
	if err != nil {
		cancel()

		return nil, fmt.Errorf("getting image: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		_ = resp.Body.Close()

		cancel()

		return nil, fmt.Errorf("%w: %s", ErrCameraUnavailable, c.Name)
	}

	if resp.StatusCode != http.StatusOK {
		preview, _ := io.ReadAll(io.LimitReader(resp.Body, jpegSOIPreviewMax))
		_ = resp.Body.Close()

		cancel()

		return nil, fmt.Errorf("%w: status %d (%q)", server.ErrCmdNotOK, resp.StatusCode, preview)
	}

	data, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	cancel()

	if err != nil {
		return nil, fmt.Errorf("reading image: %w", err)
	}

	if len(data) < 2 || data[0] != 0xFF || data[1] != 0xD8 {
		preview := data
		if len(preview) > jpegSOIPreviewMax {
			preview = preview[:jpegSOIPreviewMax]
		}

		return nil, fmt.Errorf("%w: got %q", ErrInvalidJPEG, preview)
	}

	return data, nil
}

// ToggleContinuous arms or disarms continuous capture via ++ssControlContinuous.
//
// ToggleMotion and ToggleActions use similar ++ssControl* endpoints and work on
// tested SecuritySpy v5 servers. Continuous control is often unavailable (HTTP 404);
// in that case this method returns ErrUnsupported. Use SetSchedule with
// CameraModeContinuous to arm or disarm continuous capture via the schedule API instead.
func (c *Camera) ToggleContinuous(arm CameraArmMode) error {
	params := make(url.Values)
	params.Set("arm", string(arm))

	if err := c.server.SimpleReq("++ssControlContinuous", params, c.Number); err != nil {
		if errors.Is(err, server.ErrNotFound) {
			return ErrUnsupported
		}

		return fmt.Errorf("request failed: %w", err)
	}

	return nil
}

// ToggleMotion arms (true) or disarms (false) a camera's motion capture mode.
func (c *Camera) ToggleMotion(arm CameraArmMode) error {
	params := make(url.Values)
	params.Set("arm", string(arm))

	if err := c.server.SimpleReq("++ssControlMotionCapture", params, c.Number); err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	return nil
}

// ToggleActions arms (true) or disarms (false) a camera's actions.
func (c *Camera) ToggleActions(arm CameraArmMode) error {
	params := make(url.Values)
	params.Set("arm", string(arm))

	if err := c.server.SimpleReq("++ssControlActions", params, c.Number); err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	return nil
}

// TriggerMotion sets a camera as currently seeing motion.
// Other actions likely occur because of this!
func (c *Camera) TriggerMotion() error {
	if err := c.server.SimpleReq("++triggermd", make(url.Values), c.Number); err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	return nil
}

// Modes returns the armed/disarmed status of continuous, motion, and actions modes.
func (c *Camera) Modes() (*CameraModes, error) {
	resp, err := c.server.Get("++cameramodes", c.makeRequestParams(nil))
	if err != nil {
		return nil, fmt.Errorf("getting camera modes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %s", ErrCameraModesStatus, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading camera modes: %w", err)
	}

	return parseCameraModes(string(body))
}

// HLSURL returns the authenticated HTTP Live Streaming URL for this camera.
func (c *Camera) HLSURL() string {
	params := c.makeRequestParams(nil)
	if auth := c.server.Auth(); auth != "" {
		params.Set("auth", auth)
	}

	return c.server.BaseURL() + "++hls?" + params.Encode()
}

// HLSMediaPlaylistURL returns the fixed-quality HLS media playlist URL (v6+).
// quality: 0 low, 1 medium, 2 high.
func (c *Camera) HLSMediaPlaylistURL(quality int) string {
	params := c.makeRequestParams(nil)
	params.Set("quality", strconv.Itoa(quality))

	if auth := c.server.Auth(); auth != "" {
		params.Set("auth", auth)
	}

	return c.server.BaseURL() + "++hls_mediaplaylist?" + params.Encode()
}

// LiveURL returns the HTML live-view page URL for this camera.
func (c *Camera) LiveURL() string {
	params := c.makeRequestParams(nil)
	if auth := c.server.Auth(); auth != "" {
		params.Set("auth", auth)
	}

	return c.server.BaseURL() + "++live?" + params.Encode()
}

// MultiplexURL builds an authenticated ++multiplex grid URL (v6+).
func (s *Server) MultiplexURL(ops *MultiplexOps) string {
	params := make(url.Values)

	if ops != nil {
		if len(ops.Cameras) > 0 {
			cams := make([]string, len(ops.Cameras))
			for i, n := range ops.Cameras {
				cams[i] = strconv.Itoa(n)
			}

			params.Set("cameras", strings.Join(cams, ","))
		}

		params.Set("cropMode", strconv.Itoa(ops.CropMode))
		params.Set("format", strconv.Itoa(ops.Format))

		if ops.FPS > 0 {
			params.Set("fps", strconv.Itoa(ops.FPS))
		}

		params.Set("border", strconv.Itoa(ops.Border))

		if ops.CamInfo {
			params.Set("camInfo", "1")
		}

		if ops.HiRes {
			params.Set("hiRes", "1")
		}
	}

	if auth := s.Auth(); auth != "" {
		params.Set("auth", auth)
	}

	return s.BaseURL() + "++multiplex?" + params.Encode()
}

// SetSchedule configures a camera mode's primary schedule.
// Get a list of schedules IDs you can use here from server.Info.ServerSchedules.
// CameraModes are constants with names that start with CameraMode*.
// Uses ++ssSetSchedule (also available as documented ++setSchedule on the server).
func (c *Camera) SetSchedule(mode CameraMode, scheduleID int) error {
	params := make(url.Values)
	params.Set("mode", string(mode))
	params.Set("id", strconv.Itoa(scheduleID))

	if err := c.server.SimpleReq("++ssSetSchedule", params, c.Number); err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	return nil
}

// SetScheduleOverride temporarily overrides a camera mode's current schedule.
// Get a list of overrides IDs you can use here from server.Info.ScheduleOverrides.
// CameraModes are constants with names that start with CameraMode*.
func (c *Camera) SetScheduleOverride(mode CameraMode, overrideID int) error {
	params := make(url.Values)
	params.Set("mode", string(mode))
	params.Set("id", strconv.Itoa(overrideID))

	if err := c.server.SimpleReq("++ssSetOverride", params, c.Number); err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	return nil
}

/* INTERFACE HELPER METHODS FOLLOW */

const (
	maxQuality     = 100
	maxFPS         = 60
	modeKeyParts   = 2
	authColonParts = 2
	aacStr         = "aac"
)

// makeRequestParams converts passed in ops to url.Values.
func (c *Camera) makeRequestParams(ops *VidOps) url.Values {
	params := make(url.Values)
	params.Set("cameraNum", strconv.Itoa(c.Number))

	if ops == nil {
		return params
	}

	if ops.Width != 0 {
		params.Set("width", strconv.Itoa(ops.Width))
	}

	if ops.Height != 0 {
		params.Set("height", strconv.Itoa(ops.Height))
	}

	if ops.Quality > maxQuality {
		ops.Quality = maxQuality
	}

	if ops.Quality > 0 {
		params.Set("quality", strconv.Itoa(ops.Quality))
	}

	if ops.FPS > maxFPS {
		ops.FPS = maxFPS
	}

	if ops.FPS > 0 {
		params.Set("req_fps", strconv.Itoa(ops.FPS))
	}

	return params
}

func (c *Camera) streamHTTPClient() *http.Client {
	client := c.server.HTTPClient()
	client.Timeout = 0

	return client
}

// PreferredVCodec returns the RTSP vcodec query value that matches this camera's
// native VideoFormat from ++systemInfo ("h265" for H.265/HEVC, otherwise "h264").
func (c *Camera) PreferredVCodec() string {
	format := strings.ToUpper(c.VideoFormat)
	if strings.Contains(format, "265") || strings.Contains(format, "HEVC") {
		return "h265"
	}

	return "h264"
}

// makeVideoURL builds an RTSP(S) ++stream URL with userinfo credentials.
// SecuritySpy rejects query auth= on RTSP; HTTP auth= is unchanged elsewhere.
// UseHTTP is unsupported for SaveVideo/StreamVideo.
func (c *Camera) makeVideoURL(ops *VidOps, params url.Values) (string, error) { //nolint:cyclop // ops/codec branches
	if ops != nil && ops.UseHTTP {
		return "", ErrHTTPVideoUnsupported
	}

	base, err := url.Parse(c.server.BaseURL())
	if err != nil {
		return "", fmt.Errorf("parse base url: %w", err)
	}

	scheme := "rtsp"
	if base.Scheme == "https" {
		scheme = "rtsps"
	}

	if ops != nil && ops.FPS > 0 {
		params.Set("fps", strconv.Itoa(ops.FPS))
	}

	vcodec, acodec := "h264", aacStr
	if ops != nil && ops.VCodec != "" {
		vcodec = ops.VCodec
	}

	if ops != nil && ops.ACodec != "" {
		acodec = ops.ACodec
	}

	params.Del("req_fps")
	params.Del("auth")
	params.Del("quality") // JPEG ++image/++video only; ignored/harmful on RTSP
	params.Del("vcodec")
	params.Del("acodec")
	params.Del("cameraNum")

	// SecuritySpy mishandles vcodec when acodec is the first query key.
	// url.Values.Encode() sorts alphabetically (acodec, cameraNum, ..., vcodec)
	// and that order makes SS serve H.264 even when vcodec=h265. Keep cameraNum
	// and codecs first; append the rest in Encode() order.
	parts := []string{
		"cameraNum=" + url.QueryEscape(strconv.Itoa(c.Number)),
		"vcodec=" + url.QueryEscape(vcodec),
		"acodec=" + url.QueryEscape(acodec),
	}
	if encoded := params.Encode(); encoded != "" {
		parts = append(parts, encoded)
	}

	out := (&url.URL{
		Scheme: scheme,
		Host:   base.Host,
		Path:   base.Path,
	}).JoinPath("stream")
	out.RawQuery = strings.Join(parts, "&")

	if user, pass := c.rtspUserPassword(); user != "" {
		out.User = url.UserPassword(user, pass)
	}

	return out.String(), nil
}

// rtspUserPassword recovers plaintext credentials from the library auth blob
// (base64 of username:password) for RTSP userinfo.
// Returns empty strings when the blob is missing or invalid so makeVideoURL omits userinfo
// instead of emitting rtsps://user:@host/... with a useless username.
func (c *Camera) rtspUserPassword() (string, string) {
	blob := c.server.Auth()
	if blob == "" {
		return "", ""
	}

	raw, err := base64.URLEncoding.DecodeString(blob)
	if err != nil {
		return "", ""
	}

	parts := strings.SplitN(string(raw), ":", authColonParts)
	if len(parts) != authColonParts || parts[0] == "" {
		return "", ""
	}

	return parts[0], parts[1]
}

func parseCameraModes(text string) (*CameraModes, error) {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	modes := &CameraModes{}

	for line := range strings.SplitSeq(strings.TrimSpace(text), "\n") {
		parts := strings.SplitN(strings.TrimSpace(line), ":", modeKeyParts)
		if len(parts) != modeKeyParts {
			continue
		}

		switch parts[0] {
		case "C":
			modes.Continuous = parts[1]
		case "M":
			modes.Motion = parts[1]
		case "A":
			modes.Actions = parts[1]
		}
	}

	if modes.Continuous == "" && modes.Motion == "" && modes.Actions == "" {
		return nil, fmt.Errorf("%w: %q", ErrCameraModesParse, text)
	}

	return modes, nil
}
