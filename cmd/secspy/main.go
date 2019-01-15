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
}

func main() {
	config := parseFlags()
	securityspy.Encoder = "/usr/local/bin/ffmpg"

	switch config.Cmd {

	case "events", "event", "e":
		spy := config.getHandle()
		log.Println("Watching Event Stream")
		spy.BindEvent(securityspy.EventAllEvents, showEvents)
		spy.WatchEvents(5 * time.Second)

	case "cams", "cam", "c":
		spy := config.getHandle()
		printCamData(spy.Cameras())

	case "video", "vid", "v":
		if config.Arg == "" || !strings.Contains(config.Arg, ":") {
			fmt.Println("Saves a 10 second video from a camera.")
			fmt.Println("Supply a camera name and file path with -a <cam>:<path>")
			fmt.Println("Example: secspy -c pic -a Gate:/tmp/Gate.mov")
			fmt.Println("See camera names with -c cams")
			os.Exit(1)
		}
		split := strings.Split(config.Arg, ":")
		cam := config.getHandle().CameraByName(split[0])
		if cam == nil {
			fmt.Println("Camera does not exist:", split[0])
			os.Exit(1)
		}
		saveVideo(cam, split[1], 10)

	case "picture", "pic", "p":
		if config.Arg == "" || !strings.Contains(config.Arg, ":") {
			fmt.Println("Saves a single still JPEG image from a camera.")
			fmt.Println("Supply a camera name and file path with -a <cam>:<path>")
			fmt.Println("Example: secspy -c pic -a Gate:/tmp/Gate.jpg")
			fmt.Println("See camera names with -c cams")
			os.Exit(1)
		}
		split := strings.Split(config.Arg, ":")
		cam := config.getHandle().CameraByName(split[0])
		if cam == nil {
			fmt.Println("Camera does not exist:", split[0])
			os.Exit(1)
		}
		savePicture(cam, split[1])

	default:
		// We have no valid cmd, so exit. Add more actions here as more are created.
		_, _ = fmt.Fprintln(os.Stderr, "invalid command:", config.Cmd)
		flg.Usage()
		os.Exit(1)
	}
}

func (c *Config) getHandle() securityspy.SecuritySpy {
	spy, err := securityspy.Handle(c.User, c.Pass, c.URL, c.UseSSL)
	if err != nil {
		log.Fatalln("Handle Error:", err)
	}
	fmt.Println("Server:", spy.ServerInfo().Name, "("+spy.ServerInfo().IP1+")")
	return spy
}

func parseFlags() *Config {
	config := &Config{}
	flg.Usage = func() {
		fmt.Println("Usage: secspy [--user <user>] [--pass <pass>] [--url <url>] [--ssl] [-c <cmd>] [-a <arg>] [-v]")
		flg.PrintDefaults()
	}
	flg.StringVarP(&config.User, "user", "u", "", "Username to authenticate with")
	flg.StringVarP(&config.Pass, "pass", "p", "", "Password to authenticate with")
	flg.StringVarP(&config.URL, "url", "U", "http://127.0.0.1:8000", "SecuritySpy URL")
	flg.BoolVarP(&config.UseSSL, "use-ssl", "s", false, "Validate SSL certificate if using https")
	flg.StringVarP(&config.Cmd, "command", "c", "", "Command to run. Currently supports: events, cams, pic, vid")
	flg.StringVarP(&config.Arg, "arg", "a", "", "if cmd supports an argument, pass it here. ie. -c pic -a Porch:/tmp/filename.jpg")
	version := flg.BoolP("version", "v", false, "Print the version and exit")
	if flg.Parse(); *version {
		fmt.Println("secspy version:", Version)
		os.Exit(0) // don't run anything else.
	}
	return config
}

func showEvents(e securityspy.Event) {
	fmt.Printf("[%v] Event %d: %v, Camera %v, Raw: %v\n", e.When, e.ID, e.Name, e.Camera, strings.Join(strings.Split(e.Raw, " ")[3:], " "))
}

func printCamData(cams []securityspy.Camera) {
	for _, cam := range cams {
		c := cam.Conf()
		fmt.Printf("%d: %v (%dx%d, %v/%v: %v) connected: %v, modes: C:%v M:%v, A:%v, FPS: %v, Audio? %v, Script: %v (reset %vs), MD: %v/pre:%vs/post:%vs\n",
			c.Number, c.Name, c.Width, c.Height, c.DeviceName, c.DeviceType, c.Address, c.Connected, c.ModeC, c.ModeM, c.ModeA,
			c.CurrentFPS, c.HasAudio, c.ActionScriptName, c.ActionResettime, c.MDenabled, c.MDpreCapture, c.MDpostCapture)
	}
}

func savePicture(cam securityspy.Camera, path string) {
	if err := cam.SaveJPEG(&securityspy.VidOps{}, path); err != nil {
		log.Fatalf("Error Saving Image for camera '%v' to file '%v': %v\n", cam.Name(), path, err)
	}
	fmt.Printf("Image for camera '%v' saved to: %v\n", cam.Name(), path)
}

func saveVideo(cam securityspy.Camera, path string, length int) {
	if err := cam.SaveVideo(&securityspy.VidOps{}, time.Duration(length)*time.Second, 9999999999, path); err != nil {
		log.Fatalf("Error Saving Video for camera '%v' to file '%v': %v\n", cam.Name(), path, err)
	}
	fmt.Printf("10 Second video for camera '%v' saved to: %v\n", cam.Name(), path)
}
