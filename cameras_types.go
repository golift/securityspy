package securityspy

import (
	"encoding/xml"
	"errors"
	"fmt"
)

// ErrUnsupported is returned when the SecuritySpy server does not implement an API endpoint.
// ToggleContinuous returns this on many builds where ++ssControlContinuous is missing;
// use SetSchedule with CameraModeContinuous instead.
var ErrUnsupported = errors.New("securityspy endpoint unsupported")

// ErrCameraModesStatus is returned when ++cameramodes returns a non-200 status.
var ErrCameraModesStatus = errors.New("camera modes request failed")

// ErrCameraModesParse is returned when ++cameramodes body cannot be parsed.
var ErrCameraModesParse = errors.New("camera modes response invalid")

// DefaultEncoder is the path to ffmpeg.
const DefaultEncoder = "/usr/local/bin/ffmpeg"

// CameraArmMode locks arming to an integer of 0 or 1.
type CameraArmMode rune

// Arming is either 0 or 1.
// Use these constants as inputs to a camera's schedule methods.
const (
	CameraDisarm CameraArmMode = '0'
	CameraArm    CameraArmMode = '1'
)

// VidOps are the frame options for a video that can be requested from SecuritySpy.
// This same data struct is used for capturing JPEG files, in that case FPS is discarded.
// Use this data type in the Camera methods that retrieve live videos/images.
type VidOps struct {
	// Optional width override for video stream (defaults to camera width).
	Width int
	// Optional height override for video stream (defaults to camera height).
	Height int
	// Optional frame rate override for video stream (defaults to camera frame rate).
	FPS int
	// Optional quality override for video stream (defaults to camera quality).
	Quality int
	// If true, use HTTP video endpoint instead of RTSP(S) stream endpoint.
	UseHTTP bool
	// Optional codec override for video stream (defaults to h264).
	// For ++video HTTP streams: jpeg, h264, h265, or h26x.
	VCodec string
	// Optional codec override for audio stream (defaults to aac).
	ACodec string
}

// MultiplexOps are options for the ++multiplex HTML grid page (v6+).
type MultiplexOps struct {
	Cameras  []int // camera numbers to include
	CropMode int   // 0 black bars, 1 crop, 2 stretch
	Format   int   // 0 JPEG, 1 H.264/H.265
	FPS      int   // max frame rate per feed
	Border   int   // border width in pixels
	CamInfo  bool  // include camera info bar
	HiRes    bool  // double-resolution feeds
}

// Cameras is an interface into the Camera system. Use the methods bound here
// to retrieve camera interfaces.
type Cameras struct {
	cameras []*Camera
	server  *Server
}

// CameraModes reports armed/disarmed status for each capture mode from ++cameramodes.
type CameraModes struct {
	Continuous string // "ARMED" or "DISARMED"
	Motion     string
	Actions    string
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
	if err := d.DecodeElement(&bit.ID, &start); err != nil {
		return fmt.Errorf("decoding xml: %w", err)
	}

	return nil
}

