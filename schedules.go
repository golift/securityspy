package securityspy

/* There are no docuemnted methods for updating or changing schedules. */

// Schedule is for arming and disarming motion, capture, actions, etc.
type Schedule struct {
	Name string `xml:"name"` // Unarmed 24/7, Armed 24/7,...
	ID   int    `xml:"id"`   // 0, 1, 2, 3
}

// SchedulePresets defines the presets for schedules returned from SecuritySpy
// (I dont have any (yet))
type SchedulePresets interface{}
