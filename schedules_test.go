package securityspy_test

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golift.io/securityspy/v2"
)

func TestSetSchedulePreset(t *testing.T) {
	t.Parallel()

	secspyServer, recorder, _ := testServerWithCamera(t)

	require.NoError(t, secspyServer.SetSchedulePreset(1), "this method must not return an error during testing")

	req, ok := recorder.findLast("/++ssSetPreset")
	require.True(t, ok)
	require.Equal(t, "1", req.Query.Get("id"))
}

func TestUnmarshalXMLscheduleContainer(t *testing.T) {
	t.Parallel()

	asert := assert.New(t)

	var schedule securityspy.ScheduleContainer

	err := xml.Unmarshal([]byte(testScheduleList), &schedule)
	require.NoError(t, err, "valid data must not produce an error")
	asert.Equal("Armed 24/7", schedule[1], "the scheduleContainer data did not unmarshal properly")
	asert.Len(schedule, 6)

	err = xml.Unmarshal([]byte("<gotrekt>"), &schedule)
	require.Error(t, err, "invalid data must produce an error")

	err = xml.Unmarshal([]byte("<gotrekt><server></gotrekt>"), &schedule)
	require.Error(t, err, "invalid data must produce an error")
}

/**/
