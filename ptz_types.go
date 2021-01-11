package securityspy

import "fmt"

var (
	// ErrorPTZNotOK is returned for any command that has a successful web request,
	// but the reply does not end with the word OK.
	ErrorPTZNotOK = fmt.Errorf("PTZ command not OK")

	// ErrorPTZRange returns when a PTZ preset outside of 1-8 is provided.
	ErrorPTZRange = fmt.Errorf("PTZ preset out of range 1-8")
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
