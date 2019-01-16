package securityspy

import (
	"crypto/tls"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	encode "github.com/davidnewhall/go-securityspy/ffmpegencode"
)

/* Cameras-specific concourse methods are at the top. */

// GetCameras returns interfaces for every camera.
func (c *concourse) GetCameras() (cams []Camera) {
	for _, cam := range c.SystemInfo.CameraList.Cameras {
		cams = append(cams, &cameraInterface{Camera: cam, config: c.Config})
	}
	return
}

// GetCamera returns an interface for a single camera.
func (c *concourse) GetCamera(number int) Camera {
	for _, cam := range c.SystemInfo.CameraList.Cameras {
		if cam.Number == number {
			return &cameraInterface{Camera: cam, config: c.Config}
		}
	}
	return nil
}

// GetCameraByName returns an interface for a single camera, using the name.
func (c *concourse) GetCameraByName(name string) Camera {
	for _, cam := range c.SystemInfo.CameraList.Cameras {
		if cam.Name == name {
			return &cameraInterface{Camera: cam, config: c.Config}
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
	e := encode.Get(&encode.VidOps{
		Encoder: Encoder,
		Time:    int(length.Seconds()),
		Audio:   true,    // Sure why not.
		Size:    maxsize, // max file size (always goes over). use 2000000 for 2.5MB
		Copy:    true,    // Always copy securityspy RTSP urls.
	})
	params := nakeRequestParams(ops)
	params.Set("cameraNum", strconv.Itoa(c.Camera.Number))
	params.Set("auth", c.config.AuthB64)
	params.Set("codec", "h264")
	// This is kinda crude, but will handle 99%.
	url := strings.Replace(c.config.BaseURL, "http", "rtsp", 1) + "++stream"
	_, video, err := e.GetVideo(url+"?"+params.Encode(), c.Camera.Name)
	return video, err
}

// SaveVideo saves a segment of video from a camera to a file using FFMPEG.
func (c *cameraInterface) SaveVideo(ops *VidOps, length time.Duration, maxsize int64, outputFile string) error {
	if _, err := os.Stat(outputFile); !os.IsNotExist(err) {
		return ErrorPathExists
	}
	e := encode.Get(&encode.VidOps{
		Encoder: Encoder,
		Time:    int(length.Seconds()),
		Audio:   true,
		Size:    maxsize, // max file size (always goes over). use 2000000 for 2.5MB
		Copy:    true,    // Always copy securityspy RTSP urls.
	})
	params := nakeRequestParams(ops)
	params.Set("cameraNum", strconv.Itoa(c.Camera.Number))
	params.Set("auth", c.config.AuthB64)
	params.Set("codec", "h264")
	// This is kinda crude, but will handle 99%.
	url := strings.Replace(c.config.BaseURL, "http", "rtsp", 1) + "++stream"
	_, _, err := e.SaveVideo(url+"?"+params.Encode(), outputFile, c.Camera.Name)
	return err
}

// StreamMJPG makes a web request to retreive a motion JPEG stream.
// Returns an io.ReadCloser that will (hopefully) never end.
func (c *cameraInterface) StreamMJPG(ops *VidOps) (io.ReadCloser, error) {
	params := nakeRequestParams(ops)
	resp, err := c.camReq("++video", params)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// StreamH264 makes a web request to retreive an H264 stream.
// Returns an io.ReadCloser that will (hopefully) never end.
func (c *cameraInterface) StreamH264(ops *VidOps) (io.ReadCloser, error) {
	params := nakeRequestParams(ops)
	resp, err := c.camReq("++stream", params)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// StreamG711 makes a web request to retreive an G711 audio stream.
// Returns an io.ReadCloser that will (hopefully) never end.
func (c *cameraInterface) StreamG711() (io.ReadCloser, error) {
	resp, err := c.camReq("++audio", make(url.Values))
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
}

// GetJPEG returns a picture from a camera.
func (c *cameraInterface) GetJPEG(ops *VidOps) (image.Image, error) {
	ops.FPS = -1 // not used for single image
	params := nakeRequestParams(ops)
	resp, err := c.camReq("++image", params)
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
	// set the above to resp and turn it into an image.
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
	defer f.Close()
	return jpeg.Encode(f, jpgImage, nil)
}

// ContinuousCapture arms (true) or disarms (false).
func (c *cameraInterface) ContinuousCapture(arm CameraArmMode) error {
	params := make(url.Values)
	params.Set("arm", strconv.Itoa(int(arm)))
	return c.simpleReq("++ssControlContinuous", params)
}

// Actions arms (true) or disarms (false).
func (c *cameraInterface) Actions(arm CameraArmMode) error {
	params := make(url.Values)
	params.Set("arm", strconv.Itoa(int(arm)))
	return c.simpleReq("++ssControlActions", params)
}

// MotionCapture arms (true) or disarms (false).
func (c *cameraInterface) MotionCapture(arm CameraArmMode) error {
	params := make(url.Values)
	params.Set("arm", strconv.Itoa(int(arm)))
	return c.simpleReq("++ssControlMotionCapture", params)
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

// camReq is a helper function that formats the http request to SecuritySpy
func (c *cameraInterface) camReq(apiPath string, params url.Values) (*http.Response, error) {
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: !c.config.VerifySSL}}}
	req, err := http.NewRequest("GET", c.config.BaseURL+apiPath, nil)
	if err != nil {
		return nil, err
	}
	params.Set("cameraNum", strconv.Itoa(c.Camera.Number))
	params.Set("auth", c.config.AuthB64)
	req.URL.RawQuery = params.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *cameraInterface) simpleReq(apiURI string, params url.Values) error {
	resp, err := c.camReq(apiURI, params)
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
// it's only public in case it's useful.
func nakeRequestParams(ops *VidOps) url.Values {
	params := make(url.Values)
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
