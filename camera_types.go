package securityspy

import (
	"image"
	"io"
	"time"
)

// Encoder is the path to ffmpeg.
var Encoder = "/usr/local/bin/ffmpeg"

// ARM or DISARM a trigger
const (
	ErrorPTZNotOK = Error("PTZ command not OK")
	ErrorPTZRange = Error("PTZ preset out of range 1-8")
	ErrorARMNotOK = Error("arming camera unsuccessful")
)

// CameraArmOrDisarm locks arming to an integer of 0 or 1.
type CameraArmOrDisarm int

// Arming is either 0 or 1.
const (
	CameraDisarm CameraArmOrDisarm = iota
	CameraArm
)

// Preset locks our poresets to a max of 8
type Preset int

// Presets are 1 through 8.
const (
	_ Preset = iota
	Preset1
	Preset2
	Preset3
	Preset4
	Preset5
	Preset6
	Preset7
	Preset8
)

// PTZcommand are the possible PTZ commands.
type PTZcommand int

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

// PTZCapabilities are what "things" a camera can do.
type PTZCapabilities struct {
	PanTilt    bool
	Home       bool
	Zoom       bool
	Presets    bool
	Othersssss bool // This is missing full documentation.
}

// CameraInterface defines all the public and private elements of a camera.
type CameraInterface struct {
	Camera CameraDevice
	config *config // Server url, auth, ssl, etc
}

