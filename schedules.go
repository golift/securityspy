package securityspy

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"strconv"
)

/* There are no methods for updating or changing schedules through the API,
   but you can assign already-existing schedules to cameras using a Camera
   method. You can also assign hard coded Schedule Overrides to cameras using
	 a different method.

	 There is one "schedule" method for invoking a system-wide schedule preset.
*/

// CameraMode is a set of constants to deal with three specific camera modes.
type CameraMode rune

// CameraMode* are used by the Camera scheduling methods. Use these constants
// as inputs to a Camera's schedule methods.
const (
	CameraModeAll        CameraMode = 'X'
	CameraModeMotion     CameraMode = 'M'
	CameraModeActions    CameraMode = 'A'
	CameraModeContinuous CameraMode = 'C'
)

// scheduleContainer allows unmarshalling of ScheduleOverrides and SchedulePresets into a map.
type ScheduleContainer map[int]string

// UnmarshalXML turns the XML schedule lists returned by SecuritySpy's API into a map[int]string.
func (m *ScheduleContainer) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// Loop each list element.
	for (*m) = make(ScheduleContainer); ; {
		var schedule struct {
			Name string `xml:"name"`
			ID   int    `xml:"id"`
		}

		token, err := d.Token()
		if err != nil {
			return fmt.Errorf("bad XML token: %w", err)
		}

		switch e := token.(type) {
		case xml.StartElement:
			if err = d.DecodeElement(&schedule, &e); err != nil {
				return fmt.Errorf("XML decode: %w", err)
			}

			(*m)[schedule.ID] = schedule.Name
		case xml.EndElement:
			if e == start.End() {
				return nil
			}
		}
	}
}

// SetSchedulePreset invokes a schedule preset. This [may/will] affect all camera arm modes.
// Find preset IDs you can pass into this method at server.Info.SchedulePresets.
func (s *Server) SetSchedulePreset(presetID int) error {
	params := make(url.Values)
	params.Set("id", strconv.Itoa(presetID))

	if err := s.SimpleReq("++ssSetPreset", params, -1); err != nil {
		return fmt.Errorf("http request: %w", err)
	}

	return nil
}
