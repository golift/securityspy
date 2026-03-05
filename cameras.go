package securityspy

import (
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golift.io/ffmpeg"
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

// StreamVideo streams a segment of video from a camera using FFMPEG.
// VidOps defines the video options for the video stream.
// Returns an io.ReadCloser with the video stream. Close() it when finished.
func (c *Camera) StreamVideo(ops *VidOps, length time.Duration, maxsize int64) (io.ReadCloser, error) {
	ffmpg := ffmpeg.Get(&ffmpeg.Config{
		FFMPEG: c.server.Encoder,
		Time:   int(length.Seconds()),
		Audio:  true,    // Sure why not.
		Size:   maxsize, // max file size (always goes over). use 2000000 for 2.5MB
		Copy:   true,    // Always copy securityspy RTSP urls.
	})

	params := c.makeRequestParams(ops)

	if p := c.server.Auth(); p != "" {
		params.Set("auth", p)
	}

	args, video, err := ffmpg.GetVideo(c.makeVideoURL(ops, params), c.Name)
	if err != nil {
		return nil, fmt.Errorf("capturing stream for %s: %w; ffmpeg command: %s",
			c.Name, err, redactAuth(strings.ReplaceAll(args, "\n", " ")))
	}

	return video, nil
}

// SaveVideo saves a segment of video from a camera to a file using FFMPEG.
func (c *Camera) SaveVideo(ops *VidOps, length time.Duration, maxsize int64, outputFile string) error {
	if _, err := os.Stat(outputFile); !os.IsNotExist(err) {
		return ErrPathExists
	}

	ffmpg := ffmpeg.Get(&ffmpeg.Config{
		FFMPEG: c.server.Encoder,
		Time:   int(length.Seconds()),
		Audio:  true,
		Size:   maxsize, // max file size (always goes over). use 2000000 for 2.5MB
		Copy:   true,    // Always copy securityspy RTSP urls.
	})

	params := c.makeRequestParams(ops)

	if p := c.server.Auth(); p != "" {
		params.Set("auth", p)
	}

	cmd, out, err := ffmpg.SaveVideo(c.makeVideoURL(ops, params), outputFile, c.Name)
	if err != nil {
		return fmt.Errorf("capturing video for %s: %w; ffmpeg command: %s; ffmpeg output: %s",
			c.Name,
			err,
			redactAuth(strings.ReplaceAll(cmd, "\n", " ")),
			redactAuth(trimTail(strings.ReplaceAll(out, "\n", " "), ffmpegOutputTail)),
		)
	}

	return nil
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
func (c *Camera) GetJPEG(ops *VidOps) (image.Image, error) {
	if ops == nil {
		ops = &VidOps{}
	}

	ops.FPS = -1 // not used for single image

	ctx, cancel := context.WithTimeout(context.Background(), c.server.TimeoutDur())
	defer cancel()

	resp, err := c.server.GetContext(ctx, "++image", c.makeRequestParams(ops))
	if err != nil {
		return nil, fmt.Errorf("getting image: %w", err)
	}
	defer resp.Body.Close()

	jpgImage, err := jpeg.Decode(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("decoding jpeg: %w", err)
	}

	return jpgImage, nil
}

// SaveJPEG gets a picture from a camera and puts it in a file (path).
// The file will be overwritten if it exists.
// VidOps defines the image size. ops.FPS is ignored.
func (c *Camera) SaveJPEG(ops *VidOps, path string) error {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return ErrPathExists
	}

	jpgImage, err := c.GetJPEG(ops)
	if err != nil {
		return fmt.Errorf("getting jpeg: %w", err)
	}

	oFile, err := os.Create(path) //nolint:gosec // we are creating a file in a safe way.
	if err != nil {
		return fmt.Errorf("os.Create: %w", err)
	}
	defer oFile.Close()

	err = jpeg.Encode(oFile, jpgImage, nil)
	if err != nil {
		return fmt.Errorf("encoding jpeg: %w", err)
	}

	return nil
}

// ToggleContinuous arms (true) or disarms (false) a camera's continuous capture mode.
func (c *Camera) ToggleContinuous(arm CameraArmMode) error {
	params := make(url.Values)
	params.Set("arm", string(arm))

	if err := c.server.SimpleReq("++ssControlContinuous", params, c.Number); err != nil {
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

// SetSchedule configures a camera mode's primary schedule.
// Get a list of schedules IDs you can use here from server.Info.Schedules.
// CameraModes are constants with names that start with CameraMode*.
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
	maxQuality       = 100
	maxFPS           = 60
	ffmpegOutputTail = 2048
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

// makeVideoURL creates a video URL for the camera based on if it's rtsp or not, and other input options.
func (c *Camera) makeVideoURL(ops *VidOps, params url.Values) string { //nolint:cyclop // oh well?
	if ops != nil && ops.UseHTTP {
		if ops.FPS > 0 {
			params.Set("fps", strconv.Itoa(ops.FPS))
		}

		vcodec := "h264"
		if ops.VCodec != "" {
			vcodec = ops.VCodec
		}

		params.Del("req_fps")
		params.Set("vcodec", vcodec)

		return c.server.BaseURL() + "video?" + params.Encode()
	}

	if ops != nil && ops.FPS > 0 {
		params.Set("fps", strconv.Itoa(ops.FPS))
	}

	vcodec, acodec := "h264", "aac"
	if ops != nil && ops.VCodec != "" {
		vcodec = ops.VCodec
	}

	if ops != nil && ops.ACodec != "" {
		acodec = ops.ACodec
	}

	params.Del("req_fps")
	params.Set("vcodec", vcodec)
	params.Set("acodec", acodec)

	baseURL := strings.Replace(c.server.BaseURL(), "https://", "rtsps://", 1)
	baseURL = strings.Replace(baseURL, "http://", "rtsp://", 1)

	return baseURL + "stream?" + params.Encode()
}

var authRegex = regexp.MustCompile(`auth=[^&\s]+`)

func redactAuth(input string) string {
	return authRegex.ReplaceAllString(input, "auth=REDACTED")
}

func trimTail(input string, maximum int) string {
	if maximum <= 0 || len(input) <= maximum {
		return input
	}

	return input[len(input)-maximum:]
}
