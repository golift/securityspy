package securityspy_test

import (
	"net/http"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestRefreshConcurrentReaders is a race-detector test. Refresh() replaces
// Cameras, Groups and Info, and apps commonly call it from a background retry
// loop or an event handler while requests read those same fields. The Get
// accessors are the only safe way to read them, so the readers below run for
// as long as the refresher does.
//
// The workers report with t.Errorf rather than require: FailNow may only be
// called from the goroutine running the test.
func TestRefreshConcurrentReaders(t *testing.T) {
	t.Parallel()

	serverObj, _, _ := testServerWithCamera(t)
	done := make(chan struct{})

	var wait sync.WaitGroup

	wait.Go(func() {
		defer close(done)

		for range 25 {
			err := serverObj.Refresh()
			if err != nil {
				t.Errorf("Refresh: %v", err)

				return
			}
		}
	})

	wait.Go(func() {
		for !isDone(done) {
			cams := serverObj.GetCameras()
			if cams == nil {
				t.Error("GetCameras returned nil during a refresh")

				return
			}

			if cams.ByNum(3) == nil {
				t.Error("camera 3 went missing during a refresh")

				return
			}

			if len(cams.All()) == 0 {
				t.Error("camera list emptied during a refresh")

				return
			}
		}
	})

	wait.Go(func() {
		for !isDone(done) {
			info := serverObj.GetInfo()
			if info == nil {
				t.Error("GetInfo returned nil during a refresh")

				return
			}

			_ = info.Version
			_ = serverObj.GetGroups()
		}
	})

	wait.Wait()
}

// TestRefreshDoesNotBlockReaders: a refresh holds the write lock only for the
// swap, so a slow (or hung) systemInfo request must not stall readers. The
// handler parks the second refresh mid-request; if the refresh held the lock
// across the round trip, the reads below would block until the test timed out.
func TestRefreshDoesNotBlockReaders(t *testing.T) {
	t.Parallel()

	var (
		requests atomic.Int64
		parked   = make(chan struct{})
		release  = make(chan struct{})
	)

	serverObj := newTestServer(t, func(resp http.ResponseWriter, req *http.Request) {
		if req.URL.Path != systemInfoPath {
			http.NotFound(resp, req)

			return
		}

		if requests.Add(1) > 1 { // let the first refresh load the snapshot
			close(parked)
			<-release
		}

		resp.Header().Set("Content-Type", "application/xml")
		_, _ = resp.Write([]byte(testSystemInfoV6))
	})

	require.NoError(t, serverObj.Refresh())

	refreshed := make(chan error, 1)
	go func() { refreshed <- serverObj.Refresh() }()

	<-parked

	// The refresh is parked mid-request; reads still come from the old snapshot.
	cams := serverObj.GetCameras()
	if cams == nil || cams.ByNum(3) == nil {
		t.Error("readers were blocked by an in-flight refresh")
	}

	if serverObj.GetInfo() == nil {
		t.Error("GetInfo was blocked by an in-flight refresh")
	}

	close(release)
	require.NoError(t, <-refreshed)
}

func isDone(done <-chan struct{}) bool {
	select {
	case <-done:
		return true
	default:
		return false
	}
}
