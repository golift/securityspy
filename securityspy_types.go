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
	Config     *config
	SystemInfo *systemInfo
	EventBinds map[EventName][]func(Event)
	StopChan   chan bool
	Running    bool
	sync.RWMutex
}

// config is the data passed into the Handler function.
type config struct {
	VerifySSL bool
	BaseURL   string
	AuthB64   string
	Username  string
}

// VidOps are the options for a video that can be requested from SecuritySpy
type VidOps struct {
	Width   int
	Height  int
	FPS     int
	Quality int
}

// Server is the interface to the Kingdom.
type Server interface {
	Refresh() error
	RefreshScripts() error
	RefreshSounds() error
	Info() ServerInfo
	GetCameras() (cams []Camera)
	GetCamera(camerNum int) (cam Camera)
	GetCameraByName(name string) (cam Camera)
	BindEvent(event EventName, callBack func(Event))
	UnbindEvent(event EventName)
	UnbindAllEvents()
	WatchEvents(retryInterval, refreshInterval time.Duration)
	StopWatch()
	Files() (files Files) // this still sucks.
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
	// These are shoehorned in.
	Scripts struct {
		Names []string `xml:"name"`
	} `xml:"scripts"`
	Sounds struct {
		Names []string `xml:"name"`
	} `xml:"sounds"`
}

// systemInfo reresents /++systemInfo
type systemInfo struct {
	XMLName    xml.Name   `xml:"system"`
	Server     ServerInfo `xml:"server"`
	CameraList struct {
		Cameras []CameraDevice `xml:"camera"`
	} `xml:"cameralist"`
	ScheduleList struct {
		Schedules []Schedule `xml:"schedule"`
	} `xml:"schedulelist"`
	SchedulePresetList struct {
		SchedulePresets []SchedulePresets `xml:"schedulepreset"`
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
	yes := bit.Txt == "1" || strings.EqualFold(bit.Txt, "true") || strings.EqualFold(bit.Txt, "yes") ||
		strings.EqualFold(bit.Txt, "armed") || strings.EqualFold(bit.Txt, "active") || strings.EqualFold(bit.Txt, "enabled")
	bit.Val = yes
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
