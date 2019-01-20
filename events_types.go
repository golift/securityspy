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
)

// eventTimeFormat is the go-time-format returned by SecuritySpy's eventStream
var eventTimeFormat = "20060102150405"

type events struct {
	stopChan  chan bool
	eventChan chan Event
	Running   bool
	*Server
	sync.RWMutex // lock for both maps.
}

// Events is the interface into the event stream.
type Events interface {
	Watch(retryInterval time.Duration, refreshOnConfigChange bool)
	Stop()
	BindFunc(event EventName, callBack func(Event))
	BindChan(event EventName, channel chan Event)
	UnbindFunc(event EventName)
	UnbindChan(event EventName)
	UnbindAll()
	Custom(cameraNum int, msg string)
}

// Event Stream Reply
type Event struct {
	When   time.Time
	ID     int
	Camera Camera
	Event  EventName
	Msg    string
	Errors []error
}

// EventName is a set of constants validated with Event() method
type EventName string

// Events
const (
	EventArmContinuous    EventName = "ARM_C"
	EventDisarmContinuous EventName = "DISARM_C"
	EventArmMotion        EventName = "ARM_M"
	EventDisarmMotion     EventName = "DISARM_M"
	EventDisarmActions    EventName = "DISARM_A"
	EventArmActions       EventName = "ARM_A"
	EventSecSpyError      EventName = "ERROR"
	EventConfigChange     EventName = "CONFIGCHANGE"
	EventMotionDetected   EventName = "MOTION"
	EventOnline           EventName = "ONLINE"
	EventOffline          EventName = "OFFLINE"
	// The following belong to the library, not securityspy.
	EventStreamDisconnect   EventName = "DISCONNECTED"
	EventStreamConnect      EventName = "CONNECTED"
	EventUnknownEvent       EventName = "UNKNOWN"
	EventAllEvents          EventName = "ALL"
	EventWatcherRefreshed   EventName = "REFRESH"
	EventWatcherRefreshFail EventName = "REFRESHFAIL"
	EventStreamCustom       EventName = "CUSTOM"
)

// Event provides a description of an event.
func (e EventName) Event() string {
	switch e {
	case EventArmContinuous:
		return "Continuous Capture Armed"
	case EventDisarmContinuous:
		return "Continuous Capture Disarmed"
	case EventArmMotion:
		return "Motion Capture Armed"
	case EventDisarmMotion:
		return "Motion Capture Disarmed"
	case EventArmActions:
		return "Actions Armed"
	case EventDisarmActions:
		return "Actions Disarmed"
	case EventSecSpyError:
		return "SecuritySpy Error"
	case EventConfigChange:
		return "Configuration Change"
	case EventMotionDetected:
		return "Motion Detection"
	case EventOffline:
		return "Camera Offline"
	case EventOnline:
		return "Camera Online"
		// The following belong to the library, not securityspy.
	case EventStreamDisconnect:
		return "Event Stream Disconnected"
	case EventStreamConnect:
		return "Event Stream Connected"
	case EventUnknownEvent:
		return "Unknown Event"
	case EventAllEvents:
		return "Any Event"
	case EventWatcherRefreshed:
		return "SystemInfo Refresh Success"
	case EventWatcherRefreshFail:
		return "SystemInfo Refresh Failure"
	case EventStreamCustom:
		return "Custom Event"

	}
	return EventUnknownEvent.Event()
}

// String provides the string form of an Event.
func (e EventName) String() string {
	return string(e)
}
