package ffmpegencode

/* Encode videos from RTSP URLs using FFMPEG */

import (
	"bytes"
	"io"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Default, Maximum and Minimum Values
var (
	DefaultFrameRate   = 5
	MinimumFrameRate   = 1
	MaximumFrameRate   = 60
	DefaultFrameHeight = 720
	DefaultFrameWidth  = 1280
	MinimumFrameSize   = 100
	MaximumFrameSize   = 5000
	DefaultEncodeCRF   = 21
	MinimumEncodeCRF   = 16
	MaximumEncodeCRF   = 30
	DefaultCaptureTime = 15
	MaximumCaptureTime = 1200             // 10 minute max.
	DefaultCaptureSize = int64(2500000)   // 2.5MB default (roughly 5-10 seconds)
	MaximumCaptureSize = int64(104857600) // 100MB max.
	DefaultFFmpegPath  = "/usr/local/bin/ffmpeg"
	DefaultProfile     = "main"
	DefaultLevel       = "3.0"
)

// These are custom errors this library may return.
const (
	ErrorInvalidOutput = Error("output path is not valid")
	ErrorInvalidInput  = Error("input path is not valid")
)

// Error enables constant errors.
type Error string

// Error allows a string to satisfy the error type.
func (e Error) Error() string {
	return string(e)
}

// Encoder provides an inteface to mock.
type Encoder interface {
	Values() *VidOps
	GetVideo(input, title string) (cmd string, cmdout io.ReadCloser, err error)
	SaveVideo(input, output, title string) (cmd string, out string, err error)
	SetAudio(audio string) (value bool)
	SetRate(rate string) (value int)
	SetLevel(level string) (value string)
	SetWidth(width string) (value int)
	SetHeight(height string) (value int)
	SetCRF(crf string) (value int)
	SetTime(seconds string) (value int)
	SetSize(size string) (value int64)
	SetProfile(profile string) (value string)
}

// EncodeInterface allows us to porovide an interface with sub-type'd data.
type EncodeInterface struct {
	ops *VidOps
}

// VidOps defines how to ffmpeg shall transcode a stream.
type VidOps struct {
	Encoder string // "/usr/local/bin/ffmpeg"
	Level   string // 3.0, 3.1 ..
	Width   int    // 1920
	Height  int    // 1080
	CRF     int    // 24
	Time    int    // 15 (seconds)
	Audio   bool   // include audio?
	Rate    int    // framerate (5-20)
	Size    int64  // max file size (always goes over). use 2000000 for 2.5MB
	Prof    string // main, high, baseline
	Copy    bool   // Copy original stream, rather than transcode.
}

// Get an encoder interface.
func Get(v *VidOps) Encoder {
	encoder := &EncodeInterface{ops: v}
	if encoder.ops.Encoder == "" {
		encoder.ops.Encoder = DefaultFFmpegPath
	}
	encoder.SetLevel(v.Level)
	encoder.SetProfile(v.Prof)
	encoder.fixValues()
	return encoder
}

// Values returns the current values in the encoder.
func (v *EncodeInterface) Values() *VidOps {
	// This pointer can be manipulated directly by the receiving function.
	return v.ops
}

// SetAudio turns audio on or off based on a string value.
func (v *EncodeInterface) SetAudio(audio string) bool {
	v.ops.Audio, _ = strconv.ParseBool(audio)
	return v.ops.Audio
}

// SetLevel sets the h264 transcode level.
func (v *EncodeInterface) SetLevel(level string) string {
	if v.ops.Level = level; level != "3.0" && level != "3.1" && level != "4.0" && level != "4.1" && level != "4.2" {
		v.ops.Level = DefaultLevel
	}
	return v.ops.Level
}

// SetProfile sets the h264 transcode profile.
func (v *EncodeInterface) SetProfile(profile string) string {
	if v.ops.Prof = profile; v.ops.Prof != "main" && v.ops.Prof != "baseline" && v.ops.Prof != "high" {
		v.ops.Prof = DefaultProfile
	}
	return v.ops.Prof
}

// SetWidth sets the transcode frame width.
func (v *EncodeInterface) SetWidth(width string) int {
	v.ops.Width, _ = strconv.Atoi(width)
	v.fixValues()
	return v.ops.Width
}

// SetHeight sets the transcode frame width.
func (v *EncodeInterface) SetHeight(height string) int {
	v.ops.Height, _ = strconv.Atoi(height)
	v.fixValues()
	return v.ops.Height
}

// SetCRF sets the h264 transcode CRF value.
func (v *EncodeInterface) SetCRF(crf string) int {
	v.ops.CRF, _ = strconv.Atoi(crf)
	v.fixValues()
	return v.ops.CRF
}

// SetTime sets the maximum transcode duration.
func (v *EncodeInterface) SetTime(seconds string) int {
	v.ops.Time, _ = strconv.Atoi(seconds)
	v.fixValues()
	return v.ops.Time
}

// SetRate sets the transcode framerate.
func (v *EncodeInterface) SetRate(rate string) int {
	v.ops.Rate, _ = strconv.Atoi(rate)
	v.fixValues()
	return v.ops.Rate
}

// SetSize sets the maximum transcode file size.
func (v *EncodeInterface) SetSize(size string) int64 {
	v.ops.Size, _ = strconv.ParseInt(size, 10, 64)
	v.fixValues()
	return v.ops.Size
}

func (v *EncodeInterface) getVideoHandle(input, output, title string) (string, *exec.Cmd) {
	if title == "" {
		title = filepath.Base(output)
	}
	arg := []string{
		v.ops.Encoder,
		"-v", "16", // log level
		"-rtsp_transport", "tcp",
		"-i", input,
		"-f", "mov",
		"-metadata", `title="` + title + `"`,
		"-y", "-map", "0",
	}
	if v.ops.Size > 0 {
		arg = append(arg, "-fs", strconv.FormatInt(v.ops.Size, 10))
	}
	if v.ops.Time > 0 {
		arg = append(arg, "-t", strconv.Itoa(v.ops.Time))
	}
	if !v.ops.Copy {
		arg = append(arg, "-vcodec", "libx264",
			"-profile:v", v.ops.Prof,
			"-level", v.ops.Level,
			"-pix_fmt", "yuv420p",
			"-movflags", "faststart",
			"-s", strconv.Itoa(v.ops.Width)+"x"+strconv.Itoa(v.ops.Height),
			"-preset", "superfast",
			"-crf", strconv.Itoa(v.ops.CRF),
			"-r", strconv.Itoa(v.ops.Rate),
		)
	} else {
		arg = append(arg, "-c", "copy")
	}
	if !v.ops.Audio {
		arg = append(arg, "-an")
	} else {
		arg = append(arg, "-c:a", "copy")
	}
	arg = append(arg, output)
	cmd := exec.Command(arg[0], arg[1:]...)

	return strings.Join(arg, " "), cmd
}

// GetVideo retreives video from an input and returns a Reader to consume the output.
// The Reader contains output messages if output is a filepath.
// The Reader contains the video if the output is "-"
func (v *EncodeInterface) GetVideo(input, title string) (string, io.ReadCloser, error) {
	if input == "" {
		return "", nil, ErrorInvalidInput
	}
	cmdStr, cmd := v.getVideoHandle(input, "-", title)
	stdoutpipe, err := cmd.StdoutPipe()
	if err != nil {
		return cmdStr, nil, err
	}
	return cmdStr, stdoutpipe, cmd.Run()
}

// SaveVideo saves a video snippet to a file.
func (v *EncodeInterface) SaveVideo(input, output, title string) (string, string, error) {
	if input == "" {
		return "", "", ErrorInvalidInput
	} else if output == "" || output == "-" {
		return "", "", ErrorInvalidOutput
	}
	cmdStr, cmd := v.getVideoHandle(input, output, title)
	// log.Println(cmdStr) // DEBUG
	var out bytes.Buffer
	cmd.Stderr = &out
	cmd.Stdout = &out
	if err := cmd.Start(); err != nil {
		return cmdStr, "", err
	}
	err := cmd.Wait()
	return cmdStr, strings.TrimSpace(out.String()), err
}

// fixValues makes sure video request values are sane.
func (v *EncodeInterface) fixValues() {
	if v.ops.Height == 0 {
		v.ops.Height = DefaultFrameHeight
	} else if v.ops.Height > MaximumFrameSize {
		v.ops.Height = MaximumFrameSize
	} else if v.ops.Height < MinimumFrameSize {
		v.ops.Height = MinimumFrameSize
	}

	if v.ops.Width == 0 {
		v.ops.Width = DefaultFrameWidth
	} else if v.ops.Width > MaximumFrameSize {
		v.ops.Width = MaximumFrameSize
	} else if v.ops.Width < MinimumFrameSize {
		v.ops.Width = MinimumFrameSize
	}

	if v.ops.CRF == 0 {
		v.ops.CRF = DefaultEncodeCRF
	} else if v.ops.CRF < MinimumEncodeCRF {
		v.ops.CRF = MinimumEncodeCRF
	} else if v.ops.CRF > MaximumEncodeCRF {
		v.ops.CRF = MaximumEncodeCRF
	}

	if v.ops.Rate == 0 {
		v.ops.Rate = DefaultFrameRate
	} else if v.ops.Rate < MinimumFrameRate {
		v.ops.Rate = MinimumFrameRate
	} else if v.ops.Rate > MaximumFrameRate {
		v.ops.Rate = MaximumFrameRate
	}

	// No minimums.
	if v.ops.Time == 0 {
		v.ops.Time = DefaultCaptureTime
	} else if v.ops.Time > MaximumCaptureTime {
		v.ops.Time = MaximumCaptureTime
	}

	if v.ops.Size == 0 {
		v.ops.Size = DefaultCaptureSize
	} else if v.ops.Size > MaximumCaptureSize {
		v.ops.Size = MaximumCaptureSize
	}
}