// Camera defines the data returned from the SecuritySpy API ++systemInfo method.
// Exported field names stay stable across v5/v6; XML unmarshaling accepts both schemas.
type Camera struct {
	server *Server

	Number              int
	Connected           YesNoBool
	Width               int
	Height              int
	Mode                YesNoBool // v5 "mode" (active); often empty on v6
	ModeC               YesNoBool
	ModeM               YesNoBool
	ModeA               YesNoBool
	HasAudio            YesNoBool
	PTZ                 *PTZ
	TimeSinceLastFrame  Duration
	TimeSinceLastMotion Duration
	DeviceName          string
	DeviceType          string
	Address             string
	Port                int
	PortRTSP            int
	Request             string // v5 manual RTSP path; v6 uses Path
	Name                string
	Overlay             YesNoBool
	OverlayText         string
	Transformation      int
	AudioNetwork        YesNoBool
	AudioDeviceName     string
	MDenabled           YesNoBool // v5; v6 uses MotionSensitivity / triggers
	MDsensitivity       int       // aliased from MotionSensitivity on v6
	MDtriggerTimeX2     Duration  // v5
	MDcapture           YesNoBool // v5
	MDcaptureFPS        float64   // v5 / mc-movie-fps on v6
	MDpreCapture        Duration
	MDpostCapture       Duration
	MDcaptureImages     YesNoBool
	MDuploadImages      YesNoBool
	MDeecordAudio       YesNoBool
	MDaudioTrigger      YesNoBool
	MDaudioThreshold    int
	ActionScriptName    string
	ActionSoundName     string
	ActionResetTime     Duration
	TLcapture           YesNoBool
	TLrecordAudio       YesNoBool
	CurrentFPS          float64
	ScheduleIDCC        CameraSchedule
	ScheduleIDMC        CameraSchedule
	ScheduleIDA         CameraSchedule
	ScheduleOverrideCC  CameraSchedule
	ScheduleOverrideMC  CameraSchedule
	ScheduleOverrideA   CameraSchedule
	PresetName1         string
	PresetName2         string
	PresetName3         string
	PresetName4         string
	PresetName5         string
	PresetName6         string
	PresetName7         string
	PresetName8         string
	PresetName9         string
	PresetName10        string
	Permissions         int64
	CapturePath         string

	// v6+ fields
	DataRate             int
	VideoFormat          string
	LastError            int
	LastErrorDescription string
	Brightness           int
	Contrast             int
	AudioFormat          string
	AudioSensitivity     int
	MotionSensitivity    int
	Path                 string
	CCMovie              YesNoBool
	CCImage              YesNoBool
	MCMovie              YesNoBool
	MCImage              YesNoBool
	MCMovieFPS           float64
	MCTriggerVideo       YesNoBool
	MCTriggerAudio       YesNoBool
	ATriggerVideo        YesNoBool
	ATriggerAudio        YesNoBool
	ActionSoundCam       string
	ActionSoundMac       string
	CustomModel          YesNoBool
	SinceLastCapture     Duration
	NetworkAudio         YesNoBool // mirrors AudioNetwork when set
}

