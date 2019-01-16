package securityspy

import (
	"bufio"
	"bytes"
	"io"
	"strconv"
	"strings"
	"time"
)

/* Events-specific concourse methods are at the top. */

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
		c.Running = false
		c.StopChan <- c.Running
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

// WatchEvents kicks off he routines to watch the eventStream and fire callback bindings.
func (c *concourse) WatchEvents(retryInterval, refreshInterval time.Duration) {
	c.Running = true
	c.EventChan = make(chan Event, 1)
	go c.eventChannelSelector(refreshInterval)
	c.eventStreamScanner(retryInterval)
}

// NewEvent fires an event into the running event Watcher.
func (c *concourse) NewEvent(cameraNum int, msg string) {
	if !c.Running {
		return
	}
	c.EventChan <- c.parseEvent(time.Now().Format(eventTimeFormat) +
		" -11000 CAM" + strconv.Itoa(cameraNum) + " " +
		EventStreamCustom.String() + ": " + msg)
}

/* INTERFACE HELPER METHODS FOLLOW */

// eventStreamScanner connects to the securityspy event stream and fires events into a channel.
func (c *concourse) eventStreamScanner(retryInterval time.Duration) {
	body, scanner := c.eventStreamConnect(retryInterval)
	scanner.Split(scanLinesCR)
	for {
		if !c.Running {
			_ = body.Close()
			return // we all done here. stop got called
		}
		// Constantly scan for new events, then report them to the event channel.
		if scanner != nil && scanner.Scan() {
			c.EventChan <- c.parseEvent(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			raw := time.Now().Format(eventTimeFormat) + " -10000 CAM " + EventStreamDisconnect.String() + ": " + err.Error()
			c.EventChan <- c.parseEvent(raw)
			_ = body.Close()
			body, scanner = c.eventStreamConnect(retryInterval)
			scanner.Split(scanLinesCR)
		}
	}
}

// eventStreamConnect establishes a connection to the event stream and passes off the http Reader.
func (c *concourse) eventStreamConnect(retryInterval time.Duration) (io.ReadCloser, *bufio.Scanner) {
	resp, err := c.secReq("++eventStream", nil, 0)
	for err != nil {
		raw := time.Now().Format(eventTimeFormat) + " -9999 CAM " + EventStreamDisconnect.String() + ": " + err.Error()
		c.EventChan <- c.parseEvent(raw)
		time.Sleep(retryInterval)
		if !c.Running {
			return nil, nil
		}
		resp, err = c.secReq("++eventStream", nil, 0)
	}
	return resp.Body, bufio.NewScanner(resp.Body)
}

// eventChannelSelector watches a few internal channels for events and updates.
// Fires bound event call back functions.
func (c *concourse) eventChannelSelector(refreshInterval time.Duration) {
	ticker := time.NewTicker(refreshInterval)
	for {
		// Watch for new events, a stop signal, or a refresh interval.
		select {
		case <-c.StopChan:
			return

		case <-ticker.C:
			if refreshInterval == 0 {
				break
			}
			raw := time.Now().Format(eventTimeFormat) + " -9998 CAM " + EventWatcherRefreshed.String() + " every " + refreshInterval.String()
			if err := c.Refresh(); err != nil {
				raw = time.Now().Format(eventTimeFormat) + " -9997 CAM " + EventWatcherRefreshFail.String() + ": " + err.Error()
			}
			c.EventChan <- c.parseEvent(raw)

		case event := <-c.EventChan:
			c.RLock()
			event.callBacks(c.EventBinds)
			c.RUnlock()
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
			 20190114200911 104519 CAM2 MOTION
			 20190114201129 104520 CAM5 DISARM_C
			 20190114201129 104521 CAM5 DISARM_M
			 20190114201129 104522 CAM5 DISARM_A
			 20190114201129 104523 CAM5 OFFLINE
			 20190114201139 104524 CAM0 ERROR 10,835 Error communicating with the network device "Porch".
			 20190114201155 104525 CAM5 ERROR 70900,800 Error communicating with the network device "Pool".
			 20190114201206 104526 CAM5 ONLINE
			 20190114201206 104527 CAM5 ARM_C
			 20190114201206 104528 CAM5 ARM_M
			 20190114201206 104529 CAM5 ARM_A */
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
	} else if cameraNum, err := strconv.Atoi(parts[2][3:]); err != nil {
		e.Camera = nil
		e.Errors = append(e.Errors, ErrorCAMParseFail)
	} else if e.Camera = c.GetCamera(cameraNum); e.Camera == nil {
		e.Errors = append(e.Errors, ErrorCAMParseFail)
	}
	// Parse and convert the type string to EventType.
	e.Event = EventName(strings.Split(parts[3], " ")[0])
	// Check if the type we just converted is a known event.
	if e.Event.Event() == EventUnknownEvent.Event() {
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
