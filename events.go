package securityspy

import (
	"bufio"
	"bytes"
	"io"
	"strconv"
	"strings"
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
)

// Event Stream Reply
type Event struct {
	When   time.Time
	ID     int
	Camera Camera
	Event  EventName
	Msg    string
	Errors []error
}

// EventName is a set of constants validated with a read-only map.
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
	// These belong to the library, not securityspy.
	EventStreamDisconnect   EventName = "DISCONNECTED"
	EventUnknownEvent       EventName = "UNKNOWN"
	EventAllEvents          EventName = "ALL"
	EventWatcherRefreshed   EventName = "REFRESH"
	EventWatcherRefreshFail EventName = "REFRESHFAIL"
)

// events is a read-only mapping of type to human text. Read As: When <event name> Occurs
var events = map[EventName]string{
	EventArmContinuous:    "Continuous Capture Armed",
	EventDisarmContinuous: "Continuous Capture Disarmed",
	EventArmMotion:        "Motion Capture Armed",
	EventDisarmMotion:     "Motion Capture Disarmed",
	EventArmActions:       "Actions Armed",
	EventDisarmActions:    "Actions Disarmed",
	EventSecSpyError:      "SecuritySpy Error",
	EventConfigChange:     "Configuration Change",
	EventMotionDetected:   "Motion Detection",
	EventOffline:          "Camera Offline",
	EventOnline:           "Camera Online",
	// These belong to the library, not securityspy.
	EventStreamDisconnect:   "Event Stream Disconnected",
	EventUnknownEvent:       "Unknown Event",
	EventAllEvents:          "Any Event",
	EventWatcherRefreshed:   "SystemInfo Refresh Success",
	EventWatcherRefreshFail: "SystemInfo Refresh Failure",
}

// BindEvent binds a call-back function to an Event in SecuritySpy.
func (c *concourse) BindEvent(event EventName, callBack func(Event)) {
	if callBack == nil {
		return
	}
	c.Lock()
	defer c.Unlock()
	if val, ok := c.EventBinds[event]; ok {
		c.EventBinds[event] = append(val, callBack)
		return
	}
	c.EventBinds[event] = []func(Event){callBack}
}

// StopWatch stops WatchEvents loop
func (c *concourse) StopWatch() {
	if c.Running {
		c.StopChan <- false
	}
}

// UnbindAllEvents removes all event bindings.
func (c *concourse) UnbindAllEvents() {
	c.Lock()
	defer c.Unlock()
	c.EventBinds = make(map[EventName][]func(Event))
}

// UnbindEvent removes all bound callbacks for a particular event.
func (c *concourse) UnbindEvent(event EventName) {
	c.Lock()
	defer c.Unlock()
	delete(c.EventBinds, event)
}

