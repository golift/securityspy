package securityspy

import (
	"net/url"
	"strconv"
)

/* There are no methods for updating or changing schedules through the API,
   but you can assign already-existing schedules to cameras using a Camera
   method. You can also assign hard coded Schedule Overrides to cameras using
	 a different method.

	 There is one "schedule" method for invoking a system-wide schedule preset.
*/

// SetSchedulePreset invokes a schedule preset. This [may/will] affect all camera arm modes.
// Find preset IDs you can pass into this method at server.Info.SchedulePresets
func (s *Server) SetSchedulePreset(presetID int) error {
	params := make(url.Values)
	params.Set("id", strconv.Itoa(presetID))
	return s.simpleReq("++ssSetPreset", params, -1)
}
