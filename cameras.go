package securityspy

import (
	"image"
	"image/jpeg"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	// Because I didn't feel like dealing with RTSP in Go. Maybe one day.
	ffmpeg "github.com/davidnewhall/go-securityspy/ffmpegencode"
	"github.com/pkg/errors"
)

// All returns interfaces for every camera.
func (c *Cameras) All() (cams []*Camera) {
	for _, cam := range c.server.systemInfo.CameraList.Cameras {
		// Add cameras' pointer to sub interfaces.
		cam.server = c.server
		cam.PTZ.camera = cam
		cams = append(cams, cam)
	}
	return
}

// ByNum returns an interface for a single camera.
func (c *Cameras) ByNum(number int) *Camera {
	for _, cam := range c.server.systemInfo.CameraList.Cameras {
		if cam.Number == number {
			cam.server = c.server
			cam.PTZ.camera = cam
			return cam
		}
	}
	return nil
}

// ByName returns an interface for a single camera, using the name.
func (c *Cameras) ByName(name string) *Camera {
	for _, cam := range c.server.systemInfo.CameraList.Cameras {
		if cam.Name == name {
			// Add camera pointer to sub interface(s).
			cam.server = c.server
			cam.PTZ.camera = cam
			return cam
		}
	}
	return nil
}

// StreamVideo streams a segment of video from a camera using FFMPEG.
func (c *Camera) StreamVideo(ops *VidOps, length time.Duration, maxsize int64) (io.ReadCloser, error) {
	f := ffmpeg.Get(&ffmpeg.VidOps{
		Encoder: Encoder,
		Time:    int(length.Seconds()),
		Audio:   true,    // Sure why not.
		Size:    maxsize, // max file size (always goes over). use 2000000 for 2.5MB
		Copy:    true,    // Always copy securityspy RTSP urls.
	})
	params := c.nakeRequestParams(ops)
	params.Set("auth", c.server.authB64)
	params.Set("codec", "h264")
	// This is kinda crude, but will handle 99%.
	url := strings.Replace(c.server.baseURL, "http", "rtsp", 1) + "++stream"
	// RTSP doesn't rewally work with HTTPS, and FFMPEG doesn't care about the cert.
	args, video, err := f.GetVideo(url+"?"+params.Encode(), c.Name)
	return video, errors.Wrap(err, strings.Replace(args, "\n", " ", -1))
}

// SaveVideo saves a segment of video from a camera to a file using FFMPEG.
func (c *Camera) SaveVideo(ops *VidOps, length time.Duration, maxsize int64, outputFile string) error {
	if _, err := os.Stat(outputFile); !os.IsNotExist(err) {
		return ErrorPathExists
	}
	f := ffmpeg.Get(&ffmpeg.VidOps{
		Encoder: Encoder,
		Time:    int(length.Seconds()),
		Audio:   true,
		Size:    maxsize, // max file size (always goes over). use 2000000 for 2.5MB
		Copy:    true,    // Always copy securityspy RTSP urls.
	})
	params := c.nakeRequestParams(ops)
	params.Set("auth", c.server.authB64)
	params.Set("codec", "h264")
	// This is kinda crude, but will handle 99%.
	url := strings.Replace(c.server.baseURL, "http", "rtsp", 1) + "++stream"
	_, out, err := f.SaveVideo(url+"?"+params.Encode(), outputFile, c.Name)
	return errors.Wrap(err, strings.Replace(out, "\n", " ", -1))
}

