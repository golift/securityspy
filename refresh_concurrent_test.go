package securityspy_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestRefreshConcurrentReaders is a race-detector test. Refresh() replaces
// Cameras, Groups and Info, and apps commonly call it from a background retry
// loop or an event handler while requests read those same fields. The Get
// accessors are the only safe way to read them, so the readers below run for
// as long as the refresher does.
func TestRefreshConcurrentReaders(t *testing.T) {
	t.Parallel()

	serverObj, _, _ := testServerWithCamera(t)
	done := make(chan struct{})

	var wait sync.WaitGroup

	wait.Go(func() {
		defer close(done)

		for range 25 {
			require.NoError(t, serverObj.Refresh())
		}
	})

	wait.Go(func() {
		for {
			select {
			case <-done:
				return
			default:
			}

			cams := serverObj.GetCameras()
			require.NotNil(t, cams)
			require.NotNil(t, cams.ByNum(3))
			require.NotEmpty(t, cams.All())
		}
	})

	wait.Go(func() {
		for {
			select {
			case <-done:
				return
			default:
			}

			info := serverObj.GetInfo()
			require.NotNil(t, info)
			_ = info.Version
			_ = serverObj.GetGroups()
		}
	})

	wait.Wait()
}
