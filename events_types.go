package securityspy

import (
	"sync"
	"time"

	"github.com/pkg/errors"
)

var (
	// ErrorUnknownEvent never really returns, but will fire if SecuritySpy
	// adds new events this library doesn't know about.
	ErrorUnknownEvent = errors.New("unknown event")
	// ErrorCAMParseFail will return if the camera number in an event stream does not exist.
	// If you see this, run Refresh() more often, or fix your flaky camera connection.
	ErrorCAMParseFail = errors.New("CAM parse failed")
	// ErrorIDParseFail will return if the camera number provided by the event stream is not a number.
	// This should never happen, but future versions of SecuritySpy could trigger this if formats change.
	ErrorIDParseFail = errors.New("ID parse failed")
	// ErrorCAMMissing like the errors above should never return.
	// This is triggered by a corrupted event format.
	ErrorCAMMissing = errors.New("CAM missing")
	// ErrorDateParseFail will only trigger if the time stamp format for events changes.
	ErrorDateParseFail = errors.New("timestamp parse failed")
	// ErrorDisconnect becomes the msg in a custom event when the SecSpy event stream is disconnected.
	ErrorDisconnect = errors.New("event stream disconnected")
	// unknownEventText should only happen if SecuritySpy adds new event types.
	unknownEventText = "Unknown Event"
	// eventTimeFormat is the go-time-format returned by SecuritySpy's eventStream
	// The GMT offset from ++systemInfo is appended later for unmarshaling.
	eventTimeFormat = "20060102150405"
)

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

// Event Stream Reply is sent to callback channels and/or functions.
type Event struct {
	Time   time.Time // Local time.
	When   time.Time // Server time.
	ID     int       // Negative numbers are custom events.
	Camera *Camera   // Each event gets a camera interface.
	Type   EventType // Event identifier
	Msg    string    // Event Text
	Errors []error   // Errors populated by parse errors.
}

// EventType is a set of constant strings validated with Event() method
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
