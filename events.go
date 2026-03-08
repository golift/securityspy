package securityspy

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// String provides a description of an event.
func (e *Event) String() string {
	txt := EventName(e.Type)
	if txt == "" {
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
// Closes all channels that were passed to BindChan if closeChans=true.
// Stop writing to the channels with Custom() before calling Stop().
func (e *Events) Stop(closeChans bool) {
	e.mu.Lock()
	running := e.Running
	cancel := e.cancel
	stream := e.stream
	e.Running = false
	e.cancel = nil
	e.ctx = nil
	e.stream = nil
	e.mu.Unlock()

	if running {
		if cancel != nil {
			cancel()
		}

		if stream != nil {
			_ = stream.Close()
		}

		e.wg.Wait()
	}

	if !closeChans {
		return
	}

	e.chans.Lock()
	defer e.chans.Unlock()

	closed := make(map[chan Event]struct{})

	for _, chans := range e.eventChans {
		for idx := range chans {
			if _, ok := closed[chans[idx]]; ok {
				continue
			}

			close(chans[idx])
			closed[chans[idx]] = struct{}{}
		}
	}
}

// UnbindAll removes all event bindings and channels.
func (e *Events) UnbindAll() {
	e.binds.Lock()
	e.chans.Lock()

	defer func() {
		e.binds.Unlock()
		e.chans.Unlock()
	}()

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
// EventType is a set of constants that begin with Event*.
func (e *Events) UnbindFunc(event EventType) {
	e.binds.Lock()
	defer e.binds.Unlock()

	delete(e.eventBinds, event)
}

// Watch kicks off the routines to watch the eventStream and fire callback bindings.
// If your application relies on event stream messages, call this at least once
// to connect the stream. If you have no call back functions or channels then do not
// call this. Call Stop() to close the connection when you're done with it.
func (e *Events) Watch(retryInterval time.Duration, refreshOnConfigChange bool) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.Running {
		return
	}

	e.ctx, e.cancel = context.WithCancel(context.Background())
	watchCtx := e.ctx
	e.eventChan = make(chan *Event, EventBuffer)
	e.Running = true

	e.wg.Go(func() { e.eventStreamSelector(watchCtx, refreshOnConfigChange) })
	e.wg.Go(func() { e.eventStreamScanner(watchCtx, retryInterval) })
}

// Custom fires an event into the running event Watcher. Any functions or
// channels bound to the CUSTOM Event type will also be called.
func (e *Events) Custom(cameraNum int, msg string) {
	e.custom(EventStreamCustom, -11000, cameraNum, msg)
}

// custom allows a quick way to make events.
func (e *Events) custom(eventType EventType, eventID, cam int, msg string) {
	now := time.Now().Round(time.Second)

	var camera *Camera
	if e.server.Cameras != nil {
		camera = e.server.Cameras.ByNum(cam)
	}

	e.enqueue(&Event{
		Time:   now,
		When:   now,
		ID:     eventID,
		Msg:    string(eventType) + " " + msg,
		Type:   eventType,
		Camera: camera,
	})
}

/* INTERFACE HELPER METHODS FOLLOW */

// eventStreamScanner connects to the securityspy event stream and fires events into a channel.
//
//nolint:cyclop // but it runs forever!
func (e *Events) eventStreamScanner(ctx context.Context, retryInterval time.Duration) {
	for {
		if ctx.Err() != nil {
			return
		}

		stream, err := e.eventStreamConnect(ctx)
		if err != nil {
			e.custom(EventStreamDisconnect, -10000, -1, err.Error())

			select {
			case <-ctx.Done():
				return
			case <-time.After(retryInterval):
				continue
			}
		}

		scanner := bufio.NewScanner(stream)
		scanner.Split(scanLinesCR)

		for scanner.Scan() {
			// Constantly scan for new events, then report them to the event channel.
			if text := scanner.Text(); strings.Count(text, " ") > 2 { //nolint:mnd // we need at least 2.
				e.enqueue(e.UnmarshalEvent(text))
			}

			if ctx.Err() != nil {
				break
			}
		}

		err = scanner.Err()
		_ = stream.Close()
		e.clearStream(stream)

		if ctx.Err() != nil {
			return
		}

		msg := "Connection Closed"
		if err != nil && !errors.Is(err, ErrDisconnect) {
			msg = err.Error()
		}

		e.custom(EventStreamDisconnect, -10000, -1, msg)

		select {
		case <-ctx.Done():
			return
		case <-time.After(retryInterval):
		}
	}
}

// eventStreamConnect establishes a connection to the event stream and passes off the http Reader.
func (e *Events) eventStreamConnect(ctx context.Context) (io.ReadCloser, error) {
	client := e.server.HTTPClient()
	client.Timeout = 0

	resp, err := e.server.GetContextClient(ctx, "++eventStream", url.Values{"version": []string{"3"}}, client)
	if err != nil {
		return nil, fmt.Errorf("connecting event stream: %w", err)
	}

	e.mu.Lock()
	e.stream = resp.Body
	e.mu.Unlock()

	e.custom(EventStreamConnect, -9999, -1, EventName(EventStreamConnect))

	return resp.Body, nil
}

// eventStreamSelector watches the event channel.
// Fires bound event call back functions.
// Also reconnects to the event stream if the connection fails.
// There is a "loop" that occurs among the eventStream* methods.
// Stop() properly handles the shutdown of the loop, so if can be safely restarted w/ Watch().
func (e *Events) eventStreamSelector(ctx context.Context, refreshOnConfigChange bool) { //nolint:cyclop // oh well?
	for {
		var (
			event *Event
			ok    bool //nolint:varnamelen // ok is a valid variable name.
		)

		select {
		case <-ctx.Done():
			return
		case event, ok = <-e.eventChan:
			if !ok {
				return
			}
		}

		switch event.Type {
		case eventStreamStop:
			return
		case EventConfigChange:
			if refreshOnConfigChange {
				e.serverRefresh(ctx)
			}
		}

		for _, callback := range e.callbacksFor(event.Type) {
			if callback != nil {
				go callback(*event)
			}
		}

		for _, ch := range e.channelsFor(event.Type) {
			select {
			case ch <- *event:
			default:
			}
		}
	}
}

func (e *Events) serverRefresh(ctx context.Context) {
	if err := e.server.RefreshContext(ctx); err != nil {
		e.custom(EventWatcherRefreshFail, -9997, -1, err.Error())

		return
	}

	e.custom(EventWatcherRefreshed, -9998, -1, EventName(EventWatcherRefreshed))
}

/* 	Example Event Stream Flow:
(new, v5)
20190927092026 3 3 CLASSIFY HUMAN 99
20190927092026 4 3 TRIGGER_M 9
20190927092036 5 3 CLASSIFY HUMAN 5 VEHICLE 95
20190927092040 5 X NULL
20190927092050 6 3 FILE /Volumes/VolName/Cam/2019-07-26/26-07-2019 15-52-00 C Cam.m4v
20190927092055 7 3 DISARM_M
20190927092056 8 3 OFFLINE
*/

// UnmarshalEvent turns raw text into an Event that can fire callbacks.
// You generally shouldn't need to call this method, it's exposed for convenience.
/* [TIME] is specified in the order: "year, month, day, hour, minute, second" and is always 14 characters long.
 * [EVENT NUMBER] increases by 1 for each subsequent event.
 * [CAMERA NUMBER] specifies the camera that this event relates to, for example CAM15 for camera number 15.
 * [EVENT] describes the event: ARM_C, DISARM_C, ARM_M, DISARM_M, ARM_A, DISARM_A, ERROR,
           CONFIGCHANGE, MOTION, OFFLINE, ONLINE */
//
//nolint:cyclop,funlen,mnd // Events are hard.
func (e *Events) UnmarshalEvent(text string) *Event {
	var (
		err      error
		parts    = strings.SplitN(text, " ", 4) //nolint:mnd // events have 4 parts...
		newEvent = &Event{Msg: text, ID: -1, Time: time.Now()}
		// Parse the time stamp; append the Offset from ++systemInfo to get the right time-location.
		eventTime string
	)

	if len(parts) < 4 {
		newEvent.Errors = append(newEvent.Errors, ErrUnknownEvent)
		newEvent.Type = EventUnknownEvent

		return newEvent
	}

	newEvent.Msg = parts[3]
	eventTime = fmt.Sprintf("%v%+03.0f", parts[0], e.server.Info.GmtOffset.Hours())

	//nolint:gosmopolitan // The event stream uses the system's local time.
	if newEvent.When, err = time.ParseInLocation(EventTimeFormat+"-07", eventTime, time.Local); err != nil {
		newEvent.When = time.Now()
		newEvent.Errors = append(newEvent.Errors, ErrDateParseFail)
	}

	// Parse the ID
	if newEvent.ID, err = strconv.Atoi(parts[1]); err != nil {
		newEvent.ID = BadID
		newEvent.Errors = append(newEvent.Errors, ErrIDParseFail)
	}

	// Parse the camera number.
	parts[2] = strings.TrimPrefix(parts[2], "CAM")
	if parts[2] != "X" {
		if cameraNum, err := strconv.Atoi(parts[2]); err != nil {
			newEvent.Errors = append(newEvent.Errors, ErrCAMParseFail)
		} else if newEvent.Camera = e.server.Cameras.ByNum(cameraNum); newEvent.Camera == nil {
			newEvent.Errors = append(newEvent.Errors, ErrCAMMissing)
		}
	}

	// Parse and convert the type string to EventType.
	parts = strings.Split(newEvent.Msg, " ")

	newEvent.Type = EventType(parts[0])
	// Check if the type we just converted is a known event.
	if name := EventName(newEvent.Type); name == "" {
		newEvent.Errors = append(newEvent.Errors, ErrUnknownEvent)
		newEvent.Type = EventUnknownEvent
	}

	// If this is a trigger-type event, add the trigger reason(s)
	if (newEvent.Type == EventTriggerAction || newEvent.Type == EventTriggerMotion) && len(parts) == 2 {
		b, _ := strconv.Atoi(parts[1])
		msg := ""

		// Check if this bitmask contains any of our known reasons.
		for flag, txt := range Reasons() {
			if b&int(flag) != 0 {
				if msg != "" {
					msg += ", "
				}

				newEvent.Reasons = append(newEvent.Reasons, flag)
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

func (e *Events) enqueue(event *Event) {
	if event == nil {
		return
	}

	e.mu.RLock()
	chn := e.eventChan
	ctx := e.ctx
	running := e.Running
	e.mu.RUnlock()

	if !running || chn == nil || ctx == nil {
		return
	}

	select {
	case chn <- event:
	case <-ctx.Done():
	}
}

func (e *Events) clearStream(stream io.ReadCloser) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.stream == stream {
		e.stream = nil
	}
}

func (e *Events) callbacksFor(eventType EventType) []func(Event) {
	e.binds.RLock()
	defer e.binds.RUnlock()

	callbacks := make([]func(Event), 0, len(e.eventBinds[eventType])+len(e.eventBinds[EventAllEvents])+1)
	if vals, ok := e.eventBinds[eventType]; ok {
		callbacks = append(callbacks, vals...)
	} else if eventType != EventUnknownEvent {
		callbacks = append(callbacks, e.eventBinds[EventUnknownEvent]...)
	}

	callbacks = append(callbacks, e.eventBinds[EventAllEvents]...)

	return callbacks
}

func (e *Events) channelsFor(eventType EventType) []chan Event {
	e.chans.RLock()
	defer e.chans.RUnlock()

	channels := make([]chan Event, 0, len(e.eventChans[eventType])+len(e.eventChans[EventAllEvents]))
	channels = append(channels, e.eventChans[eventType]...)
	channels = append(channels, e.eventChans[EventAllEvents]...)

	return channels
}

// scanLinesCR is a custom bufio.Scanner to read SecuritySpy eventStream.
func scanLinesCR(data []byte, atEOF bool) (int, []byte, error) {
	if atEOF && len(data) == 0 {
		return 0, nil, ErrDisconnect
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
