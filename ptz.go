package securityspy

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"strconv"
)

var (
	// ErrPTZNotOK is returned for any command that has a successful web request,
	// but the reply does not end with the word OK.
	ErrPTZNotOK = fmt.Errorf("PTZ command not OK")
	// ErrPTZRange returns when a PTZ preset outside of 1-8 is provided.
	ErrPTZRange = fmt.Errorf("PTZ preset out of range 1-8")
)

// PTZ are what "things" a camera can do. Use the bound methods to interact
// with a camera's PTZ controls.
type PTZ struct {
	camera     *Camera
	rawCaps    int
	HasPanTilt bool // true if a camera can pan and tilt using PTZ controls.
	HasHome    bool // true if the camera supports the home position PTZ command.
	HasZoom    bool // true if the camera supports zooming in and out.
	HasPresets bool // true when the camera allows user-defined preset positions.
	Continuous bool // true if the camera supports continuous movement.
}

// PTZpreset locks our presets to a max of 8.
type PTZpreset rune

// Presets are 1 through 8. Use these constants as inputs to the PTZ methods.
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

// Bitmask values for PTZ Capabilities.
const (
	ptzPanTilt    int = 1 << iota // 1
	ptzHome                       // 2
	ptzZoom                       // 4
	ptzPresets                    // 8
	ptzContinuous                 // 16
)

// ptzCommand are the possible PTZ commands.
type ptzCommand int

// ptzCommand in list form
// These constants come directly from the SecuritySpy API doc.
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

// Down makes a camera look down one click.
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

// UpRight sends a camera up and to the right.
func (z *PTZ) UpRight() error {
	return z.ptzReq(ptzCommandUpRight)
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

// Preset instructs a a camera to move a preset position.
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
	default:
		return ErrPTZRange
	}
}

// PresetSave instructs a preset to be permanently saved. good luck!
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
	default:
		return ErrPTZRange
	}
}

// Stop instructs a camera to stop moving, assuming it supports continuous movement.
func (z *PTZ) Stop() error {
	return z.ptzReq(ptzCommandStop)
}

/* INTERFACE HELPER METHODS FOLLOW */

// ptzReq wraps all the ptz-specific calls.
func (z *PTZ) ptzReq(command ptzCommand) error {
	params := make(url.Values)
	params.Set("command", strconv.Itoa(int(command)))

	if err := z.camera.server.SimpleReq("++ptz/command", params, z.camera.Number); err != nil {
		return fmt.Errorf("ptz failed: %w", err)
	}

	return nil
}

// UnmarshalXML method converts ptzCapbilities bitmask from an XML payload into true/false abilities.
// This isn't a method you should ever call directly; it is only used during data initialization.
func (z *PTZ) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if err := d.DecodeElement(&z.rawCaps, &start); err != nil {
		return fmt.Errorf("ptz caps: %w", err)
	}

	z.HasPanTilt = z.rawCaps&ptzPanTilt == ptzPanTilt
	z.HasHome = z.rawCaps&ptzHome == ptzHome
	z.HasZoom = z.rawCaps&ptzZoom == ptzZoom
	z.HasPresets = z.rawCaps&ptzPresets == ptzPresets
	z.Continuous = z.rawCaps&ptzContinuous == ptzContinuous

	return nil
}
