package securityspy

import (
	"net/url"
	"strconv"
)

// PTZcommand are the possible PTZ commands.
type PTZcommand int

// Bitmask values for PTZ Capabilities.
const (
	PTZPanTilt = 1 << iota // 1
	PTZHome                // 2
	PTZZoom                // 4
	PTZPresets             // 8
	PTZSpeed               // 16
)

// PTZcommand in list form
const (
	PTZcommandLeft         PTZcommand = 1
	PTZcommandRight        PTZcommand = 2
	PTZcommandUp           PTZcommand = 3
	PTZcommandDown         PTZcommand = 4
	PTZcommandZoomIn       PTZcommand = 5
	PTZcommandZoomOut      PTZcommand = 6
	PTZcommandHome         PTZcommand = 7
	PTZcommandUpLeft       PTZcommand = 8
	PTZcommandUpRight      PTZcommand = 9
	PTZcommandDownLeft     PTZcommand = 10
	PTZcommandDownRight    PTZcommand = 11
	PTZcommandPreset1      PTZcommand = 12
	PTZcommandPreset2      PTZcommand = 13
	PTZcommandPreset3      PTZcommand = 14
	PTZcommandPreset4      PTZcommand = 15
	PTZcommandPreset5      PTZcommand = 16
	PTZcommandPreset6      PTZcommand = 17
	PTZcommandPreset7      PTZcommand = 18
	PTZcommandPreset8      PTZcommand = 19
	PTZcommandStopMovement PTZcommand = 99
	PTZcommandSavePreset1  PTZcommand = 112
	PTZcommandSavePreset2  PTZcommand = 113
	PTZcommandSavePreset3  PTZcommand = 114
	PTZcommandSavePreset4  PTZcommand = 115
	PTZcommandSavePreset5  PTZcommand = 116
	PTZcommandSavePreset6  PTZcommand = 117
	PTZcommandSavePreset7  PTZcommand = 118
	PTZcommandSavePreset8  PTZcommand = 119
)

// PTZSupports are what "things" a camera can do.
type PTZSupports struct {
	PanTilt bool
	Home    bool
	Zoom    bool
	Presets bool
	Speed   bool // This is missing full documentation; may not be accurate.
}

/* PTZ-specific concourse methods are at the top. */

// GetPTZ provides PTZ capabalities of a camera, such as panning, tilting, zomming, speed control, presets, home, etc.
func (c *CameraInterface) GetPTZ() PTZSupports {
	return PTZSupports{
		// Unmask them bits.
		PanTilt: c.Camera.PTZcapabilities&PTZPanTilt == PTZPanTilt,
		Home:    c.Camera.PTZcapabilities&PTZHome == PTZHome,
		Zoom:    c.Camera.PTZcapabilities&PTZZoom == PTZZoom,
		Presets: c.Camera.PTZcapabilities&PTZPresets == PTZPresets,
		Speed:   c.Camera.PTZcapabilities&PTZSpeed == PTZSpeed,
	}
}

/* Camera Interface, PTZ-specific methods follow. */

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

/* INTERFACE HELPER METHODS FOLLOW */

// ptzReq wraps all the ptz-specific calls.
func (c *CameraInterface) ptzReq(command PTZcommand) error {
	params := make(url.Values)
	params.Set("command", strconv.Itoa(int(command)))
	return c.simpleReq("/++ptz/command", params)
}
