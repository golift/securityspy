package securityspy

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

/* Event methods follow. */

// String provides a description of an event.
func (e *Event) String() string {
	switch e.Type {
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
		return unknownEventText
	case EventAllEvents:
		return "Any Event"
	case EventWatcherRefreshed:
		return "SystemInfo Refresh Success"
	case EventWatcherRefreshFail:
		return "SystemInfo Refresh Failure"
	case EventStreamCustom:
		return "Custom Event"
	}
	return unknownEventText
}

/* Events methods follow. */

// BindFunc binds a call-back function to an Event in SecuritySpy.
func (e *Events) BindFunc(event EventType, callBack func(Event)) {
	if callBack == nil {
		return
	}
	e.binds.Lock()
	defer e.binds.Unlock()
	if val, ok := e.eventBinds[event]; ok {
		e.eventBinds[event] = append(val, callBack)
		return
	}
	e.eventBinds[event] = []func(Event){callBack}
}

// BindChan binds a receiving channel to an Event in SecuritySpy.
func (e *Events) BindChan(event EventType, channel chan Event) {
	if channel == nil {
		return
	}
	e.chans.Lock()
	defer e.chans.Unlock()
	if val, ok := e.eventChans[event]; ok {
		e.eventChans[event] = append(val, channel)
		return
	}
	e.eventChans[event] = []chan Event{channel}
}

// Stop stops Watch() loops
func (e *Events) Stop() {
	defer func() { e.Running = false }()
	if e.Running {
		e.stopChan <- e.Running
	}
}

// UnbindAll removes all event bindings and channels.
func (e *Events) UnbindAll() {
	e.binds.Lock()
	e.chans.Lock()
	defer e.binds.Unlock()
	defer e.chans.Unlock()
	e.eventBinds = make(map[EventType][]func(Event))
	e.eventChans = make(map[EventType][]chan Event)
}

// UnbindChan removes all bound channels for a particular event.
func (e *Events) UnbindChan(event EventType) {
	e.chans.Lock()
	defer e.chans.Unlock()
	delete(e.eventChans, event)
}

// UnbindFunc removes all bound callbacks for a particular event.
func (e *Events) UnbindFunc(event EventType) {
	e.binds.Lock()
	defer e.binds.Unlock()
	delete(e.eventBinds, event)
}

// Watch kicks off the routines to watch the eventStream and fire callback bindings.
func (e *Events) Watch(retryInterval time.Duration, refreshOnConfigChange bool) {
	e.Running = true
	e.eventChan = make(chan Event, 10) // allow 10 events to buffer
	e.stopChan = make(chan bool)
	go e.eventChannelSelector(refreshOnConfigChange)
	e.eventStreamScanner(retryInterval)
}

// Custom fires an event into the running event Watcher.
func (e *Events) Custom(cameraNum int, msg string) {
	if !e.Running {
		return
	}
	e.eventChan <- e.parseEvent(time.Now().Format(eventTimeFormat) +
		" -11000 CAM" + strconv.Itoa(cameraNum) + " " + string(EventStreamCustom) + " " + msg)
}

/* INTERFACE HELPER METHODS FOLLOW */

// eventStreamScanner connects to the securityspy event stream and fires events into a channel.
func (e *Events) eventStreamScanner(retryInterval time.Duration) {
	body, scanner := e.eventStreamConnect(retryInterval)
	if scanner != nil {
		scanner.Split(scanLinesCR)
	}
	for {
		if !e.Running {
			if body != nil {
				_ = body.Close()
			}
			return // we all done here. stop got called
		}
		// Constantly scan for new events, then report them to the event channel.
		if scanner != nil && scanner.Scan() {
			if text := scanner.Text(); strings.Count(text, " ") > 2 {
				e.eventChan <- e.parseEvent(text)
			}
		}
		if err := scanner.Err(); err != nil {
			raw := time.Now().Format(eventTimeFormat) + " -10000 CAM " + string(EventStreamDisconnect) + " " + err.Error()
			e.eventChan <- e.parseEvent(raw)
			_ = body.Close()
			time.Sleep(retryInterval)
			body, scanner = e.eventStreamConnect(retryInterval)
			scanner.Split(scanLinesCR)
		}
	}
}

// eventStreamConnect establishes a connection to the event stream and passes off the http Reader.
func (e *Events) eventStreamConnect(retryInterval time.Duration) (io.ReadCloser, *bufio.Scanner) {
	resp, err := e.server.secReq("++eventStream", nil, 0)
	for err != nil {
		raw := time.Now().Format(eventTimeFormat) + " -9999 CAM " + string(EventStreamDisconnect) + " " + err.Error()
		e.eventChan <- e.parseEvent(raw)
		time.Sleep(retryInterval)
		if !e.Running {
			return nil, nil
		}
		resp, err = e.server.secReq("++eventStream", nil, 0)
	}
	raw := time.Now().Format(eventTimeFormat) + " -1 CAM " + string(EventStreamConnect)
	e.eventChan <- e.parseEvent(raw)
	return resp.Body, bufio.NewScanner(resp.Body)
}