// cameraXML holds both v5 and v6 ++systemInfo camera tags for dual unmarshaling.
type cameraXML struct {
	Number         int       `xml:"number"`
	Connected      YesNoBool `xml:"connected"`
	Name           string    `xml:"name"`
	Address        string    `xml:"address"`
	Port           int       `xml:"port"`
	PortRTSP       int       `xml:"port-rtsp"`
	Permissions    int64     `xml:"permissions"`
	Transformation int       `xml:"transformation"`
	CurrentFPS     float64   `xml:"current-fps"`
	OverlayText    string    `xml:"overlaytext"`
	OverlayTextV6  string    `xml:"overlay-text"`
	Overlay        YesNoBool `xml:"overlay"`

	// geometry
	WidthV5  int `xml:"width"`
	HeightV5 int `xml:"height"`
	WidthV6  int `xml:"video-width"`
	HeightV6 int `xml:"video-height"`

	// modes
	Mode    YesNoBool `xml:"mode"`
	ModeCV5 YesNoBool `xml:"mode-c"`
	ModeMV5 YesNoBool `xml:"mode-m"`
	ModeAV5 YesNoBool `xml:"mode-a"`
	ModeCV6 YesNoBool `xml:"cc-mode"`
	ModeMV6 YesNoBool `xml:"mc-mode"`
	ModeAV6 YesNoBool `xml:"a-mode"`

	HasAudioV5 YesNoBool `xml:"hasaudio"`
	HasAudioV6 YesNoBool `xml:"has-audio"`

	PTZV5 *PTZ `xml:"ptzcapabilities"`
	PTZV6 *PTZ `xml:"ptz-features"`

	TimeSinceLastFrameV5  Duration `xml:"timesincelastframe"`
	TimeSinceLastMotionV5 Duration `xml:"timesincelastmotion"`
	TimeSinceLastFrameV6  Duration `xml:"time-since-last-frame"`
	TimeSinceLastMotionV6 Duration `xml:"time-since-last-motion"`

	DeviceNameV5 string `xml:"devicename"`
	DeviceTypeV5 string `xml:"devicetype"`
	DeviceNameV6 string `xml:"device-name"`
	DeviceTypeV6 string `xml:"device-type"`

	Request string `xml:"request"`
	Path    string `xml:"path"`

	AudioNetworkV5    YesNoBool `xml:"audio_network"`
	AudioDeviceNameV5 string    `xml:"audio_devicename"`
	AudioNetworkV6    YesNoBool `xml:"network-audio"`
	AudioDeviceNameV6 string    `xml:"audio-device-name"`

	MDenabled        YesNoBool `xml:"md_enabled"`
	MDsensitivity    int       `xml:"md_sensitivity"`
	MDtriggerTimeX2  Duration  `xml:"md_triggertime_x2"`
	MDcapture        YesNoBool `xml:"md_capture"`
	MDcaptureFPS     float64   `xml:"md_capturefps"`
	MDpreCapture     Duration  `xml:"md_precapture"`
	MDpostCapture    Duration  `xml:"md_postcapture"`
	MDcaptureImages  YesNoBool `xml:"md_captureimages"`
	MDuploadImages   YesNoBool `xml:"md_uploadimages"`
	MDeecordAudio    YesNoBool `xml:"md_recordaudio"`
	MDaudioTrigger   YesNoBool `xml:"md_audiotrigger"`
	MDaudioThreshold int       `xml:"md_audiothreshold"`

	ActionScriptNameV5 string   `xml:"action_scriptname"`
	ActionSoundNameV5  string   `xml:"action_soundname"`
	ActionResetTimeV5  Duration `xml:"action_resettime"`
	ActionScriptV6     string   `xml:"a-script"`
	ActionSoundCam     string   `xml:"a-sound-cam"`
	ActionSoundMac     string   `xml:"a-sound-mac"`
	ActionResetTimeV6  Duration `xml:"a-reset-time"`
	ActionDelayTime    Duration `xml:"a-delay-time"`

	TLcapture     YesNoBool `xml:"tl_capture"`
	TLrecordAudio YesNoBool `xml:"tl_recordaudio"`

	ScheduleIDCCV5       CameraSchedule `xml:"schedule-id-cc"`
	ScheduleIDMCV5       CameraSchedule `xml:"schedule-id-mc"`
	ScheduleIDAV5        CameraSchedule `xml:"schedule-id-a"`
	ScheduleOverrideCCV5 CameraSchedule `xml:"schedule-override-cc"`
	ScheduleOverrideMCV5 CameraSchedule `xml:"schedule-override-mc"`
	ScheduleOverrideAV5  CameraSchedule `xml:"schedule-override-a"`
	ScheduleIDCCV6       CameraSchedule `xml:"cc-schedule-id"`
	ScheduleIDMCV6       CameraSchedule `xml:"mc-schedule-id"`
	ScheduleIDAV6        CameraSchedule `xml:"a-schedule-id"`
	ScheduleOverrideCCV6 CameraSchedule `xml:"cc-schedule-override"`
	ScheduleOverrideMCV6 CameraSchedule `xml:"mc-schedule-override"`
	ScheduleOverrideAV6  CameraSchedule `xml:"a-schedule-override"`

	PresetName1  string `xml:"preset-name-1"`
	PresetName2  string `xml:"preset-name-2"`
	PresetName3  string `xml:"preset-name-3"`
	PresetName4  string `xml:"preset-name-4"`
	PresetName5  string `xml:"preset-name-5"`
	PresetName6  string `xml:"preset-name-6"`
	PresetName7  string `xml:"preset-name-7"`
	PresetName8  string `xml:"preset-name-8"`
	PresetName9  string `xml:"preset-name-9"`
	PresetName10 string `xml:"preset-name-10"`

	CapturePathV5 string `xml:"capture-path"`
	CapturePathV6 string `xml:"storage-path"`

	DataRate             int       `xml:"data-rate"`
	VideoFormat          string    `xml:"video-format"`
	LastError            int       `xml:"last-error"`
	LastErrorDescription string    `xml:"last-error-description"`
	Brightness           int       `xml:"brightness"`
	Contrast             int       `xml:"contrast"`
	AudioFormat          string    `xml:"audio-format"`
	AudioSensitivity     int       `xml:"audio-sensitivity"`
	MotionSensitivity    int       `xml:"motion-sensitivity"`
	CCMovie              YesNoBool `xml:"cc-movie"`
	CCImage              YesNoBool `xml:"cc-image"`
	MCMovie              YesNoBool `xml:"mc-movie"`
	MCImage              YesNoBool `xml:"mc-image"`
	MCMovieFPS           float64   `xml:"mc-movie-fps"`
	MCMoviePre           Duration  `xml:"mc-movie-pre"`
	MCMoviePost          Duration  `xml:"mc-movie-post"`
	MCTriggerVideo       YesNoBool `xml:"mc-trigger-video"`
	MCTriggerAudio       YesNoBool `xml:"mc-trigger-audio"`
	ATriggerVideo        YesNoBool `xml:"a-trigger-video"`
	ATriggerAudio        YesNoBool `xml:"a-trigger-audio"`
	CustomModel          YesNoBool `xml:"custom-model"`
	SinceLastCapture     Duration  `xml:"since-last-capture"`
}