// CameraDevice defines the data returned from the SecuritySpy API.
type CameraDevice struct {
	Number              int       `xml:"number"`              // 0, 1, 2, 3, 4, 5, 6
	Connected           YesNoBool `xml:"connected"`           // yes, yes, yes, yes, yes, ...
	Width               int       `xml:"width"`               // 2560, 2592, 2592, 3072, 2...
	Height              int       `xml:"height"`              // 1440, 1520, 1520, 2048, 1...
	Mode                YesNoBool `xml:"mode"`                // active, active, active, a...
	ModeC               YesNoBool `xml:"mode-c"`              // armed, armed, armed, arme...
	ModeM               YesNoBool `xml:"mode-m"`              // armed, armed, armed, arme...
	ModeA               YesNoBool `xml:"mode-a"`              // armed, armed, armed, arme...
	HasAudio            YesNoBool `xml:"hasaudio"`            // yes, yes, no, yes, yes, y...
	PTZcapabilities     int       `xml:"ptzcapabilities"`     // 0, 0, 31, 0, 0, 0, 0
	TimeSinceLastFrame  Duration  `xml:"timesincelastframe"`  // 0, 0, 0, 0, 0, 0, 0
	TimeSinceLastMotion Duration  `xml:"timesincelastmotion"` // 689, 3796, 201, 12477, 15...
	DeviceName          string    `xml:"devicename"`          // ONVIF, ONVIF, ONVIF, ONVI...
	DeviceType          string    `xml:"devicetype"`          // Network, Network, Network...
	Address             string    `xml:"address"`             // 192.168.69.12, 192.168.69...
	Port                int       `xml:"port"`
	PortRtsp            int       `xml:"port-rtsp"`
	Request             string    `xml:"request"`
	Name                string    `xml:"name"`           // Porch, Door, Road, Garage...
	Overlay             YesNoBool `xml:"overlay"`        // no, no, no, no, no, no, n...
	OverlayText         string    `xml:"overlaytext"`    // +d, +d, +d, +d, +d, +d, +...
	Transformation      int       `xml:"transformation"` // 0, 0, 0, 0, 0, 0, 0
	AudioNetwork        YesNoBool `xml:"audio_network"`  // yes, yes, yes, yes, yes, ...
	AudioDeviceName     string    `xml:"audio_devicename"`
	MDenabled           YesNoBool `xml:"md_enabled"`        // yes, yes, yes, yes, yes, ...
	MDsensitivity       int       `xml:"md_sensitivity"`    // 51, 50, 47, 50, 50, 50, 5...  - this got returned twice under different key names.
	MDtriggerTimeX2     Duration  `xml:"md_triggertime_x2"` // 2, 2, 1, 2, 2, 2, 2
	MDcapture           YesNoBool `xml:"md_capture"`        // yes, yes, yes, yes, yes, ...
	MDcaptureFPS        int       `xml:"md_capturefps"`     // 20, 20, 20, 20, 20, 20, 2...
	MDpreCapture        Duration  `xml:"md_precapture"`     // 3, 4, 3, 3, 3, 2, 0
	MDpostCapture       Duration  `xml:"md_postcapture"`    // 10, 5, 5, 5, 5, 20, 15
	MDcaptureImages     YesNoBool `xml:"md_captureimages"`  // no, no, no, no, no, no, n...
	MDuploadImages      YesNoBool `xml:"md_uploadimages"`   // no, no, no, no, no, no, n...
	MDeecordAudio       YesNoBool `xml:"md_recordaudio"`    // yes, yes, yes, yes, yes, ...
	MDaudioTrigger      YesNoBool `xml:"md_audiotrigger"`   // no, no, no, no, no, no, n...
	MDaudioThreshold    int       `xml:"md_audiothreshold"` // 50, 50, 50, 50, 50, 50, 5...
	ActionScriptName    string    `xml:"action_scriptname"` // SS_SendiMessages.scpt, SS...
	ActionSoundName     string    `xml:"action_soundname"`
	ActionResetTime     Duration  `xml:"action_resettime"`     // 60, 60, 60, 60, 60, 60, 4...
	TLcapture           YesNoBool `xml:"tl_capture"`           // no, no, no, no, no, no, n...
	TLrecordAudio       YesNoBool `xml:"tl_recordaudio"`       // yes, yes, yes, yes, yes, ...
	CurrentFPS          float64   `xml:"current-fps"`          // 20.000, 20.000, 20.000, 2...
	ScheduleIDCC        int       `xml:"schedule-id-cc"`       // 1, 1, 1, 1, 1, 1, 0
	ScheduleIDMC        int       `xml:"schedule-id-mc"`       // 1, 1, 1, 1, 1, 1, 1
	ScheduleIDA         int       `xml:"schedule-id-a"`        // 1, 1, 1, 1, 1, 1, 1
	ScheduleOverrideCC  int       `xml:"schedule-override-cc"` // 0, 0, 0, 0, 0, 0, 0
	ScheduleOverrideMC  int       `xml:"schedule-override-mc"` // 0, 0, 0, 0, 0, 0, 0
	ScheduleOverrideA   int       `xml:"schedule-override-a"`  // 0, 0, 0, 0, 0, 0, 0
	PresetName1         string    `xml:"preset-name-1"`
	PresetName2         string    `xml:"preset-name-2"`
	PresetName3         string    `xml:"preset-name-3"`
	PresetName4         string    `xml:"preset-name-4"`
	PresetName5         string    `xml:"preset-name-5"`
	PresetName6         string    `xml:"preset-name-6"`
	PresetName7         string    `xml:"preset-name-7"`
	PresetName8         string    `xml:"preset-name-8"`
	Permissions         int64     `xml:"permissions"` // 63167, 63167, 62975, 6316...
}

// The Camera interface is used to manipulate and acquire data from cameras.
type Camera interface {
	Cam() CameraDevice
	Name() (name string)
	Number() int
	Num() string
	Size() string
	StreamMJPG(ops *VidOps) (video io.ReadCloser, err error)
	StreamH264(ops *VidOps) (video io.ReadCloser, err error)
	StreamG711() (audio io.ReadCloser, err error)
	StreamVideo(ops *VidOps, length time.Duration, maxSize int64) (video io.ReadCloser, err error)
	SaveVideo(ops *VidOps, length time.Duration, maxSize int64, outputFile string) error
	PostG711(audio io.ReadCloser) error
	GetJPEG(ops *VidOps) (image.Image, error)
	SaveJPEG(ops *VidOps, path string) error
	GetPTZ() (PTZCapabilities, error)
	PTLeft() error
	PTRight() error
	PTUp() error
	PTDown() error
	PTUpLeft() error
	PTDownLeft() error
	PTUpRight() error
	PTDownRight() error
	PTZoom(in bool) error
	PTZPreset(preset Preset) error
	PTZPresetSave(preset Preset) error
	PTZStop() error
	ContinuousCapture(arm CameraArmOrDisarm) error
	Actions(arm CameraArmOrDisarm) error
	MotionCapture(arm CameraArmOrDisarm) error
	TriggerMotion() error
}