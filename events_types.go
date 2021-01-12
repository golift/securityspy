package securityspy

import (
	"fmt"
	"io"
	"sync"
	"time"
)

// This is a list of errors returned by the Events methods.
var (
	// ErrorUnknownEvent never really returns, but will fire if SecuritySpy
	// adds new events this library doesn't know about.
	ErrorUnknownEvent = fmt.Errorf("unknown event")
	// ErrorCAMParseFail will return if the camera number in an event stream does not exist.
	// If you see this, run Refresh() more often, or fix your flaky camera connection.
	ErrorCAMParseFail = fmt.Errorf("CAM parse failed")
	// ErrorIDParseFail will return if the camera number provided by the event stream is not a number.
	// This should never happen, but future versions of SecuritySpy could trigger this if formats change.
	ErrorIDParseFail = fmt.Errorf("ID parse failed")
	// ErrorCAMMissing like the errors above should never return.
	// This is triggered by a corrupted event format.
	ErrorCAMMissing = fmt.Errorf("camera number missing")
	// ErrorDateParseFail will only trigger if the time stamp format for events changes.
	ErrorDateParseFail = fmt.Errorf("timestamp parse failed")
	// ErrorDisconnect becomes the msg in a custom event when the SecSpy event stream is disconnected.
	ErrorDisconnect = fmt.Errorf("server connection closed")
)

const (
	// BadID happens when the ID cannot be parsed.
	BadID = -2
	// EventBuffer is the channel buffer size for securityspy events.
	EventBuffer = 10000
	// UnknownEventText should only appear if SecuritySpy adds new event types.
	UnknownEventText = "Unknown Event"
	// UnknownReasonText should only appear if SecuritySpy adds new motion detection reasons.
	UnknownReasonText = "Unknown Reason"
	// EventTimeFormat is the go-time-format returned by SecuritySpy's eventStream
	// The GMT offset from ++systemInfo is appended later for unmarshaling w/ localization.
	EventTimeFormat = "20060102150405"
)

// Events is the main Events interface. Use the methods bound here to bind your
// own functions, methods and channels to SecuritySpy events. Call Watch() to
// connect to the event stream. The Running bool is true when the event stream
// watcher routine is active.
type Events struct {
	server     *Server
	stream     io.ReadCloser
	eventChan  chan Event
	eventBinds map[EventType][]func(event Event)
	eventChans map[EventType][]chan Event
	binds      sync.RWMutex
	chans      sync.RWMutex
	Running    bool
}

// Event represents a SecuritySpy event from the Stream Reply.
// This is the INPUT data for an event that is sent to a bound callback method or channel.
type Event struct {
	Time   time.Time // Local time event was recorded.
	When   time.Time // Event time according to server.
	ID     int       // Negative numbers are custom events.
	Camera *Camera   // Each event gets a camera interface.
	Type   EventType // Event identifier
	Msg    string    // Event Text
	Errors []error   // Errors populated by parse errors.
}

// EventType is a set of constant strings validated by the EventNames map.
type EventType string

// Events that can be returned by the event stream.
// These events can have channels or callback functions bound to them.
// The DISCONNECTED event fires when the event stream is disconnected, so
// watch that event with a binding to detect stream interruptions.
const (
	EventArmContinuous    EventType = "ARM_C"
	EventDisarmContinuous EventType = "DISARM_C"
	EventArmMotion        EventType = "ARM_M"
	EventDisarmMotion     EventType = "DISARM_M"
	EventDisarmActions    EventType = "DISARM_A"
	EventArmActions       EventType = "ARM_A"
	EventSecSpyError      EventType = "ERROR"
	EventConfigChange     EventType = "CONFIGCHANGE"
	EventMotionDetected   EventType = "MOTION" // Legacy (v4)
	EventOnline           EventType = "ONLINE"
	EventOffline          EventType = "OFFLINE"
	EventClassify         EventType = "CLASSIFY"
	EventTriggerMotion    EventType = "TRIGGER_M"
	EventTriggerAction    EventType = "TRIGGER_A"
	EventFileWritten      EventType = "FILE"
	EventKeepAlive        EventType = "NULL"
	// The following belong to the library, not securityspy.
	EventStreamDisconnect   EventType = "DISCONNECTED"
	EventStreamConnect      EventType = "CONNECTED"
	EventUnknownEvent       EventType = "UNKNOWN"
	EventAllEvents          EventType = "ALL"
	EventWatcherRefreshed   EventType = "REFRESH"
	EventWatcherRefreshFail EventType = "REFRESHFAIL"
	EventStreamCustom       EventType = "CUSTOM"
	eventStreamStop         EventType = "STOP"
)

// EventName returns the human readable names for each event.
func EventName(e EventType) string {
	return map[EventType]string{
		EventArmContinuous:    "Continuous Capture Armed",
		EventDisarmContinuous: "Continuous Capture Disarmed",
		EventArmMotion:        "Motion Capture Armed",
		EventDisarmMotion:     "Motion Capture Disarmed",
		EventArmActions:       "Actions Armed",
		EventDisarmActions:    "Actions Disarmed",
		EventSecSpyError:      "SecuritySpy Error",
		EventConfigChange:     "Configuration Change",
		EventMotionDetected:   "Motion Detected", // Legacy (v4)
		EventOffline:          "Camera Offline",
		EventOnline:           "Camera Online",
		EventClassify:         "Classification",
		EventTriggerMotion:    "Triggered Motion",
		EventTriggerAction:    "Triggered Action",
		EventFileWritten:      "File Written",
		EventKeepAlive:        "Stream Keep Alive",
		// The following belong to the library, not securityspy.
		EventStreamDisconnect:   "Event Stream Disconnected",
		EventStreamConnect:      "Event Stream Connected",
		EventUnknownEvent:       UnknownEventText,
		EventAllEvents:          "Any Event",
		EventWatcherRefreshed:   "SystemInfo Refresh Success",
		EventWatcherRefreshFail: "SystemInfo Refresh Failure",
		EventStreamCustom:       "Custom Event",
	}[e]
}

// TriggerEvent represent the "Reason" a motion or action trigger occurred. v5+ only.
type TriggerEvent int

// These are the trigger reasons SecuritySpy exposes. v5+ only.
const (
	TriggerByMotion = TriggerEvent(1) << iota
	TriggerByAudio
	TriggerByScript
	TriggerByCameraEvent
	TriggerByWebServer
	TriggerByOtherCamera
	TriggerByManual
	TriggerByHumanDetection
	TriggerByVehicleDetection
)

// Reasons is the human-readable explanation for a motion detection reason.
var Reasons = map[TriggerEvent]string{ //nolint:gochecknoglobals
	TriggerByMotion:           "Motion Detected",
	TriggerByAudio:            "Audio Detected",
	TriggerByScript:           "AppleScript",
	TriggerByCameraEvent:      "Camera Event",
	TriggerByWebServer:        "Web Server",
	TriggerByOtherCamera:      "Other Camera",
	TriggerByManual:           "Manual",
	TriggerByHumanDetection:   "Human Detected",
	TriggerByVehicleDetection: "Vehicle Detected",
}
