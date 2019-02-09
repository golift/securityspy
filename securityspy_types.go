package securityspy

import (
	"encoding/xml"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

var (
	ErrorCmdNotOK = errors.New("command unsuccessful")
	// DefaultTimeout it used for almost every request to SecuritySpy. Adjust as needed.
	DefaultTimeout = 10 * time.Second
)

// Server is the main interface for this library.
// Contains sub-interfaces for cameras, ptz, files & events
type Server struct {
	verifySSL  bool
	baseURL    string
	authB64    string
	username   string
	systemInfo *systemInfo
	Files      *Files
	Events     *Events
	Cameras    *Cameras
	Info       *ServerInfo
}

// ServerInfo represents all the SecuritySpy server's information.
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
	CurrentTime      time.Time `xml:"current-local-time"`
	GmtOffset        Duration  `xml:"seconds-from-gmt"`
	DateFormat       string    `xml:"date-format"`
	TimeFormat       string    `xml:"time-format"`
	// These are all copied in by Refresh()
	Refreshed         time.Time
	ScriptsNames      []string
	SoundsNames       []string
	ServerSchedules   map[int]string
	SchedulePresets   map[int]string
	ScheduleOverrides map[int]string
	// If there is a chance of calling Refresh() while reading these maps, lock them.
	sync.RWMutex
}

// systemInfo reresents ++systemInfo
type systemInfo struct {
	XMLName    xml.Name    `xml:"system"`
	Server     *ServerInfo `xml:"server"`
	CameraList struct {
		Cameras []*Camera `xml:"camera"`
	} `xml:"cameralist"`
	// All of these sub-lists get copied into ServerInfo by Refresh()
	Schedules         scheduleContainer `xml:"schedulelist"`
	SchedulePresets   scheduleContainer `xml:"schedulepresetlist"`
	ScheduleOverrides scheduleContainer `xml:"scheduleoverridelist"`
	// These are shoehorned in.
	Scripts struct {
		Names []string `xml:"name"`
	} `xml:"scripts"`
	Sounds struct {
		Names []string `xml:"name"`
	} `xml:"sounds"`
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
	time.Duration
	Val string
}

// UnmarshalXML method converts seconds to time.Duration.
func (bit *Duration) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	_ = d.DecodeElement(&bit.Val, &start)
	r, _ := strconv.Atoi(bit.Val)
	if bit.Duration = time.Second * time.Duration(r); bit.Val == "" {
		// In the context of this application -1ns will significantly make
		// obvious the fact that this value was empty and not a number.
		// This typically happens for a camera's last motion event ticker
		// when one has yet to happen [since securityspy started].
		bit.Duration = -1
	}
	return nil
}
