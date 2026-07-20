package securityspy

import (
	"encoding/xml"
	"strconv"
	"strings"
	"sync"
	"time"

	"golift.io/securityspy/v2/server"
)

// Server is the main interface for this library.
// Contains sub-interfaces for cameras, ptz, files & events
// This is provided in exchange for a url, username and password.
// If your app calls Refresh(), it is your duty to use Rlock() on
// this struct if there's a chance you may call methods while
// Refresh() is running.
type Server struct {
	*server.Config

	// Encoder was previously the path to an ffmpeg binary.
	//
	// Deprecated: unused; video capture is pure Go and does not shell out to ffmpeg.
	Encoder string
	Files   *Files       // Files interface.
	Events  *Events      // Events interface.
	Cameras *Cameras     // Cameras & PTZ interfaces.
	Groups  []*Group     // Camera groups from systemInfo (v6+).
	Info    *ServerInfo  // ServerInfo struct (no methods).
	mu      sync.RWMutex // Lock for Refresh().
}

// Group is a named camera group from ++systemInfo (v6+).
type Group struct {
	Number  int    `xml:"number"`
	Name    string `xml:"name"`
	Cameras string `xml:"cameras"` // comma-separated camera numbers
}

// CameraNumbers returns the camera numbers listed in the group.
func (g *Group) CameraNumbers() []int {
	if g == nil || g.Cameras == "" {
		return nil
	}

	parts := strings.Split(g.Cameras, ",")
	nums := make([]int, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		n, err := strconv.Atoi(part)
		if err != nil {
			continue
		}

		nums = append(nums, n)
	}

	return nums
}

// ServerInfo represents all the SecuritySpy server's information.
// This becomes available as server.Info.
type ServerInfo struct {
	Name             string    `xml:"name"`               // SecuritySpy
	Version          string    `xml:"version"`            // 6.20
	UUID             string    `xml:"uuid"`               // EXAMPLEUUID000000001
	EventStreamCount int64     `xml:"eventstreamcount"`   // legacy v5
	DDNSName         string    `xml:"ddns-name"`          // domain.name.dyn
	WanAddress       string    `xml:"wan-address"`        // wan.example.test
	ServerName       string    `xml:"server-name"`        // display name
	BonjourName      string    `xml:"bonjour-name"`       // securityspy.local
	IP1              string    `xml:"ip1"`                // 192.0.2.1
	IP2              string    `xml:"ip2"`                // 198.51.100.1
	HTTPEnabled      YesNoBool `xml:"http-enabled"`       // true/false or yes/no
	HTTPPort         int       `xml:"http-port"`          // 8000
	HTTPPortWan      int       `xml:"http-port-wan"`      // 8000
	HTTPSEnabled     YesNoBool `xml:"https-enabled"`      // true/false
	HTTPSPort        int       `xml:"https-port"`         // 8001
	HTTPSPortWan     int       `xml:"https-port-wan"`     // 8001
	CurrentTime      time.Time `xml:"current-local-time"` // 2019-02-10T03:08:12-08:00
	GmtOffset        Duration  `xml:"seconds-from-gmt"`   // -28800
	DateFormat       string    `xml:"date-format"`        // MM/DD/YYYY
	TimeFormat       string    `xml:"time-format"`        // 12, 24
	CPUUsage         int       `xml:"cpu-usage"`          // 37
	// v6+ server fields
	CameraCount         int       `xml:"camera-count"`
	MemoryPressure      int       `xml:"memory-pressure"`
	CertExpiryDays      int       `xml:"cert-expiry-days"`
	CertExpiryTime      time.Time `xml:"cert-expiry-time"`
	ArchiveStatus       string    `xml:"archive-status"`
	WanProxy            YesNoBool `xml:"wan-proxy"`
	WanProxyProtocol    int       `xml:"wan-proxy-protocol"`
	NewVersion          string    `xml:"new-version"`
	CurrentAbsoluteTime float64   `xml:"current-absolute-time"`
	// These are all copied in/created by Refresh()
	Refreshed         time.Time
	ServerSchedules   map[int]string
	SchedulePresets   map[int]string
	ScheduleOverrides map[int]string
}

// systemInfo represents ++systemInfo api path (v6 primary tags + v5 legacy lists).
type systemInfo struct {
	XMLName    xml.Name    `xml:"system"`
	Server     *ServerInfo `xml:"server"`
	CameraList struct {
		Cameras []*Camera `xml:"camera"`
	} `xml:"camera-list"`
	CameraListLegacy struct {
		Cameras []*Camera `xml:"camera"`
	} `xml:"cameralist"`
	GroupList struct {
		Groups []*Group `xml:"group"`
	} `xml:"group-list"`
	Schedules               ScheduleContainer `xml:"schedule-list"`
	SchedulesLegacy         ScheduleContainer `xml:"schedulelist"`
	SchedulePresets         ScheduleContainer `xml:"schedule-preset-list"`
	SchedulePresetsLegacy   ScheduleContainer `xml:"schedulepresetlist"`
	ScheduleOverrides       ScheduleContainer `xml:"schedule-override-list"`
	ScheduleOverridesLegacy ScheduleContainer `xml:"scheduleoverridelist"`
}

// cameras returns the v6 camera list, falling back to the legacy v5 list.
func (s *systemInfo) cameras() []*Camera {
	if len(s.CameraList.Cameras) > 0 {
		return s.CameraList.Cameras
	}

	return s.CameraListLegacy.Cameras
}

func (s *systemInfo) schedules() ScheduleContainer {
	if len(s.Schedules) > 0 {
		return s.Schedules
	}

	return s.SchedulesLegacy
}

func (s *systemInfo) schedulePresets() ScheduleContainer {
	if len(s.SchedulePresets) > 0 {
		return s.SchedulePresets
	}

	return s.SchedulePresetsLegacy
}

func (s *systemInfo) scheduleOverrides() ScheduleContainer {
	if len(s.ScheduleOverrides) > 0 {
		return s.ScheduleOverrides
	}

	return s.ScheduleOverridesLegacy
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
