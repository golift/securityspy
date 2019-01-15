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
	Name   string
	When   time.Time
	ID     int
	Camera int
	Event  EventName
	Raw    string
	Errors []error
}

// EventName is a set of constants validated with a read-only map.
type EventName string

// Events
const (
	EventArmContinuous    EventName = "ARM_C"
	EventDisarmContinuous           = "DISARM_C"
	EventArmMotion                  = "ARM_M"
	EventDisarmMotion               = "DISARM_M"
	EventDisarmActions              = "DISARM_A"
	EventArmActions                 = "ARM_A"
	EventSecSpyError                = "ERROR"
	EventConfigChange               = "CONFIGCHANGE"
	EventMotionDetected             = "MOTION"
	EventOnline                     = "ONLINE"
	EventOffline                    = "OFFLINE"
	EventStreamDisconnect           = "DISCONNECTED"
	EventUnknownEvent               = "UNKNOWN"
	EventAllEvents                  = "ALL"
)

// events is a read-only mapping of type to human text. Read As: When <event name> Occurs
var events = map[EventName]string{
	EventArmContinuous:    "Continuous Capture Armed",
	EventDisarmContinuous: "Continuous Capture Disarmed",
	EventArmMotion:        "Motion Armed",
	EventDisarmMotion:     "Motion Disarmed",
	EventArmActions:       "Actions Armed",
	EventDisarmActions:    "Actions Disarmed",
	EventSecSpyError:      "SecuritySpy Error",
	EventConfigChange:     "Configuration Change",
	EventMotionDetected:   "Motion Detection",
	EventOffline:          "Camera Goes Offline",
	EventOnline:           "Camera Comes Online",
	EventStreamDisconnect: "Event Stream Disconnection",
	EventUnknownEvent:     "Unknown Event",
	EventAllEvents:        "Any Event",
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
func (c *concourse) WatchEvents(retryInterval time.Duration) {
	c.Running = true
	eventChan := make(chan Event)
	reconnect := func() (io.ReadCloser, *bufio.Scanner) {
		if !c.Running {
			return nil, nil
		}
		resp, err := c.secReq("/++eventStream", nil, 0)
		for err != nil {
			eventChan <- Event{Event: EventStreamDisconnect, When: time.Now(), Camera: -1, ID: -1, Raw: err.Error()}
			time.Sleep(retryInterval)
			resp, err = c.secReq("/++eventStream", nil, 0)
		}
		return resp.Body, bufio.NewScanner(resp.Body)
	}

	go func() {
		for {
			select {
			case <-c.StopChan:
				c.Running = false
				return
			case event := <-eventChan:
				c.RLock()
				event.CallBacks(c.EventBinds)
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
		if scanner != nil && scanner.Scan() {
			eventChan <- ParseEvent(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			eventChan <- Event{Event: EventStreamDisconnect, When: time.Now(), Camera: -1, ID: -1, Raw: err.Error()}
			_ = body.Close()
			body, scanner = reconnect()
			scanner.Split(scanLinesCR)
		}
	}
}

// ParseEvent turns raw text into an Event that can fire callbacks.
func ParseEvent(text string) Event {
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
	e := Event{Raw: text, Event: EventUnknownEvent, Camera: -1, ID: -1, Errors: nil}
	parts := strings.SplitN(e.Raw, " ", 4)
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
	} else if e.Camera, err = strconv.Atoi(parts[2][3:]); err != nil {
		e.Camera = -1
		e.Errors = append(e.Errors, ErrorCAMParseFail)
	}
	// Parse the Event Type.
	event := EventName(strings.Split(parts[3], " ")[0])
	if name, ok := events[event]; ok {
		e.Event = event
		e.Name = name
	} else {
		e.Errors = append(e.Errors, ErrorUnknownEvent)
	}
	return e
}

// CallBacks is run for each event to execute callback functions.
func (e *Event) CallBacks(binds map[EventName][]func(Event)) {
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
