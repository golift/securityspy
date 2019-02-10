package securityspy

import (
	"encoding/xml"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
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
	}
	return ErrorPTZRange
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
	}
	return ErrorPTZRange
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
	return z.camera.server.api.simpleReq("++ptz/command", params, z.camera.Number)
}

// UnmarshalXML method converts ptzCapbilities bitmask from an XML payload into true/false abilities.
// This isn't a method you should ever call directly; it is only used during data initialization.
func (z *PTZ) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if err := d.DecodeElement(&z.rawCaps, &start); err != nil {
		return errors.Wrap(err, "ptz caps")
	}
	z.HasPanTilt = z.rawCaps&ptzPanTilt == ptzPanTilt
	z.HasHome = z.rawCaps&ptzHome == ptzHome
	z.HasZoom = z.rawCaps&ptzZoom == ptzZoom
	z.HasPresets = z.rawCaps&ptzPresets == ptzPresets
	z.Continuous = z.rawCaps&ptzContinuous == ptzContinuous
	return nil
}
