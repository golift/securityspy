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

// GetCameras returns interfaces for every camera.
func (c *concourse) GetCameras() (cams []Camera) {
	for _, cam := range c.SystemInfo.CameraList.Cameras {
		cams = append(cams, &CameraInterface{Cam: cam, config: c.Config})
	}
	return
}

// GetCamera returns an interface for a single camera.
func (c *concourse) GetCamera(number int) Camera {
	for _, cam := range c.SystemInfo.CameraList.Cameras {
		if cam.Number == number {
			return &CameraInterface{Cam: cam, config: c.Config}
		}
	}
	return nil
}

// GetCameraByName returns an interface for a single camera, using the name.
func (c *concourse) GetCameraByName(name string) Camera {
	for _, cam := range c.SystemInfo.CameraList.Cameras {
		if cam.Name == name {
			return &CameraInterface{Cam: cam, config: c.Config}
		}
	}
	return nil
}

// Conf returns the camera's configuration from the server.
func (c *CameraInterface) Conf() CameraDevice {
	return c.Cam
}

// StreamVideo streams a segment of video from a camera using FFMPEG.
func (c *CameraInterface) StreamVideo(ops *VidOps, length time.Duration, maxsize int64) (io.ReadCloser, error) {
	e := encode.Get(&encode.VidOps{
		Encoder: Encoder,
		Time:    int(length.Seconds()),
		Audio:   true,    // Sure why not.
		Size:    maxsize, // max file size (always goes over). use 2000000 for 2.5MB
		Copy:    true,    // Always copy securityspy RTSP urls.
	})
	params := makeQualityParams(ops)
	params.Set("cameraNum", strconv.Itoa(c.Cam.Number))
	params.Set("codec", "h264")
	// This is kinda crude, but will handle 99%.
	url := strings.Replace(c.config.BaseURL, "http", "rtsp", 1) + "/++stream"
	_, video, err := e.GetVideo(url+"?"+params.Encode(), c.Cam.Name)
	return video, err
}

// SaveVideo saves a segment of video from a camera to a file using FFMPEG.
func (c *CameraInterface) SaveVideo(ops *VidOps, length time.Duration, maxsize int64, outputFile string) error {
	e := encode.Get(&encode.VidOps{
		Encoder: Encoder,
		Time:    int(length.Seconds()),
		Audio:   true,
		Size:    maxsize, // max file size (always goes over). use 2000000 for 2.5MB
		Copy:    true,    // Always copy securityspy RTSP urls.
	})
	params := makeQualityParams(ops)
	params.Set("cameraNum", strconv.Itoa(c.Cam.Number))
	params.Set("codec", "h264")
	// This is kinda crude, but will handle 99%.
	url := strings.Replace(c.config.BaseURL, "http", "rtsp", 1) + "/++stream"
	_, _, err := e.SaveVideo(url+"?"+params.Encode(), outputFile, c.Cam.Name)
	return err
}

