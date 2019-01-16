package securityspy

/* There are no docuemnted methods for updating or changing schedules. */

// schedule is for arming and disarming motion, capture, actions, etc.
type schedule struct {
	Name string `xml:"name"` // Unarmed 24/7, Armed 24/7,...
	ID   int    `xml:"id"`   // 0, 1, 2, 3
}

// schedulePresets defines the presets for schedules returned from SecuritySpy
type schedulePresets struct {
	Name string `xml:"name"` // MySchedule, NowOnly, Weekends
	ID   int    `xml:"id"`   // 0, 1, 2, 3
}
