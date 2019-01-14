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

	"github.com/pkg/errors"
)

// Cameras returns interfaces for every camera.
func (c *concourse) Cameras() (cams []Camera) {
	for _, cam := range c.SystemInfo.CameraContainer.Cameras {
		cams = append(cams, &cam)
	}
	return
}

// Camera returns an interface for a single camera.
func (c *concourse) Camera(number int) Camera {
	for _, cam := range c.SystemInfo.CameraContainer.Cameras {
		if cam.Cam.Number == number {
			return &cam
		}
	}
	return nil
}

// Camera returns an interface for a single camera.
func (c *concourse) CameraByName(name string) Camera {
	for _, cam := range c.SystemInfo.CameraContainer.Cameras {
		if cam.Cam.Name == name {
			return &cam
		}
	}
	return nil
}

// Conf returns the camera's configuration from the server.
func (c *CameraInterface) Conf() *CameraDevice {
	return c.Cam
}

// StreamMJPG makes a web request to retreive a motion JPEG stream.
// Returns an io.ReadCloser that will (hopefully) never end.
func (c *CameraInterface) StreamMJPG(width, height, quality, fps int) (io.ReadCloser, error) {
	params := makeQualityParams(width, height, quality, fps)
	resp, err := c.camReq("/++video", params)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// StreamH264 makes a web request to retreive an H264 stream.
// Returns an io.ReadCloser that will (hopefully) never end.
func (c *CameraInterface) StreamH264(width, height, quality, fps int) (io.ReadCloser, error) {
	params := makeQualityParams(width, height, quality, fps)
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
func (c *CameraInterface) GetJPEG(width, height, quality int) (image.Image, error) {
	params := makeQualityParams(width, height, quality, -1)
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
func (c *CameraInterface) SaveJPEG(width, height, quality int, path string) error {
	jpgImage, err := c.GetJPEG(width, height, quality)
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
	var params url.Values
	params.Set("arm", strconv.Itoa(int(arm)))
	return c.simpleReq("/++ssControlContinuous", params)
}

// Actions arms (true) or disarms (false).
func (c *CameraInterface) Actions(arm CameraArmOrDisarm) error {
	var params url.Values
	params.Set("arm", strconv.Itoa(int(arm)))
	return c.simpleReq("/++ssControlActions", params)
}

// MotionCapture arms (true) or disarms (false).
func (c *CameraInterface) MotionCapture(arm CameraArmOrDisarm) error {
	var params url.Values
	params.Set("arm", strconv.Itoa(int(arm)))
	return c.simpleReq("/++ssControlMotionCapture", params)
}

// Modes shows current modes for a camera.
func (c *CameraInterface) Modes() (continous, motion, actions bool) {
	return bool(c.Cam.ModeC), bool(c.Cam.ModeM), bool(c.Cam.ModeA)
}

// Name returns the camera name.
func (c *CameraInterface) Name() string {
	return c.Cam.Name
}

// TriggerMotion sets a camera as currently seeing motion.
// Other actions likely occur because of this!
func (c *CameraInterface) TriggerMotion() error {
	return c.simpleReq("/++triggermd", nil)
}

/* INTERFACE HELPER METHODS FOLLOW */

// secReq is a helper function that formats the http request to SecuritySpy
func (c *CameraInterface) camReq(apiPath string, params url.Values) (*http.Response, error) {
	a := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: !c.config.VerifySSL}}}
	req, err := http.NewRequest("GET", c.config.BaseURL+apiPath, nil)
	if err != nil {
		return nil, errors.Wrap(err, "http.NewRequest()")
	}
	params.Set("cameraNum", strconv.Itoa(c.Cam.Number))
	params.Set("auth", c.config.AuthB64)
	params.Set("format", "xml")
	req.URL.RawQuery = params.Encode()
	req.Header.Add("Accept", "application/xml")
	resp, err := a.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "http.Do(req)")
	}
	return resp, nil
}

func (c *CameraInterface) ptzReq(command PTZcommand) error {
	var params url.Values
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

func makeQualityParams(width, height, quality, fps int) url.Values {
	params := make(url.Values)
	if width != 0 {
		params.Set("width", strconv.Itoa(width))
	}
	if height != 0 {
		params.Set("height", strconv.Itoa(height))
	}
	if quality > 0 {
		if quality > 100 {
			quality = 100
		}
		params.Set("quality", strconv.Itoa(quality))
	}
	if fps > 0 {
		if fps > 60 {
			fps = 60
		}
		params.Set("req_fps", strconv.Itoa(fps))
	}
	return params
}
