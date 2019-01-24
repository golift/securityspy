package securityspy

import (
	"net/url"
	"strconv"
)

/* There are no methods for updating or changing schedules through the API,
   but you can assign already-existing schedules to cameras using the Camera
   methods. There is one method for invoking a schedule preset.
*/

// Schedules returns a list of pre-existing schedules that can be assigned to cameras.
func (s *Server) Schedules() []Schedule {
	return s.systemInfo.ScheduleList.Schedules
}

// SchedulePresets provides a list of presets one can pass into SetSchedulePreset()
func (s *Server) SchedulePresets() []Schedule {
	return s.systemInfo.SchedulePresetList.SchedulePresets
}

// SetSchedulePreset configures the schedule preset for a camera mode.
func (s *Server) SetSchedulePreset(schedule Schedule) error {
	params := make(url.Values)
	params.Set("id", strconv.Itoa(schedule.ID))
	return s.simpleReq("++ssSetPreset", params, -1)
}
