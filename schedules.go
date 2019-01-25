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
// Find presets you can pass into this method at server.Info.SchedulePresets
func (s *Server) SetSchedulePreset(schedule SchedulePreset) error {
	params := make(url.Values)
	params.Set("id", strconv.Itoa(schedule.ID))
	return s.simpleReq("++ssSetPreset", params, -1)
}

// String provides a description of a Schedule Override.
func (e ScheduleOverride) String() string {
	switch e {
	case ScheduleOverrideNone:
		return "No Schedule Override"
	case ScheduleOverrideUnarmedUntilEvent:
		return "Unarmed Until Next Scheduled Event"
	case ScheduleOverrideArmedUntilEvent:
		return "Armed Until Next Scheduled Event"
	case ScheduleOverrideUnarmedOneHour:
		return "Unarmed For 1 Hour"
	case ScheduleOverrideArmedOneHour:
		return "Armed For 1 Hour"
	}
	return "Unknown Schedule Override"
}
