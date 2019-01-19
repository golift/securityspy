package securityspy

import (
	"encoding/xml"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Error enables constant errors.
type Error string

// Error allows a string to satisfy the error type.
func (e Error) Error() string {
	return string(e)
}

// concourse is the main interface.
type concourse struct {
	SystemInfo *systemInfo
	StopChan   chan bool
	EventChan  chan Event
	Running    bool
	VerifySSL  bool
	BaseURL    string
	AuthB64    string
	Username   string
	EventBinds map[EventName][]func(Event)
	sync.RWMutex
}

// ServerInfo represents all the SecuritySpy ServerInfo Info
type ServerInfo struct {
	Name             string    `xml:"name"`             // SecuritySpy
	Version          string    `xml:"version"`          // 4.2.9
	UUID             string    `xml:"uuid"`             // C03L1333F8J3AkXIZS1O
	EventStreamCount int64     `xml:"eventstreamcount"` // 99270
	DDNSName         string    `xml:"ddns-name"`        // domain.name.dyn
	WanAddress       string    `xml:"wan-address"`      // domain.name
	ServerName       string    `xml:"server-name"`
	BonjourName      string    `xml:"bonjour-name"`
	IP1              string    `xml:"ip1"`            // 192.168.3.1
	IP2              string    `xml:"ip2"`            // 192.168.69.3
	HTTPEnabled      YesNoBool `xml:"http-enabled"`   // yes
	HTTPPort         int       `xml:"http-port"`      // 8000
	HTTPPortWan      int       `xml:"http-port-wan"`  // 8000
	HTTPSEnabled     YesNoBool `xml:"https-enabled"`  // no
	HTTPSPort        int       `xml:"https-port"`     // 8001
	HTTPSPortWan     int       `xml:"https-port-wan"` // 8001
	Refreshed        time.Time // updated by Refresh()
	// These are shoehorned in.
	Scripts struct {
		Names []string `xml:"name"`
	} `xml:"scripts"`
	Sounds struct {
		Names []string `xml:"name"`
	} `xml:"sounds"`
}

// systemInfo reresents ++systemInfo
type systemInfo struct {
	XMLName    xml.Name   `xml:"system"`
	Server     ServerInfo `xml:"server"`
	CameraList struct {
		Cameras []CameraDevice `xml:"camera"`
	} `xml:"cameralist"`
	ScheduleList struct {
		Schedules []schedule `xml:"schedule"`
	} `xml:"schedulelist"`
	SchedulePresetList struct {
		SchedulePresets []schedulePresets `xml:"schedulepreset"`
	} `xml:"schedulepresetlist"`
}

// YesNoBool is used to capture strings into boolean format.
type YesNoBool struct {
	Val bool
	Txt string
}

// UnmarshalXML method converts armed/disarmed, yes/no, active/inactive or 0/1 to true/false.
// Really it converts armed, yes, active, enabled, 1, true to true. Anything else is false.
func (bit *YesNoBool) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	_ = d.DecodeElement(&bit.Txt, &start)
	bit.Val = bit.Txt == "1" || strings.EqualFold(bit.Txt, "true") || strings.EqualFold(bit.Txt, "yes") ||
		strings.EqualFold(bit.Txt, "armed") || strings.EqualFold(bit.Txt, "active") || strings.EqualFold(bit.Txt, "enabled")
	return nil
}

// Duration is used to convert the "Seconnds" given to us by the securityspy API into a go time.Duration.
type Duration struct {
	Dur time.Duration
	Sec string
}

// UnmarshalXML method converts seconds to time.Duration.
func (bit *Duration) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	_ = d.DecodeElement(&bit.Sec, &start)
	r, _ := strconv.Atoi(bit.Sec)
	bit.Dur = time.Second * time.Duration(r)
	return nil
}

// Server is the interface to the Kingdom, ie concourse.
type Server interface {
	// SecuritySpy
	Info() ServerInfo      // This is the server data struct. Has everything about the server.
	Refresh() error        // call this once in a while if you build a daemon.
	RefreshScripts() error // probably useless
	RefreshSounds() error  // same. no documented methods to do anything with this.
	// Files (2 sub interfaces)
	Files() (files Files) // Interface into saved/captured vidoes and images.
	// Cameras (2 sub interfaces)
	GetCameras() (cams []Camera)
	GetCamera(cameraNum int) (cam Camera)
	GetCameraByName(name string) (cam Camera)
	// Events (no sub interfaces)
	WatchEvents(retryInterval time.Duration, refreshOnConfigChange bool)
	StopWatch()
	BindEvent(event EventName, callBack func(Event))
	UnbindEvent(event EventName)
	UnbindAllEvents()
	CustomEvent(cameraNum int, msg string)
}
