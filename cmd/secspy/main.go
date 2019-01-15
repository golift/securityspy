package main

/* This is just a test app to demonstrate basic usage of the securityspy library. */

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	securityspy "github.com/davidnewhall/go-securityspy"
	flg "github.com/ogier/pflag"
)

// Version of the app
var Version = "1.0.0"

// Config from CLI
type Config struct {
	UseSSL bool
	User   string
	Pass   string
	URL    string
	Cmd    string
	Arg    string
	server securityspy.Server
}

func main() {
	config := parseFlags()
	securityspy.Encoder = "/usr/local/bin/ffmpg"
	switch config.Cmd {
	case "events", "event", "e":
		config.getServer()
		log.Println("Watching Event Stream")
		config.server.BindEvent(securityspy.EventAllEvents, config.showEvents)
		config.server.WatchEvents(10*time.Second, 4*time.Minute)
	case "cams", "cam", "c":
		config.printCamData()
	case "video", "vid", "v":
		config.saveVideo()
	case "picture", "pic", "p":
		config.savePicture()
	case "trigger", "t":
		config.triggerMotion()
	default:
		_, _ = fmt.Fprintln(os.Stderr, "invalid command:", config.Cmd)
		flg.Usage()
		os.Exit(1)
	}
}

func parseFlags() *Config {
	config := &Config{}
	flg.Usage = func() {
		fmt.Println("Usage: secspy [--user <user>] [--pass <pass>] [--url <url>] [-c <cmd>] [-a <arg>]")
		flg.PrintDefaults()
	}
	flg.StringVarP(&config.User, "user", "u", os.Getenv("SECSPY_USERNAME"), "Username to authenticate with")
	flg.StringVarP(&config.Pass, "pass", "p", os.Getenv("SECSPY_PASSWORD"), "Password to authenticate with")
	flg.StringVarP(&config.URL, "url", "U", "http://127.0.0.1:8000", "SecuritySpy URL")
	flg.BoolVarP(&config.UseSSL, "verify-ssl", "s", false, "Validate SSL certificate if using https")
	flg.StringVarP(&config.Cmd, "command", "c", "", "Command to run. Currently supports: events, cams, pic, vid, trigger")
	flg.StringVarP(&config.Arg, "arg", "a", "", "if cmd supports an argument, pass it here. ie. -c pic -a Porch:/tmp/filename.jpg")
	version := flg.BoolP("version", "v", false, "Print the version and exit")
	if flg.Parse(); *version {
		fmt.Println("secspy version:", Version)
		os.Exit(0) // don't run anything else.
	}
	return config
}

// getServer makes and returns a handle.
func (c *Config) getServer() securityspy.Server {
	var err error
	if c.server, err = securityspy.GetServer(c.User, c.Pass, c.URL, c.UseSSL); err != nil {
		log.Fatalln("SecuritySpy Error:", err)
	}
	fmt.Printf("%v %v (http://%v:%v/) %d cameras, %d scripts, %d sounds\n",
		c.server.Info().Name, c.server.Info().Version, c.server.Info().IP1,
		c.server.Info().HTTPPort, len(c.server.GetCameras()),
		len(c.server.Info().Scripts.Names), len(c.server.Info().Sounds.Names))
	return c.server
}

func (c *Config) triggerMotion() {
	if c.Arg == "" {
		fmt.Println("Triggers motion on a camera.")
		fmt.Println("Supply a camera name with -a <cam>")
		fmt.Println("Example: secspy -c trigger -a Door")
		fmt.Println("See camera names with -c cams")
		os.Exit(1)
	} else if cam := c.getServer().GetCameraByName(c.Arg); cam == nil {
		fmt.Println("Camera does not exist:", c.Arg)
		os.Exit(1)
	} else if err := cam.TriggerMotion(); err != nil {
		fmt.Printf("Error Trigger Motion for camera '%v': %v", c.Arg, err)
		os.Exit(1)
	}
	fmt.Println("Triggered Motion for Camera:", c.Arg)
}

func (c *Config) showEvents(e securityspy.Event) {
	camString := "No Camera"
	if e.Camera != nil {
		camString = "Camera " + e.Camera.Num() + ": " + e.Camera.Name()
	} else if e.ID < 0 {
		camString = "SecuritySpy Server"
	}
	fmt.Printf("[%v] Event %d: %v, %v, Msg: %v\n",
		e.When, e.ID, e.Event.Event(), camString, e.Msg)
}

func (c *Config) printCamData() {
	for _, cam := range c.getServer().GetCameras() {
		c := cam.Conf()
		fmt.Printf("%2v: %-10v (%-9v %5v/%-7v %v) connected: %3v, down %v, modes: C:%-8v M:%-8v A:%-8v "+
			"%2vFPS, Audio:%3v, MD: %3v/pre:%v/post:%3v idle %-10v Script: %v (reset %v)\n",
			cam.Num(), c.Name, cam.Size(), c.DeviceName, c.DeviceType, c.Address,
			c.Connected.Val, c.TimeSinceLastFrame.Dur.String(),
			c.ModeC.Txt, c.ModeM.Txt, c.ModeA.Txt+",", int(c.CurrentFPS), c.HasAudio.Txt, c.MDenabled.Txt, c.MDpreCapture.Dur.String(), c.MDpostCapture.Dur.String(),
			c.TimeSinceLastMotion.Dur.String(), c.ActionScriptName,
			c.ActionResettime.Dur.String())
	}
}

func (c *Config) savePicture() {
	if c.Arg == "" || !strings.Contains(c.Arg, ":") {
		fmt.Println("Saves a single still JPEG image from a camera.")
		fmt.Println("Supply a camera name and file path with -a <cam>:<path>")
		fmt.Println("Example: secspy -c pic -a Porch:/tmp/Porch.jpg")
		fmt.Println("See camera names with -c cams")
		os.Exit(1)
	}
	split := strings.Split(c.Arg, ":")
	cam := c.getServer().GetCameraByName(split[0])
	if cam == nil {
		fmt.Println("Camera does not exist:", split[0])
		os.Exit(1)
	} else if err := cam.SaveJPEG(&securityspy.VidOps{}, split[1]); err != nil {
		log.Fatalf("Error Saving Image for camera '%v' to file '%v': %v\n", cam.Name(), split[1], err)
	}
	fmt.Printf("Image for camera '%v' saved to: %v\n", cam.Name(), split[1])
}

func (c *Config) saveVideo() {
	if c.Arg == "" || !strings.Contains(c.Arg, ":") {
		fmt.Println("Saves a 10 second video from a camera.")
		fmt.Println("Supply a camera name and file path with -a <cam>:<path>")
		fmt.Println("Example: secspy -c pic -a Gate:/tmp/Gate.mov")
		fmt.Println("See camera names with -c cams")
		os.Exit(1)
	}
	split := strings.Split(c.Arg, ":")
	cam := c.getServer().GetCameraByName(split[0])
	if cam == nil {
		fmt.Println("Camera does not exist:", split[0])
		os.Exit(1)
	} else if err := cam.SaveVideo(&securityspy.VidOps{}, 10*time.Second, 9999999999, split[1]); err != nil {
		log.Fatalf("Error Saving Video for camera '%v' to file '%v': %v\n", cam.Name(), split[1], err)
	}
	fmt.Printf("10 Second video for camera '%v' saved to: %v\n", cam.Name(), split[1])
}