// WatchEvents connects to securityspy and watches for events.
func (c *concourse) WatchEvents(retryInterval, refreshInterval time.Duration) {
	c.Running = true
	eventChan := make(chan Event, 1)
	reconnect := func() (io.ReadCloser, *bufio.Scanner) {
		if !c.Running {
			return nil, nil
		}
		resp, err := c.secReq("/++eventStream", nil, 0)
		for err != nil {
			raw := time.Now().Format("20060102150405") + " -99 CAM " + EventStreamDisconnect.String() + ": " + err.Error()
			eventChan <- c.parseEvent(raw)
			time.Sleep(retryInterval)
			resp, err = c.secReq("/++eventStream", nil, 0)
		}
		return resp.Body, bufio.NewScanner(resp.Body)
	}

	ticker := time.NewTicker(refreshInterval)
	// Watch for new events, a stop signal, or a refresh interval.
	go func() {
		for {
			select {
			case <-c.StopChan:
				c.Running = false
				return
			case <-ticker.C:
				if refreshInterval == 0 {
					break
				}
				raw := time.Now().Format("20060102150405") + " -98 CAM " + EventWatcherRefreshed.String() + " every " + refreshInterval.String()
				if err := c.Refresh(); err != nil {
					raw = time.Now().Format("20060102150405") + " -97 CAM " + EventWatcherRefreshFail.String() + ": " + err.Error()
				}
				eventChan <- c.parseEvent(raw)
			case event := <-eventChan:
				c.RLock()
				event.callBacks(c.EventBinds)
				c.RUnlock()
			}
		}
	}()

	body, scanner := reconnect()
	scanner.Split(scanLinesCR)
	defer func() {
		_ = body.Close()
	}()

	for {
		if !c.Running {
			return
		}
		// Constantly scan for new events, then report them to the event channel.
		if scanner != nil && scanner.Scan() {
			eventChan <- c.parseEvent(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			raw := time.Now().Format("20060102150405") + " -99 CAM " + EventStreamDisconnect.String() + ": " + err.Error()
			eventChan <- c.parseEvent(raw)
			_ = body.Close()
			body, scanner = reconnect()
			scanner.Split(scanLinesCR)
		}
	}
}

// parseEvent turns raw text into an Event that can fire callbacks.
func (c *concourse) parseEvent(text string) Event {
	/* [TIME] is specified in the order year, month, day, hour, minute, second and is always 14 characters long
		 * [EVENT NUMBER] increases by 1 for each subsequent event
		 * [CAMERA NUMBER] specifies the camera that this event relates to, for example CAM15 for camera number 15
		 * [EVENT] describes the event: ARM_C, DISARM_C, ARM_M, DISARM_M, ARM_A, DISARM_A, ERROR, CONFIGCHANGE, MOTION, OFFLINE, ONLINE
	     Example Event Stream Flow:
	     20140927091955 10 CAM0 ARM_C
	     20140927091955 11 CAM15 ARM_M
	     20140927092026 12 CAM0 MOTION
	     20140927091955 13 CAM0 DISARM_M
	     20140927092031 14 CAM17 OFFLINE
	     20190113141131 100525 CAM0 MOTION
	*/
	var err error
	parts := strings.SplitN(text, " ", 4)
	e := Event{Msg: parts[3], Camera: nil, ID: -1, Errors: nil}
	// Parse the time stamp
	zone, _ := time.Now().Zone() // SecuritySpy does not (really) provide this. :(
	if e.When, err = time.Parse("20060102150405MST", parts[0]+zone); err != nil {
		e.When = time.Now()
		e.Errors = append(e.Errors, ErrorDateParseFail)
	}
	// Parse the ID
	if e.ID, err = strconv.Atoi(parts[1]); err != nil {
		e.ID = -1
		e.Errors = append(e.Errors, ErrorIDParseFail)
	}
	// Parse the camera number.
	if !strings.HasPrefix(parts[2], "CAM") || len(parts[2]) < 4 {
		e.Errors = append(e.Errors, ErrorCAMMissing)
	} else if camNum, err := strconv.Atoi(parts[2][3:]); err != nil {
		e.Camera = nil
		e.Errors = append(e.Errors, ErrorCAMParseFail)
	} else if e.Camera = c.GetCamera(camNum); e.Camera == nil {
		e.Errors = append(e.Errors, ErrorCAMParseFail)
	}
	// Parse the Event Type.
	e.Event = EventName(strings.Split(parts[3], " ")[0])
	if _, ok := events[e.Event]; !ok {
		e.Errors = append(e.Errors, ErrorUnknownEvent)
		e.Event = EventUnknownEvent
	}
	return e
}

// callBacks is run for each event to execute callback functions.
func (e *Event) callBacks(binds map[EventName][]func(Event)) {
	callbacks := func(callbacks []func(Event)) {
		for _, callBack := range callbacks {
			if callBack != nil {
				go callBack(*e) // Send it off!
			}
		}
	}
	if _, ok := binds[e.Event]; ok {
		callbacks(binds[e.Event])
	} else if _, ok := binds[EventUnknownEvent]; ok && e.Event != EventUnknownEvent {
		callbacks(binds[EventUnknownEvent])
	}
	if _, ok := binds[EventAllEvents]; ok {
		callbacks(binds[EventAllEvents])
	}
}

// String turns an event name into a string.
func (e EventName) String() string {
	return string(e)
}

// Event returns the event value for the event.
func (e EventName) Event() string {
	if o, ok := events[e]; ok {
		return o
	}
	return e.String()
}

// scanLinesCR is a custom bufio.Scanner to read SecuritySpy eventStream.
func scanLinesCR(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\r'); i >= 0 {
		// We have a full CR-terminated line.
		return i + 1, data[0:i], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}