// StreamMJPG makes a web request to retreive a motion JPEG stream.
// Returns an io.ReadCloser that will (hopefully) never end.
func (c *Camera) StreamMJPG(ops *VidOps) (io.ReadCloser, error) {
	resp, err := c.server.secReq("++video", c.nakeRequestParams(ops), 10*time.Second)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// StreamH264 makes a web request to retreive an H264 stream.
// Returns an io.ReadCloser that will (hopefully) never end.
func (c *Camera) StreamH264(ops *VidOps) (io.ReadCloser, error) {
	resp, err := c.server.secReq("++stream", c.nakeRequestParams(ops), 10*time.Second)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// StreamG711 makes a web request to retreive an G711 audio stream.
// Returns an io.ReadCloser that will (hopefully) never end.
func (c *Camera) StreamG711() (io.ReadCloser, error) {
	resp, err := c.server.secReq("++audio", c.nakeRequestParams(nil), 10*time.Second)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// PostG711 makes a POST request to send audio to a camera with a speaker.
// Accepts an io.ReadCloser that will be closed. Probably an open file.
func (c *Camera) PostG711(audio io.ReadCloser) error {
	if audio == nil {
		return nil
	}
	defer func() {
		_ = audio.Close()
	}()
	return nil
	// Incomplete.
	// No helper methods for POST, so this is going to take a few more pieces.
}

// GetJPEG returns a picture from a camera.
func (c *Camera) GetJPEG(ops *VidOps) (image.Image, error) {
	ops.FPS = -1 // not used for single image
	resp, err := c.server.secReq("++image", c.nakeRequestParams(ops), 10*time.Second)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	jpgImage, err := jpeg.Decode(resp.Body)
	if err != nil {
		return nil, err
	}
	return jpgImage, err
}

// SaveJPEG gets a picture from a camera and puts it in a file.
func (c *Camera) SaveJPEG(ops *VidOps, path string) error {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return ErrorPathExists
	}
	ops.FPS = -1 // not used for single image
	jpgImage, err := c.GetJPEG(ops)
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	return jpeg.Encode(f, jpgImage, nil)
}

// ToggleContinuous arms (true) or disarms (false).
func (c *Camera) ToggleContinuous(arm CameraArmMode) error {
	params := make(url.Values)
	params.Set("arm", string(arm))
	return c.server.simpleReq("++ssControlContinuous", params, c.Number)
}

// ToggleActions arms (true) or disarms (false).
func (c *Camera) ToggleActions(arm CameraArmMode) error {
	params := make(url.Values)
	params.Set("arm", string(arm))
	return c.server.simpleReq("++ssControlActions", params, c.Number)
}

// ToggleMotion arms (true) or disarms (false).
func (c *Camera) ToggleMotion(arm CameraArmMode) error {
	params := make(url.Values)
	params.Set("arm", string(arm))
	return c.server.simpleReq("++ssControlMotionCapture", params, c.Number)
}

// TriggerMotion sets a camera as currently seeing motion.
// Other actions likely occur because of this!
func (c *Camera) TriggerMotion() error {
	return c.server.simpleReq("++triggermd", make(url.Values), c.Number)
}

// SetSchedule configures a camera mode's primary schedule.
func (c *Camera) SetSchedule(mode CameraMode, schedule Schedule) error {
	params := make(url.Values)
	params.Set("mode", string(mode))
	params.Set("id", strconv.Itoa(schedule.ID))
	return c.server.simpleReq("++ssSetSchedule", params, c.Number)
}

// SetScheduleOverride temporarily overrides a camera mode's primary schedule.
func (c *Camera) SetScheduleOverride(mode CameraMode, schedule Schedule) error {
	params := make(url.Values)
	params.Set("mode", string(mode))
	params.Set("id", strconv.Itoa(schedule.ID))
	return c.server.simpleReq("++ssSetOverride", params, c.Number)
}

/* INTERFACE HELPER METHODS FOLLOW */

// nakeRequestParams converts passed in ops to url.Values
func (c *Camera) nakeRequestParams(ops *VidOps) url.Values {
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
	if ops.Quality > 100 {
		ops.Quality = 100
	}
	if ops.Quality > 0 {
		params.Set("quality", strconv.Itoa(ops.Quality))
	}
	if ops.FPS > 0 {
		if ops.FPS > 60 {
			ops.FPS = 60
		}
		params.Set("req_fps", strconv.Itoa(ops.FPS))
	}
	return params
}
