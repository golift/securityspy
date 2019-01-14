package securityspy

import (
	"encoding/xml"
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
	Config     Config
	SystemInfo *SystemInfo
	EventBinds map[EventName][]func(Event)
	StopChan   chan bool
	Running    bool
	*sync.RWMutex
}

// Config is the data passed into the Handler function.
type Config struct {
	VerifySSL bool
	BaseURL   string
	AuthB64   string
	Username  string
}

// SecuritySpy is the interface to the Kingdom.
type SecuritySpy interface {
	Refresh() error
	ServerInfo() Server
	Cameras() (cams []Camera)
	Camera(int) (cam Camera)
	CameraByName(name string) (cam Camera)
	Scripts() (scripts []string, err error)
	Sounds() (sounds []string, err error)
	BindEvent(event EventName, callBack func(Event))
	UnbindEvent(event EventName)
	UnbindAllEvents()
	WatchEvents(retryInterval time.Duration)
	StopWatch()
	Files() (files Files)
}

// Server represents all the SecuritySpy Server Info
type Server struct {
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
}

// SystemInfo reresents /++systemInfo
type SystemInfo struct {
	XMLName         xml.Name `xml:"system"`
	Server          Server   `xml:"server"`
	CameraContainer struct {
		Cameras []CameraInterface `xml:"camera"`
	} `xml:"cameralist"`
	Schedulelist struct {
		Schedules []Schedule `xml:"schedule"`
	} `xml:"schedulelist"`
	Schedulepresetlist struct {
		SchedulePresets []SchedulePresets `xml:"schedulepreset"`
	} `xml:"schedulepresetlist"`
}

// YesNoBool is used to capture strings into boolean format.
type YesNoBool bool

// UnmarshalJSON method converts armed/disarmed, yes/no or 0/1 to true/false.
func (bit *YesNoBool) UnmarshalJSON(data []byte) error {
	s := string(data)
	*bit = YesNoBool(s == "1" || strings.EqualFold(s, "true") || strings.EqualFold(s, "yes") ||
		strings.EqualFold(s, "armed") || strings.EqualFold(s, "active"))
	return nil
}
