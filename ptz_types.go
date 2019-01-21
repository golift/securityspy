package securityspy

// PTZ are what "things" a camera can do.
type PTZ struct {
	camera     *Camera
	HasPanTilt bool
	HasHome    bool
	HasZoom    bool
	HasPresets bool
	HasSpeed   bool // This is missing full documentation; may not be accurate.
	rawCaps    int
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

// Bitmask values for PTZ Capabilities.
const (
	ptzPanTilt int = 1 << iota // 1
	ptzHome                    // 2
	ptzZoom                    // 4
	ptzPresets                 // 8
	ptzSpeed                   // 16
)

// ptzCommand are the possible PTZ commands.
type ptzCommand int

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
