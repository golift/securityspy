package securityspy

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// String provides a description of an event.
func (e *Event) String() string {
	txt, ok := EventNames[e.Type]
	if !ok {
		return UnknownEventText
	}
	return txt
}

/* Events methods follow. */

// BindFunc binds a call-back function to an Event in SecuritySpy.
// Use this to receive incoming events via a callback method in a go routine.
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
// Use this to receive incoming events over a channel.
// Avoid using unbuffered channels as they may block further event processing.
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

// Stop stops Watch() loops and disconnects from the event stream.
// No further callback messages will fire after this is called.
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
// EventType is a set of constants that begin with Event*
func (e *Events) UnbindFunc(event EventType) {
	e.binds.Lock()
	defer e.binds.Unlock()
	delete(e.eventBinds, event)
}

// Watch kicks off the routines to watch the eventStream and fire callback bindings.
// If your application relies on event stream messages, call this at least once
// to connect the stream. If you have no call back functions or channels then do not
// call this.
func (e *Events) Watch(retryInterval time.Duration, refreshOnConfigChange bool) {
	e.Running = true
	e.eventChan = make(chan Event, 1000) // allow 1000 events to buffer
	e.stopChan = make(chan bool)
	go e.eventChannelSelector(refreshOnConfigChange)
	e.eventStreamScanner(retryInterval)
}

// Custom fires an event into the running event Watcher. Any functions or
// channels bound to the CUSTOM Event type will also be called.
func (e *Events) Custom(cameraNum int, msg string) {
	if !e.Running {
		return
	}
	e.custom(EventStreamCustom, -11000, cameraNum, msg)
}

// custom allows a quick way to make events.
func (e *Events) custom(t EventType, id int, cam int, msg string) {
	e.eventChan <- Event{
		Time:   time.Now().Round(time.Second),
		When:   time.Now().Round(time.Second),
		ID:     id,
		Msg:    string(t) + " " + msg,
		Type:   t,
		Camera: e.server.Cameras.ByNum(cam),
	}
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
				e.eventChan <- e.UnmarshalEvent(text)
			}
		}
		if err := scanner.Err(); err != nil {
			e.custom(EventStreamDisconnect, -10000, -1, err.Error())
			_ = body.Close()
			time.Sleep(retryInterval)
			body, scanner = e.eventStreamConnect(retryInterval)
			scanner.Split(scanLinesCR)
		}
	}
}

// eventStreamConnect establishes a connection to the event stream and passes off the http Reader.
func (e *Events) eventStreamConnect(retryInterval time.Duration) (io.ReadCloser, *bufio.Scanner) {
	httpClient := e.server.api.getClient(0)
	resp, err := e.server.api.secReq("++eventStream", url.Values{"version": []string{"3"}}, httpClient)
	for err != nil {
		// This for loops attempts to reconnect if the stream is down.
		e.custom(EventStreamDisconnect, -9999, -1, EventNames[EventStreamDisconnect]+" "+err.Error())
		time.Sleep(retryInterval)
		if !e.Running {
			return nil, nil // Stopped externally while sleeping, bail out.
		}
		resp, err = e.server.api.secReq("++eventStream", url.Values{"version": []string{"3"}}, httpClient)
	}
	e.custom(EventStreamConnect, -9999, -1, EventNames[EventStreamConnect])
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
					if err := e.server.Refresh(); err != nil {
						e.custom(EventWatcherRefreshFail, -9997, -1, err.Error())
						return
					}
					e.custom(EventWatcherRefreshed, -9998, -1, EventNames[EventWatcherRefreshed])
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

// UnmarshalEvent turns raw text into an Event that can fire callbacks.
// You generally shouldn't need to call this method, it's exposed for convenience.
/* [TIME] is specified in the order year, month, day, hour, minute, second and is always 14 characters long
 * [EVENT NUMBER] increases by 1 for each subsequent event
 * [CAMERA NUMBER] specifies the camera that this event relates to, for example CAM15 for camera number 15
 * [EVENT] describes the event: ARM_C, DISARM_C, ARM_M, DISARM_M, ARM_A, DISARM_A, ERROR, CONFIGCHANGE, MOTION, OFFLINE, ONLINE
	Example Event Stream Flow:
	(old, v4)
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
	20190114201206 104529 CAM5 ARM_A
	(new, v5)
	20190927092026 3 3 CLASSIFY HUMAN 99
	20190927092026 4 3 TRIGGER_M 9
	20190927092036 5 3 CLASSIFY HUMAN 5 VEHICLE 95
	20190927092040 5 X NULL
	20190927092050 6 3 FILE /Volumes/VolName/Cam/2019-07-26/26-07-2019 15-52-00 C Cam.m4v
	20190927092055 7 3 DISARM_M
	20190927092056 8 3 OFFLINE */
func (e *Events) UnmarshalEvent(text string) Event {
	var err error
	parts := strings.SplitN(text, " ", 4)
	newEvent := Event{Msg: parts[3], ID: -1, Time: time.Now()}

	// Parse the time stamp; append the Offset from ++systemInfo to get the right time-location.
	eventTime := fmt.Sprintf("%v%+03.0f", parts[0], e.server.Info.GmtOffset.Hours())
	if newEvent.When, err = time.ParseInLocation(EventTimeFormat+"-07", eventTime, time.Local); err != nil {
		newEvent.When = time.Now()
		newEvent.Errors = append(newEvent.Errors, ErrorDateParseFail)
	}

	// Parse the ID
	if newEvent.ID, err = strconv.Atoi(parts[1]); err != nil {
		newEvent.ID = -2
		newEvent.Errors = append(newEvent.Errors, ErrorIDParseFail)
	}

	// Parse the camera number.
	parts[2] = strings.TrimPrefix(parts[2], "CAM")
	if parts[2] != "X" {
		if cameraNum, err := strconv.Atoi(parts[2]); err != nil {
			newEvent.Errors = append(newEvent.Errors, ErrorCAMParseFail)
		} else if newEvent.Camera = e.server.Cameras.ByNum(cameraNum); newEvent.Camera == nil {
			newEvent.Errors = append(newEvent.Errors, ErrorCAMMissing)
		}
	}

	// Parse and convert the type string to EventType.
	parts = strings.Split(newEvent.Msg, " ")
	newEvent.Type = EventType(parts[0])
	// Check if the type we just converted is a known event.
	if _, ok := EventNames[newEvent.Type]; !ok {
		newEvent.Errors = append(newEvent.Errors, ErrorUnknownEvent)
		newEvent.Type = EventUnknownEvent
	}

	// If this is a trigger-type event, add the trigger reason(s)
	if newEvent.Type == EventTriggerAction || newEvent.Type == EventTriggerMotion && len(parts) == 2 {
		b, _ := strconv.Atoi(parts[1])
		msg := ""
		// Check if this bitmask contains any of our known reasons.
		for flag, txt := range Reasons {
			if b&int(flag) != 0 {
				if msg != "" {
					msg += ", "
				}
				msg += txt
			}
		}
		newEvent.Msg += " - Reasons: " + msg
		if msg == "" {
			newEvent.Msg += UnknownReasonText
		}
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
