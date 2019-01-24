package securityspy

import (
	"sync"
	"time"
)

// Error strings used in this file.
const (
	ErrorUnknownEvent  = Error("unknown event")
	ErrorCAMParseFail  = Error("CAM parse failed")
	ErrorIDParseFail   = Error("ID parse failed")
	ErrorCAMMissing    = Error("CAM missing")
	ErrorDateParseFail = Error("timestamp parse failed")
	ErrorUnknownError  = Error("unknown error")
	ErrorDisconnect    = Error("event stream disconnected")
	unknownEventText   = "Unknown Event"
)

// eventTimeFormat is the go-time-format returned by SecuritySpy's eventStream
var eventTimeFormat = "20060102150405"

// Events is the main Events interface.
type Events struct {
	server     *Server
	stopChan   chan bool
	eventChan  chan Event
	eventBinds map[EventType][]func(Event)
	eventChans map[EventType][]chan Event
	binds      sync.RWMutex
	chans      sync.RWMutex
	Running    bool
}

// Event Stream Reply
type Event struct {
	When   time.Time
	ID     int
	Camera *Camera
	Type   EventType
	Msg    string
	Errors []error
}

// EventType is a set of constants validated with Event() method
type EventType string

// Events
const (
	EventArmContinuous    EventType = "ARM_C"
	EventDisarmContinuous EventType = "DISARM_C"
	EventArmMotion        EventType = "ARM_M"
	EventDisarmMotion     EventType = "DISARM_M"
	EventDisarmActions    EventType = "DISARM_A"
	EventArmActions       EventType = "ARM_A"
	EventSecSpyError      EventType = "ERROR"
	EventConfigChange     EventType = "CONFIGCHANGE"
	EventMotionDetected   EventType = "MOTION"
	EventOnline           EventType = "ONLINE"
	EventOffline          EventType = "OFFLINE"
	// The following belong to the library, not securityspy.
	EventStreamDisconnect   EventType = "DISCONNECTED"
	EventStreamConnect      EventType = "CONNECTED"
	EventUnknownEvent       EventType = "UNKNOWN"
	EventAllEvents          EventType = "ALL"
	EventWatcherRefreshed   EventType = "REFRESH"
	EventWatcherRefreshFail EventType = "REFRESHFAIL"
	EventStreamCustom       EventType = "CUSTOM"
)