// UnmarshalXML decodes v5 or v6 ++systemInfo camera elements into Camera.
//
//nolint:cyclop,funlen // dual schema mapping
func (c *Camera) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var raw cameraXML
	if err := d.DecodeElement(&raw, &start); err != nil {
		return fmt.Errorf("decoding camera xml: %w", err)
	}

	isV6 := raw.WidthV6 != 0 || raw.HeightV6 != 0 || raw.CapturePathV6 != "" ||
		raw.DeviceNameV6 != "" || raw.ModeCV6.Txt != "" || raw.HasAudioV6.Txt != "" ||
		raw.PTZV6 != nil || raw.VideoFormat != "" || raw.StoragePathSet()

	c.Number = raw.Number
	c.Connected = raw.Connected
	c.Name = raw.Name
	c.Address = raw.Address
	c.Port = raw.Port
	c.PortRTSP = raw.PortRTSP
	c.Permissions = raw.Permissions
	c.Transformation = raw.Transformation
	c.CurrentFPS = raw.CurrentFPS
	c.Mode = raw.Mode
	c.Request = raw.Request
	c.Path = raw.Path
	c.Overlay = raw.Overlay
	c.OverlayText = firstNonEmpty(raw.OverlayTextV6, raw.OverlayText)

	if isV6 {
		c.Width = raw.WidthV6
		c.Height = raw.HeightV6
		c.ModeC = raw.ModeCV6
		c.ModeM = raw.ModeMV6
		c.ModeA = raw.ModeAV6
		c.HasAudio = raw.HasAudioV6
		c.PTZ = raw.PTZV6
		c.TimeSinceLastFrame = raw.TimeSinceLastFrameV6
		c.TimeSinceLastMotion = raw.TimeSinceLastMotionV6
		c.DeviceName = raw.DeviceNameV6
		c.DeviceType = raw.DeviceTypeV6
		c.AudioNetwork = raw.AudioNetworkV6
		c.AudioDeviceName = raw.AudioDeviceNameV6
		c.ActionScriptName = raw.ActionScriptV6
		c.ActionSoundName = raw.ActionSoundMac
		c.ActionResetTime = raw.ActionResetTimeV6
		c.ScheduleIDCC = raw.ScheduleIDCCV6
		c.ScheduleIDMC = raw.ScheduleIDMCV6
		c.ScheduleIDA = raw.ScheduleIDAV6
		c.ScheduleOverrideCC = raw.ScheduleOverrideCCV6
		c.ScheduleOverrideMC = raw.ScheduleOverrideMCV6
		c.ScheduleOverrideA = raw.ScheduleOverrideAV6
		c.CapturePath = raw.CapturePathV6
		c.MDsensitivity = raw.MotionSensitivity
		c.MotionSensitivity = raw.MotionSensitivity
		c.MDcaptureFPS = raw.MCMovieFPS
		c.MCMovieFPS = raw.MCMovieFPS
		c.MDpreCapture = raw.MCMoviePre
		c.MDpostCapture = raw.MCMoviePost
		c.MDcaptureImages = raw.MCImage
		c.MDaudioTrigger = raw.MCTriggerAudio
		c.MDaudioThreshold = raw.AudioSensitivity
		c.AudioSensitivity = raw.AudioSensitivity
	} else {
		c.Width = raw.WidthV5
		c.Height = raw.HeightV5
		c.ModeC = raw.ModeCV5
		c.ModeM = raw.ModeMV5
		c.ModeA = raw.ModeAV5
		c.HasAudio = raw.HasAudioV5
		c.PTZ = raw.PTZV5
		c.TimeSinceLastFrame = raw.TimeSinceLastFrameV5
		c.TimeSinceLastMotion = raw.TimeSinceLastMotionV5
		c.DeviceName = raw.DeviceNameV5
		c.DeviceType = raw.DeviceTypeV5
		c.AudioNetwork = raw.AudioNetworkV5
		c.AudioDeviceName = raw.AudioDeviceNameV5
		c.ActionScriptName = raw.ActionScriptNameV5
		c.ActionSoundName = raw.ActionSoundNameV5
		c.ActionResetTime = raw.ActionResetTimeV5
		c.ScheduleIDCC = raw.ScheduleIDCCV5
		c.ScheduleIDMC = raw.ScheduleIDMCV5
		c.ScheduleIDA = raw.ScheduleIDAV5
		c.ScheduleOverrideCC = raw.ScheduleOverrideCCV5
		c.ScheduleOverrideMC = raw.ScheduleOverrideMCV5
		c.ScheduleOverrideA = raw.ScheduleOverrideAV5
		c.CapturePath = raw.CapturePathV5
		c.MDenabled = raw.MDenabled
		c.MDsensitivity = raw.MDsensitivity
		c.MotionSensitivity = raw.MDsensitivity
		c.MDtriggerTimeX2 = raw.MDtriggerTimeX2
		c.MDcapture = raw.MDcapture
		c.MDcaptureFPS = raw.MDcaptureFPS
		c.MCMovieFPS = raw.MDcaptureFPS
		c.MDpreCapture = raw.MDpreCapture
		c.MDpostCapture = raw.MDpostCapture
		c.MDcaptureImages = raw.MDcaptureImages
		c.MDuploadImages = raw.MDuploadImages
		c.MDeecordAudio = raw.MDeecordAudio
		c.MDaudioTrigger = raw.MDaudioTrigger
		c.MDaudioThreshold = raw.MDaudioThreshold
		c.AudioSensitivity = raw.MDaudioThreshold
		c.TLcapture = raw.TLcapture
		c.TLrecordAudio = raw.TLrecordAudio
	}

	c.NetworkAudio = c.AudioNetwork
	c.ActionSoundCam = raw.ActionSoundCam
	c.ActionSoundMac = raw.ActionSoundMac
	c.PresetName1 = raw.PresetName1
	c.PresetName2 = raw.PresetName2
	c.PresetName3 = raw.PresetName3
	c.PresetName4 = raw.PresetName4
	c.PresetName5 = raw.PresetName5
	c.PresetName6 = raw.PresetName6
	c.PresetName7 = raw.PresetName7
	c.PresetName8 = raw.PresetName8
	c.PresetName9 = raw.PresetName9
	c.PresetName10 = raw.PresetName10
	c.DataRate = raw.DataRate
	c.VideoFormat = raw.VideoFormat
	c.LastError = raw.LastError
	c.LastErrorDescription = raw.LastErrorDescription
	c.Brightness = raw.Brightness
	c.Contrast = raw.Contrast
	c.AudioFormat = raw.AudioFormat
	c.CCMovie = raw.CCMovie
	c.CCImage = raw.CCImage
	c.MCMovie = raw.MCMovie
	c.MCImage = raw.MCImage
	c.MCTriggerVideo = raw.MCTriggerVideo
	c.MCTriggerAudio = raw.MCTriggerAudio
	c.ATriggerVideo = raw.ATriggerVideo
	c.ATriggerAudio = raw.ATriggerAudio
	c.CustomModel = raw.CustomModel
	c.SinceLastCapture = raw.SinceLastCapture

	return nil
}

func (x *cameraXML) StoragePathSet() bool {
	return x.CapturePathV6 != ""
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}

	return ""
}
