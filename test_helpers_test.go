package securityspy_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"golift.io/securityspy/v2"
	"golift.io/securityspy/v2/server"
)

type recordedRequest struct {
	Path  string
	Query url.Values
}

type requestRecorder struct {
	mu   sync.Mutex
	reqs []recordedRequest
}

const systemInfoPath = "/++systemInfo"

func (r *requestRecorder) add(req *http.Request) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.reqs = append(r.reqs, recordedRequest{
		Path:  req.URL.Path,
		Query: req.URL.Query(),
	})
}

func (r *requestRecorder) findLast(path string) (recordedRequest, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for idx := len(r.reqs) - 1; idx >= 0; idx-- {
		if r.reqs[idx].Path == path {
			return r.reqs[idx], true
		}
	}

	return recordedRequest{}, false
}

func newTestServer(t *testing.T, handler http.HandlerFunc) *securityspy.Server {
	t.Helper()

	fakeServer := httptest.NewServer(handler)
	t.Cleanup(fakeServer.Close)

	return securityspy.NewMust(&server.Config{
		Username: "user",
		Password: "pass",
		URL:      fakeServer.URL + "/",
	})
}

func testServerWithCamera(t *testing.T) (*securityspy.Server, *requestRecorder, *securityspy.Camera) {
	t.Helper()

	recorder := &requestRecorder{}
	serverObj := newTestServer(t, func(resp http.ResponseWriter, req *http.Request) {
		recorder.add(req)

		switch req.URL.Path {
		case systemInfoPath:
			resp.Header().Set("Content-Type", "application/xml")
			_, _ = resp.Write([]byte(testSystemInfo))
		case "/++ssControlContinuous", "/++ssControlMotionCapture", "/++ssControlActions",
			"/++triggermd", "/++ssSetSchedule", "/++ssSetOverride",
			"/++ptz/command", "/++ssSetPreset":
			_, _ = resp.Write([]byte("OK"))
		default:
			http.NotFound(resp, req)
		}
	})

	require.NoError(t, serverObj.Refresh())

	camera := serverObj.Cameras.ByNum(1)
	require.NotNil(t, camera)

	return serverObj, recorder, camera
}