// eventChannelSelector watches a few internal channels for events and updates.
// Fires bound event call back functions.
func (e *Events) eventChannelSelector(refreshOnConfigChange bool) {
	done := make(chan bool)
	for {
		// Watch for new events, a stop signal, or a refresh interval.
		select {
		case <-e.stopChan:
			return
		case event := <-e.eventChan:
			if refreshOnConfigChange && event.Type == EventConfigChange {
				go func() {
					raw := time.Now().Format(eventTimeFormat) + " -9998 CAM " + string(EventWatcherRefreshed)
					if err := e.server.Refresh(); err != nil {
						raw = time.Now().Format(eventTimeFormat) + " -9997 CAM " + string(EventWatcherRefreshFail) + " " + err.Error()
					}
					e.eventChan <- e.parseEvent(raw)
				}()
			}
			go func() {
				e.binds.RLock()
				event.callBacks(e.eventBinds)
				e.binds.RUnlock()
			}() // these can punt and fire in any order.
			go func() {
				e.chans.RLock()
				event.eventChans(e.eventChans)
				e.chans.RUnlock()
				done <- true
			}()
			<-done // channels block to keep proper ordering.
		}
	}
}

// parseEvent turns raw text into an Event that can fire callbacks.
func (e *Events) parseEvent(text string) Event {
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
	newEvent := Event{Msg: parts[3], ID: -1, Time: time.Now()}
	// Parse the time stamp; append the Offset from ++systemInfo to get the right time-location.
	eventTime := fmt.Sprintf("%v%+03.0f", parts[0], e.server.Info.GmtOffset.Hours())
	if newEvent.When, err = time.ParseInLocation(eventTimeFormat+"-07", eventTime, time.Local); err != nil {
		newEvent.When = time.Now()
		newEvent.Errors = append(newEvent.Errors, ErrorDateParseFail)
	}

	// Parse the ID
	if newEvent.ID, err = strconv.Atoi(parts[1]); err != nil {
		newEvent.ID = -2
		newEvent.Errors = append(newEvent.Errors, ErrorIDParseFail)
	}
	// Parse the camera number.
	if !strings.HasPrefix(parts[2], "CAM") || len(parts[2]) < 4 {
		newEvent.Errors = append(newEvent.Errors, ErrorCAMMissing)
	} else if cameraNum, err := strconv.Atoi(parts[2][3:]); err != nil {
		newEvent.Camera = nil
		newEvent.Errors = append(newEvent.Errors, ErrorCAMParseFail)
	} else if newEvent.Camera = e.server.Cameras.ByNum(cameraNum); newEvent.Camera == nil {
		newEvent.Errors = append(newEvent.Errors, ErrorCAMParseFail)
	}
	// Parse and convert the type string to EventType.
	newEvent.Type = EventType(strings.Split(parts[3], " ")[0])
	// Check if the type we just converted is a known event.
	if newEvent.String() == unknownEventText {
		newEvent.Errors = append(newEvent.Errors, ErrorUnknownEvent)
		newEvent.Type = EventUnknownEvent
	}
	return newEvent
}

// callBacks is run for each event to execute callback functions.
func (e *Event) callBacks(binds map[EventType][]func(Event)) {
	callbacks := func(callbacks []func(Event)) {
		for _, callBack := range callbacks {
			if callBack != nil {
				go callBack(*e) // Send it off!
			}
		}
	}
	if _, ok := binds[e.Type]; ok {
		callbacks(binds[e.Type])
	} else if _, ok := binds[EventUnknownEvent]; ok && e.Type != EventUnknownEvent {
		callbacks(binds[EventUnknownEvent])
	}
	if _, ok := binds[EventAllEvents]; ok {
		callbacks(binds[EventAllEvents])
	}
}

// eventChans is run for each event to notify external channels
func (e *Event) eventChans(chans map[EventType][]chan Event) {
	if chans, ok := chans[e.Type]; ok {
		for i := range chans {
			chans[i] <- *e
		}
	}
	if chans, ok := chans[EventAllEvents]; ok {
		for i := range chans {
			chans[i] <- *e
		}
	}
}

// scanLinesCR is a custom bufio.Scanner to read SecuritySpy eventStream.
func scanLinesCR(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, ErrorDisconnect
	}
	if i := bytes.IndexByte(data, '\r'); i >= 0 {
		// We have a full CR-terminated line.
		return i + 1, data[0:i], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, io.ErrShortBuffer
	}
	// Request more data.
	return 0, nil, nil
}
