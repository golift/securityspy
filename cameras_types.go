package securityspy

import (
	"encoding/xml"
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
	ErrorCmdNotOK = Error("command unsuccessful")
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
	_ Preset = iota // skip 0
	Preset1
	Preset2
	Preset3
	Preset4
	Preset5
	Preset6
	Preset7
	Preset8
)

// CameraSchedule contains schedule info for a camera properties.
type CameraSchedule struct {
	Name string
	ID   int
}

// UnmarshalXML stores a schedule ID into a CameraSchedule type.
func (bit *CameraSchedule) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return d.DecodeElement(&bit.ID, &start)
}

// VidOps are the options for a video that can be requested from SecuritySpy
type VidOps struct {
	Width   int
	Height  int
	FPS     int
	Quality int
}

// CameraInterface defines all the public and private elements of a camera.
type CameraInterface struct {
	Camera CameraDevice
	config *config // Server url, auth, ssl, etc
}

// CameraDevice defines the data returned from the SecuritySpy API.
type CameraDevice struct {
	Number              int            `xml:"number"`               // 0, 1, 2, 3, 4, 5, 6
	Connected           YesNoBool      `xml:"connected"`            // yes, yes, yes, yes, yes, ...
	Width               int            `xml:"width"`                // 2560, 2592, 2592, 3072, 2...
	Height              int            `xml:"height"`               // 1440, 1520, 1520, 2048, 1...
	Mode                YesNoBool      `xml:"mode"`                 // active, active, active, a...
	ModeC               YesNoBool      `xml:"mode-c"`               // armed, armed, armed, arme...
	ModeM               YesNoBool      `xml:"mode-m"`               // armed, armed, armed, arme...
	ModeA               YesNoBool      `xml:"mode-a"`               // armed, armed, armed, arme...
	HasAudio            YesNoBool      `xml:"hasaudio"`             // yes, yes, no, yes, yes, y...
	PTZcapabilities     int            `xml:"ptzcapabilities"`      // 0, 0, 31, 0, 0, 0, 0
	TimeSinceLastFrame  Duration       `xml:"timesincelastframe"`   // 0, 0, 0, 0, 0, 0, 0
	TimeSinceLastMotion Duration       `xml:"timesincelastmotion"`  // 689, 3796, 201, 12477, 15...
	DeviceName          string         `xml:"devicename"`           // ONVIF, ONVIF, ONVIF, ONVI...
	DeviceType          string         `xml:"devicetype"`           // Network, Network, Network...
	Address             string         `xml:"address"`              // 192.168.69.12, 192.168.69...
	Port                int            `xml:"port"`                 // 80, 80, 80, 0=80
	PortRTSP            int            `xml:"port-rtsp"`            // 554, 0=554
	Request             string         `xml:"request"`              // /some/rtsp/math (manual only)
	Name                string         `xml:"name"`                 // Porch, Door, Road, Garage...
	Overlay             YesNoBool      `xml:"overlay"`              // no, no, no, no, no, no, n...
	OverlayText         string         `xml:"overlaytext"`          // +d, +d, +d, +d, +d, +d, +...
	Transformation      int            `xml:"transformation"`       // 0, 1, 2, 3, 4, 5
	AudioNetwork        YesNoBool      `xml:"audio_network"`        // yes, yes, yes, yes, yes, ...
	AudioDeviceName     string         `xml:"audio_devicename"`     // Another Camera
	MDenabled           YesNoBool      `xml:"md_enabled"`           // yes, yes, yes, yes, yes, ...
	MDsensitivity       int            `xml:"md_sensitivity"`       // 51, 50, 47, 50, 50, 50, 5...  - this got returned twice under different key names.
	MDtriggerTimeX2     Duration       `xml:"md_triggertime_x2"`    // 2, 2, 1, 2, 2, 2, 2
	MDcapture           YesNoBool      `xml:"md_capture"`           // yes, yes, yes, yes, yes, ...
	MDcaptureFPS        float64        `xml:"md_capturefps"`        // 20, 20, 20, 20, 20, 20, 2...
	MDpreCapture        Duration       `xml:"md_precapture"`        // 3, 4, 3, 3, 3, 2, 0
	MDpostCapture       Duration       `xml:"md_postcapture"`       // 10, 5, 5, 5, 5, 20, 15
	MDcaptureImages     YesNoBool      `xml:"md_captureimages"`     // no, no, no, no, no, no, n...
	MDuploadImages      YesNoBool      `xml:"md_uploadimages"`      // no, no, no, no, no, no, n...
	MDeecordAudio       YesNoBool      `xml:"md_recordaudio"`       // yes, yes, yes, yes, yes, ...
	MDaudioTrigger      YesNoBool      `xml:"md_audiotrigger"`      // no, no, no, no, no, no, n...
	MDaudioThreshold    int            `xml:"md_audiothreshold"`    // 50, 50, 50, 50, 50, 50, 5...
	ActionScriptName    string         `xml:"action_scriptname"`    // SS_SendiMessages.scpt, SS...
	ActionSoundName     string         `xml:"action_soundname"`     // sound_file_name
	ActionResetTime     Duration       `xml:"action_resettime"`     // 60, 60, 60, 60, 60, 60, 4...
	TLcapture           YesNoBool      `xml:"tl_capture"`           // no, no, no, no, no, no, n...
	TLrecordAudio       YesNoBool      `xml:"tl_recordaudio"`       // yes, yes, yes, yes, yes, ...
	CurrentFPS          float64        `xml:"current-fps"`          // 20.000, 20.000, 20.000, 2...
	ScheduleIDCC        CameraSchedule `xml:"schedule-id-cc"`       // 1, 1, 1, 1, 1, 1, 0
	ScheduleIDMC        CameraSchedule `xml:"schedule-id-mc"`       // 1, 1, 1, 1, 1, 1, 1
	ScheduleIDA         CameraSchedule `xml:"schedule-id-a"`        // 1, 1, 1, 1, 1, 1, 1
	ScheduleOverrideCC  CameraSchedule `xml:"schedule-override-cc"` // 0, 0, 0, 0, 0, 0, 0
	ScheduleOverrideMC  CameraSchedule `xml:"schedule-override-mc"` // 0, 0, 0, 0, 0, 0, 0
	ScheduleOverrideA   CameraSchedule `xml:"schedule-override-a"`  // 0, 0, 0, 0, 0, 0, 0
	PresetName1         string         `xml:"preset-name-1"`
	PresetName2         string         `xml:"preset-name-2"`
	PresetName3         string         `xml:"preset-name-3"`
	PresetName4         string         `xml:"preset-name-4"`
	PresetName5         string         `xml:"preset-name-5"`
	PresetName6         string         `xml:"preset-name-6"`
	PresetName7         string         `xml:"preset-name-7"`
	PresetName8         string         `xml:"preset-name-8"`
	Permissions         int64          `xml:"permissions"` // 63167, 63167, 62975, 6316...
}

// The Camera interface is used to manipulate and acquire data from cameras.
type Camera interface {
	Device() CameraDevice
	Name() (name string)
	Number() int
	Num() string
	Size() string
	StreamVideo(ops *VidOps, length time.Duration, maxSize int64) (video io.ReadCloser, err error)
	SaveVideo(ops *VidOps, length time.Duration, maxSize int64, outputFile string) error
	StreamMJPG(ops *VidOps) (video io.ReadCloser, err error)
	StreamH264(ops *VidOps) (video io.ReadCloser, err error)
	StreamG711() (audio io.ReadCloser, err error)
	PostG711(audio io.ReadCloser) error
	GetJPEG(ops *VidOps) (image.Image, error)
	SaveJPEG(ops *VidOps, path string) error
	GetPTZ() (ptz PTZSupports)
	// like concourse.Files():
	// TODO: make GetPZT() return an interface with the following PTZ methods:
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
