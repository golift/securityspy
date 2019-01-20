package securityspy

import (
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	// Because I didn't feel like dealing with RTSP in Go. Maybe one day.
	ffmpeg "github.com/davidnewhall/go-securityspy/ffmpegencode"
	"github.com/pkg/errors"
)

/* Cameras-specific Server methods are at the top. */

// All returns interfaces for every camera.
func (c *cameras) All() (cams []Camera) {
	for _, cam := range c.systemInfo.CameraList.Cameras {
		cams = append(cams, &cameraInterface{Camera: cam, Server: c.Server})
	}
	return
}

// ByNum returns an interface for a single camera.
func (c *cameras) ByNum(number int) Camera {
	for _, cam := range c.systemInfo.CameraList.Cameras {
		if cam.Number == number {
			return &cameraInterface{Camera: cam, Server: c.Server}
		}
	}
	return nil
}

// ByName returns an interface for a single camera, using the name.
func (c *cameras) ByName(name string) Camera {
	for _, cam := range c.systemInfo.CameraList.Cameras {
		if cam.Name == name {
			return &cameraInterface{Camera: cam, Server: c.Server}
		}
	}
	return nil
}

/* Camera interface for CameraInterface follows */

// Device returns the camera's configuration from the server.
func (c *cameraInterface) Device() CameraDevice {
	return c.Camera
}

// StreamVideo streams a segment of video from a camera using FFMPEG.
func (c *cameraInterface) StreamVideo(ops *VidOps, length time.Duration, maxsize int64) (io.ReadCloser, error) {
	f := ffmpeg.Get(&ffmpeg.VidOps{
		Encoder: Encoder,
		Time:    int(length.Seconds()),
		Audio:   true,    // Sure why not.
		Size:    maxsize, // max file size (always goes over). use 2000000 for 2.5MB
		Copy:    true,    // Always copy securityspy RTSP urls.
	})
	params := c.nakeRequestParams(ops)
	params.Set("auth", c.authB64)
	params.Set("codec", "h264")
	// This is kinda crude, but will handle 99%.
	url := strings.Replace(c.baseURL, "http", "rtsp", 1) + "++stream"
	// RTSP doesn't rewally work with HTTPS, and FFMPEG doesn't care about the cert.
	args, video, err := f.GetVideo(url+"?"+params.Encode(), c.Camera.Name)
	return video, errors.Wrap(err, strings.Replace(args, "\n", " ", -1))
}

// SaveVideo saves a segment of video from a camera to a file using FFMPEG.
func (c *cameraInterface) SaveVideo(ops *VidOps, length time.Duration, maxsize int64, outputFile string) error {
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
	params.Set("auth", c.authB64)
	params.Set("codec", "h264")
	// This is kinda crude, but will handle 99%.
	url := strings.Replace(c.baseURL, "http", "rtsp", 1) + "++stream"
	_, out, err := f.SaveVideo(url+"?"+params.Encode(), outputFile, c.Camera.Name)
	return errors.Wrap(err, strings.Replace(out, "\n", " ", -1))
}

// StreamMJPG makes a web request to retreive a motion JPEG stream.
// Returns an io.ReadCloser that will (hopefully) never end.
func (c *cameraInterface) StreamMJPG(ops *VidOps) (io.ReadCloser, error) {
	resp, err := c.secReq("++video", c.nakeRequestParams(ops), 10*time.Second)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// StreamH264 makes a web request to retreive an H264 stream.
// Returns an io.ReadCloser that will (hopefully) never end.
func (c *cameraInterface) StreamH264(ops *VidOps) (io.ReadCloser, error) {
	resp, err := c.secReq("++stream", c.nakeRequestParams(ops), 10*time.Second)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// StreamG711 makes a web request to retreive an G711 audio stream.
// Returns an io.ReadCloser that will (hopefully) never end.
func (c *cameraInterface) StreamG711() (io.ReadCloser, error) {
	resp, err := c.secReq("++audio", c.nakeRequestParams(nil), 10*time.Second)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// PostG711 makes a POST request to send audio to a camera with a speaker.
// Accepts an io.ReadCloser that will be closed. Probably an open file.
func (c *cameraInterface) PostG711(audio io.ReadCloser) error {
	if audio == nil {
		return nil
	}
	defer func() {
		_ = audio.Close()
	}()
	return nil
	// Incomplete.
}

// GetJPEG returns a picture from a camera.
func (c *cameraInterface) GetJPEG(ops *VidOps) (image.Image, error) {
	ops.FPS = -1 // not used for single image
	resp, err := c.secReq("++image", c.nakeRequestParams(ops), 10*time.Second)
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
func (c *cameraInterface) SaveJPEG(ops *VidOps, path string) error {
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

// ContinuousCapture arms (true) or disarms (false).
func (c *cameraInterface) ContinuousCapture(arm CameraArmMode) error {
	return c.simpleReq("++ssControlContinuous", url.Values{"arm": []string{strconv.Itoa(int(arm))}})
}

// Actions arms (true) or disarms (false).
func (c *cameraInterface) Actions(arm CameraArmMode) error {
	return c.simpleReq("++ssControlActions", url.Values{"arm": []string{strconv.Itoa(int(arm))}})
}

// MotionCapture arms (true) or disarms (false).
func (c *cameraInterface) MotionCapture(arm CameraArmMode) error {
	return c.simpleReq("++ssControlMotionCapture", url.Values{"arm": []string{strconv.Itoa(int(arm))}})
}

// Size returns the camera frame size as a string.
func (c *cameraInterface) Size() string {
	return strconv.Itoa(c.Camera.Width) + "x" + strconv.Itoa(c.Camera.Height)
}

// Name returns the camera name.
func (c *cameraInterface) Name() string {
	return c.Camera.Name
}

// Number returns the camera number.
func (c *cameraInterface) Number() int {
	return c.Camera.Number
}

// Num returns the camera number as a string.
func (c *cameraInterface) Num() string {
	return strconv.Itoa(c.Camera.Number)
}

// TriggerMotion sets a camera as currently seeing motion.
// Other actions likely occur because of this!
func (c *cameraInterface) TriggerMotion() error {
	return c.simpleReq("++triggermd", make(url.Values))
}

/* INTERFACE HELPER METHODS FOLLOW */

// simpleReq performes HTTP req, checks for OK at end of output.
func (c *cameraInterface) simpleReq(apiURI string, params url.Values) error {
	params.Set("cameraNum", strconv.Itoa(c.Camera.Number))
	resp, err := c.secReq(apiURI, params, 10*time.Second)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if body, err := ioutil.ReadAll(resp.Body); err != nil {
		return err
	} else if !strings.HasSuffix(string(body), "OK") {
		return ErrorCmdNotOK
	}
	return nil
}

// nakeRequestParams converts passed in ops to url.Values
func (c *cameraInterface) nakeRequestParams(ops *VidOps) url.Values {
	params := make(url.Values)
	params.Set("cameraNum", strconv.Itoa(c.Camera.Number))
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
