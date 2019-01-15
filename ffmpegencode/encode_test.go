package ffmpegencode

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TODO: test v.Copy, test GetVideo (better)

func TestFixValues(t *testing.T) {
	a := assert.New(t)
	v := Get(&VidOps{})
	// Test default values.
	a.False(v.SetAudio(""), "Wrong default 'audio' value!")
	a.EqualValues(DefaultProfile, v.SetProfile(""), "Wrong default 'profile' value!")
	a.EqualValues(DefaultLevel, v.SetLevel(""), "Wrong default 'level' value!")
	a.EqualValues(DefaultFrameHeight, v.SetHeight(""), "Wrong default 'height' value!")
	a.EqualValues(DefaultFrameWidth, v.SetWidth(""), "Wrong default 'width' value!")
	a.EqualValues(DefaultEncodeCRF, v.SetCRF(""), "Wrong default 'crf' value!")
	a.EqualValues(DefaultCaptureTime, v.SetTime(""), "Wrong default 'time' value!")
	a.EqualValues(DefaultFrameRate, v.SetRate(""), "Wrong default 'rate' value!")
	a.EqualValues(int64(DefaultCaptureSize), v.SetSize(""), "Wrong default 'size' value!")
	// Text max values.
	a.EqualValues(MaximumFrameSize, v.SetHeight("9000"), "Wrong maximum 'height' value!")
	a.EqualValues(MaximumFrameSize, v.SetWidth("9000"), "Wrong maximum 'width' value!")
	a.EqualValues(MaximumEncodeCRF, v.SetCRF("9000"), "Wrong maximum 'crf' value!")
	a.EqualValues(MaximumCaptureTime, v.SetTime("9000"), "Wrong maximum 'time' value!")
	a.EqualValues(MaximumFrameRate, v.SetRate("9000"), "Wrong maximum 'rate' value!")
	a.EqualValues(int64(MaximumCaptureSize), v.SetSize("999999999"), "Wrong maximum 'size' value!")
	// Text min values.
	a.EqualValues(MinimumFrameSize, v.SetHeight("1"), "Wrong minimum 'height' value!")
	a.EqualValues(MinimumFrameSize, v.SetWidth("1"), "Wrong minimum 'width' value!")
	a.EqualValues(MinimumEncodeCRF, v.SetCRF("1"), "Wrong minimum 'CRF' value!")
	a.EqualValues(MinimumFrameRate, v.SetRate("1"), "Wrong minimum 'rate' value!")
}

func TestSaveVideo(t *testing.T) {
	a := assert.New(t)
	v := Get(&VidOps{Encoder: "/bin/echo"})
	fileTemp := "/tmp/go-securityspy-encode-test-12345.txt"
	cmd, out, err := v.SaveVideo("INPUT", fileTemp, "TITLE")
	a.Nil(err, "/bin/echo returned an error. Something may be wrong with your environment.")
	// Make sure the produced command has all the expected values.
	a.Contains(cmd, "-an", "Audio may not be correctly disabled.")
	a.Contains(cmd, "-rtsp_transport tcp -i INPUT", "INPUT value appears to be missing, or rtsp transport is out of order")
	a.Contains(cmd, "-metadata title=\"TITLE\"", "TITLE value appears to be missing.")
	a.Contains(cmd, fmt.Sprintf("-vcodec libx264 -profile:v %v -level %v", DefaultProfile, DefaultLevel), "Level or Profile are missing or out of order.")
	a.Contains(cmd, fmt.Sprintf("-crf %d", DefaultEncodeCRF), "CRF value is missing or malformed.")
	a.Contains(cmd, fmt.Sprintf("-t %d", DefaultCaptureTime), "Capture Time value is missing or malformed.")
	a.Contains(cmd, fmt.Sprintf("-s %dx%d", DefaultFrameWidth, DefaultFrameHeight), "Framesize is missing or malformed.")
	a.Contains(cmd, fmt.Sprintf("-r %d", DefaultFrameRate), "Frame Rate value is missing or malformed.")
	a.Contains(cmd, fmt.Sprintf("-fs %d", DefaultCaptureSize), "Size value is missing or malformed.")
	a.True(strings.HasPrefix(cmd, "/bin/echo"), "The command does not - but should - begin with the Encoder value.")
	a.True(strings.HasSuffix(cmd, fileTemp), "The command does not - but should - end with a dash to indicate output to stdout.")
	a.EqualValues(cmd+"\n", "/bin/echo "+out, "Somehow the wrong value was written")
	// Make sure audio can be turned on.
	v = Get(&VidOps{Encoder: "/bin/echo", Audio: true})
	cmd, _, err = v.GetVideo("INPUT", "OUTPUT", "TITLE")
	a.Nil(err, "/bin/echo returned an error. Something may be wrong with your environment.")
	a.Contains(cmd, "-c:a copy", "Audio may not be correctly enabled.")
}

func TestValues(t *testing.T) {
	a := assert.New(t)
	v := Get(&VidOps{})
	values := v.Values()
	a.EqualValues(DefaultFFmpegPath, values.Encoder)
	a.EqualValues(DefaultFrameRate, values.Rate)
	a.EqualValues(DefaultFrameHeight, values.Height)
	a.EqualValues(DefaultFrameWidth, values.Width)
	a.EqualValues(DefaultEncodeCRF, values.CRF)
	a.EqualValues(DefaultCaptureTime, values.Time)
	a.EqualValues(DefaultCaptureSize, values.Size)
	a.EqualValues(DefaultProfile, values.Prof)
	a.EqualValues(DefaultLevel, values.Level)
}

func TestError(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	sentence := "this is an error string"
	err := Error(sentence)
	a.EqualValues(sentence, err.Error())
}
