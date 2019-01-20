package securityspy

import (
	"encoding/xml"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
)

// PTZ contains a struct of PTZ capabilities and an interface to control camera PTZ.
type PTZ struct {
	ptzCapabilities
	ptzControls
	camera *Camera
}

// ptzControls interface provides access to camera ptzControls controls.
type ptzControls interface {
	Left() error
	Right() error
	Up() error
	Down() error
	Zoom(in bool) error
	Home() error
	UpLeft() error
	DownLeft() error
	UpRight() error
	DownRight() error
	Preset(preset PTZpreset) error
	Stop() error
	PresetSave(preset PTZpreset) error
}

// PTZpreset locks our poresets to a max of 8
type PTZpreset int

// Presets are 1 through 8.
const (
	_ PTZpreset = iota // skip 0
	PTZpreset1
	PTZpreset2
	PTZpreset3
	PTZpreset4
	PTZpreset5
	PTZpreset6
	PTZpreset7
	PTZpreset8
)

// ptzCommand are the possible PTZ commands.
type ptzCommand int

// Bitmask values for PTZ Capabilities.
const (
	ptzPanTilt int = 1 << iota // 1
	ptzHome                    // 2
	ptzZoom                    // 4
	ptzPresets                 // 8
	ptzSpeed                   // 16
)

// ptzCommand in list form
const (
	ptzCommandLeft        ptzCommand = 1
	ptzCommandRight       ptzCommand = 2
	ptzCommandUp          ptzCommand = 3
	ptzCommandDown        ptzCommand = 4
	ptzCommandZoomIn      ptzCommand = 5
	ptzCommandZoomOut     ptzCommand = 6
	ptzCommandHome        ptzCommand = 7
	ptzCommandUpLeft      ptzCommand = 8
	ptzCommandUpRight     ptzCommand = 9
	ptzCommandDownLeft    ptzCommand = 10
	ptzCommandDownRight   ptzCommand = 11
	ptzCommandPreset1     ptzCommand = 12
	ptzCommandPreset2     ptzCommand = 13
	ptzCommandPreset3     ptzCommand = 14
	ptzCommandPreset4     ptzCommand = 15
	ptzCommandPreset5     ptzCommand = 16
	ptzCommandPreset6     ptzCommand = 17
	ptzCommandPreset7     ptzCommand = 18
	ptzCommandPreset8     ptzCommand = 19
	ptzCommandStop        ptzCommand = 99
	ptzCommandSavePreset1 ptzCommand = 112
	ptzCommandSavePreset2 ptzCommand = 113
	ptzCommandSavePreset3 ptzCommand = 114
	ptzCommandSavePreset4 ptzCommand = 115
	ptzCommandSavePreset5 ptzCommand = 116
	ptzCommandSavePreset6 ptzCommand = 117
	ptzCommandSavePreset7 ptzCommand = 118
	ptzCommandSavePreset8 ptzCommand = 119
)

// ptzCapabilities are what "things" a camera can do.
type ptzCapabilities struct {
	HasPanTilt bool
	HasHome    bool
	HasZoom    bool
	HasPresets bool
	HasSpeed   bool // This is missing full documentation; may not be accurate.
	rawCaps    int
}

/* Camera Interface, PTZ-specific methods follow. */

// Home sends a camera to the home position.
func (z *PTZ) Home() error {
	return z.ptzReq(ptzCommandHome)
}

// Left sends a camera to the left one click.
func (z *PTZ) Left() error {
	return z.ptzReq(ptzCommandLeft)
}

// Right sends a camera to the right one click.
func (z *PTZ) Right() error {
	return z.ptzReq(ptzCommandRight)
}

// Up sends a camera to the sky one click.
func (z *PTZ) Up() error {
	return z.ptzReq(ptzCommandUp)
}

// Down puts a camera in time. no, really, it makes it look down one click.
func (z *PTZ) Down() error {
	return z.ptzReq(ptzCommandDown)
}

