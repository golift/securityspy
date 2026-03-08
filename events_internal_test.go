package securityspy

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golift.io/securityspy/v2/server"
)

func TestEventSelectorNonBlockingChannels(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events := &Events{
		eventChan:  make(chan *Event, 2),
		eventBinds: make(map[EventType][]func(event Event)),
		eventChans: make(map[EventType][]chan Event),
		Running:    true,
	}

	events.BindChan(EventAllEvents, make(chan Event)) // unbuffered, no receiver

	done := make(chan struct{})

	go func() {
		events.eventStreamSelector(ctx, false)
		close(done)
	}()

	events.eventChan <- &Event{Type: EventArmMotion}

	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("event selector blocked on subscriber channel")
	}
}

func TestWatchStopIdempotent(t *testing.T) {
	t.Parallel()

	srv := NewMust(&server.Config{
		URL:     "http://127.0.0.1:1/",
		Timeout: server.Duration{Duration: 100 * time.Millisecond},
	})
	srv.Cameras = &Cameras{server: srv}

	srv.Events.Watch(10*time.Millisecond, false)
	time.Sleep(20 * time.Millisecond)
	srv.Events.Stop(false)
	srv.Events.Stop(false)

	srv.Events.mu.RLock()
	defer srv.Events.mu.RUnlock()

	require.False(t, srv.Events.Running)
}
