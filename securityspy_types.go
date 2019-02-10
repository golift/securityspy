package securityspy

import (
	"encoding/xml"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

// ErrorCmdNotOK is returned for any command that has a successful web request,
// but the reply does not end with the word OK.
var ErrorCmdNotOK = errors.New("command unsuccessful")

// DefaultTimeout it used for almost every request to SecuritySpy. Adjust as needed.
var DefaultTimeout = 10 * time.Second

// Server is the main interface for this library.
// Contains sub-interfaces for cameras, ptz, files & events
// This is provided in exchange for a url, username and password.
type Server struct {
	// override the local methods with the interface methods.
	api
	verifySSL  bool
	baseURL    string
	authB64    string
	username   string
	systemInfo *systemInfo
	Files      *Files      // Files interface.
	Events     *Events     // Events interface.
	Cameras    *Cameras    // Cameras & PTZ interfaces.
	Info       *ServerInfo // ServerInfo struct (no methods).
}

// ServerInfo represents all the SecuritySpy server's information.
// This becomes available as server.Info.
type ServerInfo struct {
	Name             string    `xml:"name"`               // SecuritySpy
	Version          string    `xml:"version"`            // 4.2.10
	UUID             string    `xml:"uuid"`               // C03L1333F8J3AkXIZS1O
	EventStreamCount int64     `xml:"eventstreamcount"`   // 99270
	DDNSName         string    `xml:"ddns-name"`          // domain.name.dyn
	WanAddress       string    `xml:"wan-address"`        // domain.name
	ServerName       string    `xml:"server-name"`        // <usually empty>
	BonjourName      string    `xml:"bonjour-name"`       // <usually empty>
	IP1              string    `xml:"ip1"`                // 192.168.3.1
	IP2              string    `xml:"ip2"`                // 192.168.69.3
	HTTPEnabled      YesNoBool `xml:"http-enabled"`       // yes
	HTTPPort         int       `xml:"http-port"`          // 8000
	HTTPPortWan      int       `xml:"http-port-wan"`      // 8000
	HTTPSEnabled     YesNoBool `xml:"https-enabled"`      // no
	HTTPSPort        int       `xml:"https-port"`         // 8001
	HTTPSPortWan     int       `xml:"https-port-wan"`     // 8001
	CurrentTime      time.Time `xml:"current-local-time"` // 2019-02-10T03:08:12-08:00
	GmtOffset        Duration  `xml:"seconds-from-gmt"`   // -28800
	DateFormat       string    `xml:"date-format"`        // MM/DD/YYYY
	TimeFormat       string    `xml:"time-format"`        // 12, 24
	// These are all copied in/created by Refresh()
	Refreshed         time.Time
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
}

// api interface is provided only to allow overriding local methods during local testing.
// The methods in this interface connect to SecuritySpy so they become
// blockers when testing without a SecuritySpy server available. Overriding
// them with fakes makes testing (for most methods in this library) possible.
type api interface {
	secReq(apiPath string, params url.Values, timeout time.Duration) (resp *http.Response, err error)
	secReqXML(apiPath string, params url.Values) (body []byte, err error)
	simpleReq(apiURI string, params url.Values, cameraNum int) error
}

// YesNoBool is used to capture strings into boolean format. If the string has
// a Val of: 1, true, yes, armed, active, or enabled then the boolean is true.
// Any other string Val and the boolean is false.
type YesNoBool struct {
	Val bool
	Txt string
}

// UnmarshalXML method converts armed/disarmed, yes/no, active/inactive or 0/1 to true/false.
// Really it converts armed, yes, active, enabled, 1, true to true. Anything else is false.
// This isn't a method you should ever call directly; it is only used during data initialization.
func (bit *YesNoBool) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	_ = d.DecodeElement(&bit.Txt, &start)
	bit.Val = bit.Txt == "1" || strings.EqualFold(bit.Txt, "true") || strings.EqualFold(bit.Txt, "yes") ||
		strings.EqualFold(bit.Txt, "armed") || strings.EqualFold(bit.Txt, "active") || strings.EqualFold(bit.Txt, "enabled")
	return nil
}

// Duration is used to convert the "Seconds" given to us by the SecuritySpy API into a go time.Duration.
type Duration struct {
	time.Duration
	Val string
}

// UnmarshalXML method converts seconds from a string to time.Duration.
// This isn't a method you should ever call directly; it is only used during data initialization.
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
