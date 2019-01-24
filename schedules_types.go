package securityspy

// CameraMode is a set of constants to deal with three specific camera modes.
type CameraMode rune

// Camera modes used by Camera scheduling methods.
const (
	CameraModeAll        CameraMode = '*'
	CameraModeMotion     CameraMode = 'M'
	CameraModeActions    CameraMode = 'A'
	CameraModeContinuous CameraMode = 'C'
)

// Schedule is for arming and disarming motion, capture, actions, etc.
// This types holds the schedule's name and its id. Pass this type into
// Camera schedule methods to set and re-assign schedules on camera parameters.
type Schedule struct {
	Name string `xml:"name"` // Unarmed 24/7, Armed 24/7,...
	ID   int    `xml:"id"`   // 0, 1, 2, 3
}
