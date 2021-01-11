package securityspy

import (
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"golift.io/ffmpeg"
)

// All returns interfaces for every camera.
func (c *Cameras) All() (cams []*Camera) {
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

	return nil
}

// StreamVideo streams a segment of video from a camera using FFMPEG.
// VidOps defines the video options for the video stream.
// Returns an io.ReadCloser with the video stream. Close() it when finished.
func (c *Camera) StreamVideo(ops *VidOps, length time.Duration, maxsize int64) (io.ReadCloser, error) {
	f := ffmpeg.Get(&ffmpeg.Config{
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

	params.Set("codec", "h264")
	// This is kinda crude, but will handle 99%.
	url := strings.Replace(c.server.BaseURL(), "http", "rtsp", 1) + "++stream"

	// RTSP doesn't rewally work with HTTPS, and FFMPEG doesn't care about the cert.
	args, video, err := f.GetVideo(url+"?"+params.Encode(), c.Name)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, strings.ReplaceAll(args, "\n", " "))
	}

	return video, nil
}

// SaveVideo saves a segment of video from a camera to a file using FFMPEG.
func (c *Camera) SaveVideo(ops *VidOps, length time.Duration, maxsize int64, outputFile string) error {
	if _, err := os.Stat(outputFile); !os.IsNotExist(err) {
		return ErrorPathExists
	}

	f := ffmpeg.Get(&ffmpeg.Config{
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

	params.Set("codec", "h264")
	// This is kinda crude, but will handle 99%.

	url := strings.Replace(c.server.BaseURL(), "http", "rtsp", 1) + "++stream"

	_, out, err := f.SaveVideo(url+"?"+params.Encode(), outputFile, c.Name)
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.ReplaceAll(out, "\n", " "))
	}

	return nil
}

// StreamMJPG makes a web request to retrieve a motion JPEG stream.
// Returns an io.ReadCloser that will (hopefully) never end.
func (c *Camera) StreamMJPG(ops *VidOps) (io.ReadCloser, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.server.config.Timeout)
	defer cancel()

	resp, err := c.server.GetContext(ctx, "++video", c.makeRequestParams(ops))
	if err != nil {
		return nil, fmt.Errorf("getting video: %w", err)
	}

	return resp.Body, nil
}

// StreamH264 makes a web request to retrieve an H264 stream.
// Returns an io.ReadCloser that will (hopefully) never end.
func (c *Camera) StreamH264(ops *VidOps) (io.ReadCloser, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.server.config.Timeout)
	defer cancel()

	resp, err := c.server.GetContext(ctx, "++stream", c.makeRequestParams(ops))
	if err != nil {
		return nil, fmt.Errorf("getting stream: %w", err)
	}

	return resp.Body, nil
}

// StreamG711 makes a web request to retrieve an G711 audio stream.
// Returns an io.ReadCloser that will (hopefully) never end.
func (c *Camera) StreamG711() (io.ReadCloser, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.server.config.Timeout)
	defer cancel()

	resp, err := c.server.GetContext(ctx, "++audio", c.makeRequestParams(nil))
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
	ops.FPS = -1 // not used for single image

	ctx, cancel := context.WithTimeout(context.Background(), c.server.config.Timeout)
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
		return ErrorPathExists
	}

	jpgImage, err := c.GetJPEG(ops)
	if err != nil {
		return fmt.Errorf("getting jpeg: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("os.Create: %w", err)
	}
	defer f.Close()

	err = jpeg.Encode(f, jpgImage, nil)
	if err != nil {
		return fmt.Errorf("encoding jpeg: %w", err)
	}

	return nil
}

// ToggleContinuous arms (true) or disarms (false) a camera's continuous capture mode.
func (c *Camera) ToggleContinuous(arm CameraArmMode) error {
	params := make(url.Values)
	params.Set("arm", string(arm))

	return c.server.SimpleReq("++ssControlContinuous", params, c.Number)
}

// ToggleMotion arms (true) or disarms (false) a camera's motion capture mode.
func (c *Camera) ToggleMotion(arm CameraArmMode) error {
	params := make(url.Values)
	params.Set("arm", string(arm))

	return c.server.SimpleReq("++ssControlMotionCapture", params, c.Number)
}

// ToggleActions arms (true) or disarms (false) a camera's actions.
func (c *Camera) ToggleActions(arm CameraArmMode) error {
	params := make(url.Values)
	params.Set("arm", string(arm))

	return c.server.SimpleReq("++ssControlActions", params, c.Number)
}

// TriggerMotion sets a camera as currently seeing motion.
// Other actions likely occur because of this!
func (c *Camera) TriggerMotion() error {
	return c.server.SimpleReq("++triggermd", make(url.Values), c.Number)
}

// SetSchedule configures a camera mode's primary schedule.
// Get a list of schedules IDs you can use here from server.Info.Schedules.
// CameraModes are constants with names that start with CameraMode*.
func (c *Camera) SetSchedule(mode CameraMode, scheduleID int) error {
	params := make(url.Values)
	params.Set("mode", string(mode))
	params.Set("id", strconv.Itoa(scheduleID))

	return c.server.SimpleReq("++ssSetSchedule", params, c.Number)
}

// SetScheduleOverride temporarily overrides a camera mode's current schedule.
// Get a list of overrides IDs you can use here from server.Info.ScheduleOverrides.
// CameraModes are constants with names that start with CameraMode*.
func (c *Camera) SetScheduleOverride(mode CameraMode, overrideID int) error {
	params := make(url.Values)
	params.Set("mode", string(mode))
	params.Set("id", strconv.Itoa(overrideID))

	return c.server.SimpleReq("++ssSetOverride", params, c.Number)
}

/* INTERFACE HELPER METHODS FOLLOW */

const (
	maxQuality = 100
	maxFPS     = 60
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