// UpLeft will send a camera up and to the left a click.
func (z *PTZ) UpLeft() error {
	return z.ptzReq(ptzCommandUpLeft)
}

// DownLeft sends a camera down and to the left a click.
func (z *PTZ) DownLeft() error {
	return z.ptzReq(ptzCommandDownLeft)
}

// UpRight sends a camera up and to the right. like it's 1999.
func (z *PTZ) UpRight() error {
	return z.ptzReq(ptzCommandRight)
}

// DownRight is sorta like making the camera do a dab.
func (z *PTZ) DownRight() error {
	return z.ptzReq(ptzCommandDownRight)
}

// Zoom makes a camera zoom in (true) or out (false).
func (z *PTZ) Zoom(in bool) error {
	if in {
		return z.ptzReq(ptzCommandZoomIn)
	}
	return z.ptzReq(ptzCommandZoomOut)
}

// Preset instructs a preset to be used. it just might work!
func (z *PTZ) Preset(preset PTZpreset) error {
	switch preset {
	case PTZpreset1:
		return z.ptzReq(ptzCommandSavePreset1)
	case PTZpreset2:
		return z.ptzReq(ptzCommandSavePreset2)
	case PTZpreset3:
		return z.ptzReq(ptzCommandSavePreset3)
	case PTZpreset4:
		return z.ptzReq(ptzCommandSavePreset4)
	case PTZpreset5:
		return z.ptzReq(ptzCommandSavePreset5)
	case PTZpreset6:
		return z.ptzReq(ptzCommandSavePreset6)
	case PTZpreset7:
		return z.ptzReq(ptzCommandSavePreset7)
	case PTZpreset8:
		return z.ptzReq(ptzCommandSavePreset8)
	}
	return ErrorPTZRange
}

// PresetSave instructs a preset to be saved. good luck!
func (z *PTZ) PresetSave(preset PTZpreset) error {
	switch preset {
	case PTZpreset1:
		return z.ptzReq(ptzCommandPreset1)
	case PTZpreset2:
		return z.ptzReq(ptzCommandPreset2)
	case PTZpreset3:
		return z.ptzReq(ptzCommandPreset3)
	case PTZpreset4:
		return z.ptzReq(ptzCommandPreset4)
	case PTZpreset5:
		return z.ptzReq(ptzCommandPreset5)
	case PTZpreset6:
		return z.ptzReq(ptzCommandPreset6)
	case PTZpreset7:
		return z.ptzReq(ptzCommandPreset7)
	case PTZpreset8:
		return z.ptzReq(ptzCommandPreset8)
	}
	return ErrorPTZRange
}

// Stop instructs a camera to stop moving. That is, if you have a camera
// cool enough to support continuous motion. Most do not, so sadly this is
// unlikely to be useful to you.
func (z *PTZ) Stop() error {
	return z.ptzReq(ptzCommandStop)
}

/* INTERFACE HELPER METHODS FOLLOW */

// ptzReq wraps all the ptz-specific calls.
func (z *PTZ) ptzReq(command ptzCommand) error {
	params := make(url.Values)
	params.Set("command", strconv.Itoa(int(command)))
	return z.camera.simpleReq("++ptz/command", params)
}

// UnmarshalXML method converts ptzCapbilities bitmask into true/false abilities.
func (z *PTZ) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	err := d.DecodeElement(&z.rawCaps, &start)
	if err != nil {
		return errors.Wrap(err, "ptz caps")
	}
	z.HasPanTilt = z.rawCaps&ptzPanTilt == ptzPanTilt
	z.HasHome = z.rawCaps&ptzHome == ptzHome
	z.HasZoom = z.rawCaps&ptzZoom == ptzZoom
	z.HasPresets = z.rawCaps&ptzPresets == ptzPresets
	z.HasSpeed = z.rawCaps&ptzSpeed == ptzSpeed
	z.ptzControls = z // Prime the PTZ interface.
	return nil
}
