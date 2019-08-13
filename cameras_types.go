package securityspy

import (
	"encoding/xml"
)

// Encoder is the path to ffmpeg.
var Encoder = "/usr/local/bin/ffmpeg"

// CameraArmMode locks arming to an integer of 0 or 1.
type CameraArmMode rune

// Arming is either 0 or 1.
// Use these constants as inputs to a camera's schedule methods.
const (
	CameraDisarm CameraArmMode = iota
	CameraArm
)

// VidOps are the frame options for a video that can be requested from SecuritySpy.
// This same data struct is used for capturing JPEG files, in that case FPS is discarded.
// Use this data type in the Camera methods that retrieve live videos/images.
type VidOps struct {
	Width   int
	Height  int
	FPS     int
	Quality int
}

// Cameras is an interface into the Camera system. Use the methods bound here
// to retrieve camera interfaces.
type Cameras struct {
	server  *Server
	Names   []string
	Numbers []int
}

// CameraSchedule contains schedule info for a camera's properties.
// This is assigned to Motion Capture, Continuous Capture and Actions.
type CameraSchedule struct {
	Name string
	ID   int
}

// UnmarshalXML stores a schedule ID into a CameraSchedule type.
// This isn't a method you should ever call directly; it is only used during data initialization.
func (bit *CameraSchedule) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return d.DecodeElement(&bit.ID, &start)
}

// Camera defines the data returned from the SecuritySpy API. This data is directly
// unmarshalled from the XML returned by the ++systemInfo method. Use the attached
// methods to control a camera in various ways.
type Camera struct {
	server              *Server
	Number              int            `xml:"number"`               // 0, 1, 2, 3, 4, 5, 6
	Connected           YesNoBool      `xml:"connected"`            // yes, yes, yes, yes, yes, ...
	Width               int            `xml:"width"`                // 2560, 2592, 2592, 3072, 2...
	Height              int            `xml:"height"`               // 1440, 1520, 1520, 2048, 1...
	Mode                YesNoBool      `xml:"mode"`                 // active, active, active, a...
	ModeC               YesNoBool      `xml:"mode-c"`               // armed, armed, armed, arme...
	ModeM               YesNoBool      `xml:"mode-m"`               // armed, armed, armed, arme...
	ModeA               YesNoBool      `xml:"mode-a"`               // armed, armed, armed, arme...
	HasAudio            YesNoBool      `xml:"hasaudio"`             // yes, yes, no, yes, yes, y...
	PTZ                 *PTZ           `xml:"ptzcapabilities"`      // 0, 0, 31, 0, 0, 0, 0
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
	MDsensitivity       int            `xml:"md_sensitivity"`       // 51, 50, 47, 50, 50, 50, 5...
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
	Permissions         int64          `xml:"permissions"`  // 63167, 63167, 62975, 6316...
	CapturePath         string         `xml:"capture-path"` // "/Volumes/Cameras/Porch" (v5+)
}
