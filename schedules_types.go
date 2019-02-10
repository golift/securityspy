package securityspy

import (
	"encoding/xml"
)

// CameraMode is a set of constants to deal with three specific camera modes.
type CameraMode rune

// Camera modes is used by the Camera scheduling methods. Use these constants
// as inputs to a Camera's schedule methods.
const (
	CameraModeAll        CameraMode = 'X'
	CameraModeMotion     CameraMode = 'M'
	CameraModeActions    CameraMode = 'A'
	CameraModeContinuous CameraMode = 'C'
)

// scheduleContainer allows unmarshalling of ScheduleOverrides and SchedulePresets into a map.
type scheduleContainer map[int]string

func (m *scheduleContainer) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for (*m) = make(scheduleContainer); ; {
		var schedule struct {
			Name string `xml:"name"`
			ID   int    `xml:"id"`
		}
		token, err := d.Token()
		if err != nil {
			return err
		}
		switch e := token.(type) {
		case xml.StartElement:
			if err = d.DecodeElement(&schedule, &e); err != nil {
				return err
			}
			(*m)[schedule.ID] = schedule.Name
		case xml.EndElement:
			if e == start.End() {
				return nil
			}
		}
	}
}