// StreamMJPG makes a web request to retreive a motion JPEG stream.
// Returns an io.ReadCloser that will (hopefully) never end.
func (c *CameraInterface) StreamMJPG(ops *VidOps) (io.ReadCloser, error) {
	params := makeQualityParams(ops)
	resp, err := c.camReq("/++video", params)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// StreamH264 makes a web request to retreive an H264 stream.
// Returns an io.ReadCloser that will (hopefully) never end.
func (c *CameraInterface) StreamH264(ops *VidOps) (io.ReadCloser, error) {
	params := makeQualityParams(ops)
	resp, err := c.camReq("/++stream", params)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// StreamG711 makes a web request to retreive an G711 audio stream.
// Returns an io.ReadCloser that will (hopefully) never end.
func (c *CameraInterface) StreamG711() (io.ReadCloser, error) {
	resp, err := c.camReq("/++audio", nil)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// PostG711 makes a POST request to send audio to a camera with a speaker.
// Accepts an io.ReadCloser that will be closed. Probably an open file.
func (c *CameraInterface) PostG711(audio io.ReadCloser) error {
	if audio == nil {
		return nil
	}
	defer func() {
		_ = audio.Close()
	}()
	return nil
}

// GetJPEG returns a picture from a camera.
func (c *CameraInterface) GetJPEG(ops *VidOps) (image.Image, error) {
	ops.FPS = -1 // not used for single image
	params := makeQualityParams(ops)
	resp, err := c.camReq("/++image", params)
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
func (c *CameraInterface) SaveJPEG(ops *VidOps, path string) error {
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

// GetPTZ provides PTZ capabalities of a camera, such as panning, tilting, zomming, speed control, presets, home, etc.
func (c *CameraInterface) GetPTZ() (PTZCapabilities, error) {
	// TODO: Figure out c.ptzcapabilities and bitwise unmask them.
	return PTZCapabilities{}, nil

}

// PTLeft sends a camera to the left one click.
func (c *CameraInterface) PTLeft() error {
	return c.ptzReq(PTZcommandLeft)
}

// PTRight sends a camera to the right one click.
func (c *CameraInterface) PTRight() error {
	return c.ptzReq(PTZcommandRight)
}

// PTUp sends a camera to the sky one click.
func (c *CameraInterface) PTUp() error {
	return c.ptzReq(PTZcommandUp)
}

// PTDown puts a camera in time. no, really, it makes it look down one click.
func (c *CameraInterface) PTDown() error {
	return c.ptzReq(PTZcommandDown)
}

// PTUpLeft will send a camera up and to the left a click.
func (c *CameraInterface) PTUpLeft() error {
	return c.ptzReq(PTZcommandUpLeft)
}

// PTDownLeft sends a camera down and to the left a click.
func (c *CameraInterface) PTDownLeft() error {
	return c.ptzReq(PTZcommandDownLeft)
}

// PTUpRight sends a camera up and to the right. like it's 1999.
func (c *CameraInterface) PTUpRight() error {
	return c.ptzReq(PTZcommandRight)
}

// PTDownRight is sorta like making the camera do a dab.
func (c *CameraInterface) PTDownRight() error {
	return c.ptzReq(PTZcommandDownRight)
}

// PTZoom makes a camera zoom in (true) or out (false).
func (c *CameraInterface) PTZoom(in bool) error {
	if in {
		return c.ptzReq(PTZcommandZoomIn)
	}
	return c.ptzReq(PTZcommandZoomOut)
}

// PTZPreset instructs a preset to be used. it just might work!
func (c *CameraInterface) PTZPreset(preset Preset) error {
	switch preset {
	case Preset1:
		return c.ptzReq(PTZcommandSavePreset1)
	case Preset2:
		return c.ptzReq(PTZcommandSavePreset2)
	case Preset3:
		return c.ptzReq(PTZcommandSavePreset3)
	case Preset4:
		return c.ptzReq(PTZcommandSavePreset4)
	case Preset5:
		return c.ptzReq(PTZcommandSavePreset5)
	case Preset6:
		return c.ptzReq(PTZcommandSavePreset6)
	case Preset7:
		return c.ptzReq(PTZcommandSavePreset7)
	case Preset8:
		return c.ptzReq(PTZcommandSavePreset8)
	}
	return ErrorPTZRange
}

// PTZPresetSave instructs a preset to be saved. good luck!
func (c *CameraInterface) PTZPresetSave(preset Preset) error {
	switch preset {
	case Preset1:
		return c.ptzReq(PTZcommandPreset1)
	case Preset2:
		return c.ptzReq(PTZcommandPreset2)
	case Preset3:
		return c.ptzReq(PTZcommandPreset3)
	case Preset4:
		return c.ptzReq(PTZcommandPreset4)
	case Preset5:
		return c.ptzReq(PTZcommandPreset5)
	case Preset6:
		return c.ptzReq(PTZcommandPreset6)
	case Preset7:
		return c.ptzReq(PTZcommandPreset7)
	case Preset8:
		return c.ptzReq(PTZcommandPreset8)
	}
	return ErrorPTZRange
}

// PTZStop instructs a camera to stop moving. That is, if you have a camera
// cool enough to support continuous motion. Most do not, so sadly this is
// unlikely to be useful to you.
func (c *CameraInterface) PTZStop() error {
	return c.ptzReq(PTZcommandStopMovement)
}

// ContinuousCapture arms (true) or disarms (false).
func (c *CameraInterface) ContinuousCapture(arm CameraArmOrDisarm) error {
	params := make(url.Values)
	params.Set("arm", strconv.Itoa(int(arm)))
	return c.simpleReq("/++ssControlContinuous", params)
}

// Actions arms (true) or disarms (false).
func (c *CameraInterface) Actions(arm CameraArmOrDisarm) error {
	params := make(url.Values)
	params.Set("arm", strconv.Itoa(int(arm)))
	return c.simpleReq("/++ssControlActions", params)
}

// MotionCapture arms (true) or disarms (false).
func (c *CameraInterface) MotionCapture(arm CameraArmOrDisarm) error {
	params := make(url.Values)
	params.Set("arm", strconv.Itoa(int(arm)))
	return c.simpleReq("/++ssControlMotionCapture", params)
}

// Size returns the camera frame size as a string.
func (c *CameraInterface) Size() string {
	return strconv.Itoa(c.Cam.Width) + "x" + strconv.Itoa(c.Cam.Height)
}

// Name returns the camera name.
func (c *CameraInterface) Name() string {
	return c.Cam.Name
}

// Number returns the camera number.
func (c *CameraInterface) Number() int {
	return c.Cam.Number
}

// Num returns the camera number as a string.
func (c *CameraInterface) Num() string {
	return strconv.Itoa(c.Cam.Number)
}

// TriggerMotion sets a camera as currently seeing motion.
// Other actions likely occur because of this!
func (c *CameraInterface) TriggerMotion() error {
	return c.simpleReq("/++triggermd", nil)
}

/* INTERFACE HELPER METHODS FOLLOW */

// camReq is a helper function that formats the http request to SecuritySpy
func (c *CameraInterface) camReq(apiPath string, params url.Values) (*http.Response, error) {
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: !c.config.VerifySSL}}}
	req, err := http.NewRequest("GET", c.config.BaseURL+apiPath, nil)
	if err != nil {
		return nil, err
	}
	if params == nil {
		params = make(url.Values)
	}
	params.Set("cameraNum", strconv.Itoa(c.Cam.Number))
	params.Set("auth", c.config.AuthB64)
	req.URL.RawQuery = params.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *CameraInterface) ptzReq(command PTZcommand) error {
	params := make(url.Values)
	params.Set("command", strconv.Itoa(int(command)))
	return c.simpleReq("/++ptz/command", params)
}

func (c *CameraInterface) simpleReq(apiURI string, params url.Values) error {
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
		return ErrorARMNotOK
	}
	return nil
}

// makeQualityParams converts passed in ops to url.Values
func makeQualityParams(ops *VidOps) url.Values {
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
