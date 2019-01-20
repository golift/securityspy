package securityspy

import (
	"net/url"
	"strconv"
)

// ptzInterface powers the PTZ interface.
// It's really an extension of the camera interface.
type ptzInterface struct {
	Capabilities ptzCapabilities
	*camera
}

// PTZ interface provides access to camera PTZ controls.
type PTZ interface {
	Left() error
	Right() error
	Up() error
	Down() error
	UpLeft() error
	DownLeft() error
	UpRight() error
	DownRight() error
	Zoom(in bool) error
	Preset(preset Preset) error
	PresetSave(preset Preset) error
	Stop() error
	Caps() ptzCapabilities
}

// PTZcommand are the possible PTZ commands.
type PTZcommand int

// Bitmask values for PTZ Capabilities.
const (
	ptzPanTilt int = 1 << iota // 1
	ptzHome                    // 2
	ptzZoom                    // 4
	ptzPresets                 // 8
	ptzSpeed                   // 16
)

// PTZcommand in list form
const (
	PTZcommandLeft        PTZcommand = 1
	PTZcommandRight       PTZcommand = 2
	PTZcommandUp          PTZcommand = 3
	PTZcommandDown        PTZcommand = 4
	PTZcommandZoomIn      PTZcommand = 5
	PTZcommandZoomOut     PTZcommand = 6
	PTZcommandHome        PTZcommand = 7
	PTZcommandUpLeft      PTZcommand = 8
	PTZcommandUpRight     PTZcommand = 9
	PTZcommandDownLeft    PTZcommand = 10
	PTZcommandDownRight   PTZcommand = 11
	PTZcommandPreset1     PTZcommand = 12
	PTZcommandPreset2     PTZcommand = 13
	PTZcommandPreset3     PTZcommand = 14
	PTZcommandPreset4     PTZcommand = 15
	PTZcommandPreset5     PTZcommand = 16
	PTZcommandPreset6     PTZcommand = 17
	PTZcommandPreset7     PTZcommand = 18
	PTZcommandPreset8     PTZcommand = 19
	PTZcommandStop        PTZcommand = 99
	PTZcommandSavePreset1 PTZcommand = 112
	PTZcommandSavePreset2 PTZcommand = 113
	PTZcommandSavePreset3 PTZcommand = 114
	PTZcommandSavePreset4 PTZcommand = 115
	PTZcommandSavePreset5 PTZcommand = 116
	PTZcommandSavePreset6 PTZcommand = 117
	PTZcommandSavePreset7 PTZcommand = 118
	PTZcommandSavePreset8 PTZcommand = 119
)

// ptzCapabilities are what "things" a camera can do.
type ptzCapabilities struct {
	PanTilt bool
	Home    bool
	Zoom    bool
	Presets bool
	Speed   bool // This is missing full documentation; may not be accurate.
}

/* PTZ-specific concourse methods are at the top. */

// PTZ provides PTZ capabalities of a camera, such as panning, tilting, zomming, speed control, presets, home, etc.
func (c *camera) PTZ() PTZ {
	return &ptzInterface{
		Capabilities: ptzCapabilities{
			// Unmask them bits.
			PanTilt: c.Camera.PTZcapabilities&ptzPanTilt == ptzPanTilt,
			Home:    c.Camera.PTZcapabilities&ptzHome == ptzHome,
			Zoom:    c.Camera.PTZcapabilities&ptzZoom == ptzZoom,
			Presets: c.Camera.PTZcapabilities&ptzPresets == ptzPresets,
			Speed:   c.Camera.PTZcapabilities&ptzSpeed == ptzSpeed,
		},
		camera: c,
	}
}

/* Camera Interface, PTZ-specific methods follow. */

// Caps returns the supported PTZ methods.
func (c *ptzInterface) Caps() ptzCapabilities {
	return c.Capabilities
}

// Left sends a camera to the left one click.
func (c *ptzInterface) Left() error {
	return c.ptzReq(PTZcommandLeft)
}

// Right sends a camera to the right one click.
func (c *ptzInterface) Right() error {
	return c.ptzReq(PTZcommandRight)
}

// Up sends a camera to the sky one click.
func (c *ptzInterface) Up() error {
	return c.ptzReq(PTZcommandUp)
}

// Down puts a camera in time. no, really, it makes it look down one click.
func (c *ptzInterface) Down() error {
	return c.ptzReq(PTZcommandDown)
}

// UpLeft will send a camera up and to the left a click.
func (c *ptzInterface) UpLeft() error {
	return c.ptzReq(PTZcommandUpLeft)
}

// DownLeft sends a camera down and to the left a click.
func (c *ptzInterface) DownLeft() error {
	return c.ptzReq(PTZcommandDownLeft)
}

// UpRight sends a camera up and to the right. like it's 1999.
func (c *ptzInterface) UpRight() error {
	return c.ptzReq(PTZcommandRight)
}

// DownRight is sorta like making the camera do a dab.
func (c *ptzInterface) DownRight() error {
	return c.ptzReq(PTZcommandDownRight)
}

// Zoom makes a camera zoom in (true) or out (false).
func (c *ptzInterface) Zoom(in bool) error {
	if in {
		return c.ptzReq(PTZcommandZoomIn)
	}
	return c.ptzReq(PTZcommandZoomOut)
}

// Preset instructs a preset to be used. it just might work!
func (c *ptzInterface) Preset(preset Preset) error {
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

// PresetSave instructs a preset to be saved. good luck!
func (c *ptzInterface) PresetSave(preset Preset) error {
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

// Stop instructs a camera to stop moving. That is, if you have a camera
// cool enough to support continuous motion. Most do not, so sadly this is
// unlikely to be useful to you.
func (c *ptzInterface) Stop() error {
	return c.ptzReq(PTZcommandStop)
}

/* INTERFACE HELPER METHODS FOLLOW */

// ptzReq wraps all the ptz-specific calls.
func (c *ptzInterface) ptzReq(command PTZcommand) error {
	params := make(url.Values)
	params.Set("command", strconv.Itoa(int(command)))
	return c.simpleReq("++ptz/command", params)
}
