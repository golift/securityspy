# go-securityspy

## OVERVIEW

Full Featured Go Library for [SecuritySpy](https://www.bensoftware.com/securityspy/)'s
web API. Read about the [API here](https://www.bensoftware.com/securityspy/web-server-spec.html).

Everything is reasonably tested and working. Feedback is welcomed! Lots of docs to come.

`ffmpeg` is used if you want video snippets, but not required for most functions.

My server doesn't have auth enabled, so I really have no idea if this works with a password. lemme know?

A command line interface app that uses this library exists. Most of the testing is done with this app.
Find it here: https://github.com/davidnewhall/SecSpyCLI - full of great examples on how to use this library.

- Requires [SecuritySpy 4.2.10b9](https://www.bensoftware.com/securityspy/download-beta.html) or later.
- There's a lot more to learn about this package in [GODOC](https://godoc.org/code.golift.io/securityspy).

## FEATURES

#### Server

 - All server and system Info is exposed with one API web request.
 - Schedule Presets can be retrieved and invoked.

#### Cameras

- Stream live H264 or MJPEG video from an `io.ReadCloser`.
- Stream live G711 audio from an `io.ReadCloser`.
- Submit G711 audio (files or microphone) to a camera from an `io.ReadCloser`.
- Save live video snippets locally (requires `FFMPEG`).
- Get live JPEG images in `image` format, or save files locally.
- Arm and Disarm actions, motion capture and continuous capture.
- Trigger Motion.
- Set schedules and schedule overrides.
- Inspect PTZ capabilities.
- Control all PTZ actions including invoking and saving presets.

#### Events

SecuritySpy has a handy event stream; you can bind functions and/or channels to
all or specific events. When a bound event fires the callback method it's bound
to is run. In the case of a channel binding, the event is sent to the channel
fo consumption by a worker (pool).

- Exposes all 11 SecuritySpy events.
- Exposes 6 custom events.
- Method to inject custom events into the event stream.

#### Files

SecuritySpy saves video and image files based on motion and continuous capture
settings. These files can be listed and downloaded with this library.

- List and retrieve captured images.
- List and retrieve continuous captured videos.
- List and retrieve motion captured videos.
- Save files locally or stream from `io.ReadCloser`.


## EXAMPLE

This example shows some of the data that is provided by the API. None of the
actions methods are invoked here. See the [SecSpyCLI](https://github.com/davidnewhall/SecSpyCLI/blob/master/cmd/secspy/main.go)
app for examples of other methods.

```golang
package main

import (
	"fmt"
	"code.golift.io/securityspy"
)

func main() {
	server, err := securityspy.GetServer("admin", "password", "http://127.0.0.1:8000", false)
	if err != nil {
		panic(err)
	}
	scripts, _ := server.GetScripts()
	sounds, _ := server.GetSounds()

	// Print server info.
	fmt.Printf("%v %v @ %v (http://%v:%v/) %d cameras, %d scripts, %d sounds, %d schedules, %d schedule presets\n",
		server.Info.Name, server.Info.Version, server.Info.CurrentTime,
		server.Info.IP1, server.Info.HTTPPort, len(server.Cameras.Names),
		len(scripts), len(sounds), len(server.Info.ServerSchedules), len(server.Info.SchedulePresets))

	// Print info for each camera.
	for _, camera := range server.Cameras.All() {
		fmt.Printf("%2v: %-14v (%-4vx%-4v %5v/%-7v %v) connected: %3v, down %v, modes: C:%-8v M:%-8v A:%-8v "+
			"%2vFPS, Audio:%3v, MD: %3v/pre:%v/post:%3v idle %-10v Script: %v (reset %v)\n",
			camera.Number, camera.Name, camera.Width, camera.Height, camera.DeviceName, camera.DeviceType, camera.Address,
			camera.Connected.Val, camera.TimeSinceLastFrame.String(), camera.ModeC.Txt, camera.ModeM.Txt,
			camera.ModeA.Txt+",", int(camera.CurrentFPS), camera.HasAudio.Txt, camera.MDenabled.Txt,
			camera.MDpreCapture.String(), camera.MDpostCapture.String(),
			camera.TimeSinceLastMotion.String(), camera.ActionScriptName, camera.ActionResetTime.String())
	}
}
```
The output looks like this:
```
SecuritySpy 4.2.10b9 @ 2019-02-09 16:20:00 -0700 MST (http://192.168.1.1:8000/) 7 cameras, 18 scripts, 20 sounds, 6 schedules, 1 schedule presets
 0: Porch          (2304x1296 ONVIF/Network 192.168.1.12) connected: true, down 0s, modes: C:armed    M:armed    A:armed,   20FPS, Audio:yes, MD: yes/pre:3s/post:10s idle 3h5m5s     Script: SS_SendiMessages.scpt (reset 1m0s)
 1: Door           (2592x1520 ONVIF/Network 192.168.1.13) connected: true, down 0s, modes: C:armed    M:armed    A:armed,   15FPS, Audio:yes, MD: yes/pre:4s/post: 5s idle 9m24s      Script: SS_SendiMessages.scpt (reset 1m0s)
 2: Road           (2592x1520 ONVIF/Network 192.168.1.11) connected: true, down 0s, modes: C:armed    M:armed    A:disarmed, 20FPS, Audio: no, MD: yes/pre:3s/post: 5s idle 4m35s      Script: SS_SendiMessages.scpt (reset 59s)
 3: Garage         (3072x2048 ONVIF/Network 192.168.1.14) connected: true, down 0s, modes: C:armed    M:armed    A:armed,   20FPS, Audio:yes, MD: yes/pre:3s/post: 5s idle -1ns       Script: SS_SendiMessages.scpt (reset 1m0s)
 4: Gate           (2560x1440 ONVIF/Network 192.168.1.16) connected: true, down 0s, modes: C:armed    M:armed    A:armed,   29FPS, Audio:yes, MD: yes/pre:3s/post: 5s idle -1ns       Script: SS_SendiMessages.scpt (reset 1m0s)
 5: Pool           (3072x2048 ONVIF/Network 192.168.1.17) connected: true, down 0s, modes: C:armed    M:armed    A:armed,   10FPS, Audio:yes, MD: yes/pre:2s/post:20s idle 16m18s     Script: SS_SendiMessages.scpt (reset 1m0s)
 6: Car            (2048x1536 ONVIF/Network 192.168.1.18) connected: true, down 0s, modes: C:armed    M:armed    A:armed,   20FPS, Audio: no, MD: yes/pre:0s/post:15s idle -1ns       Script: SS_SendiMessages.scpt (reset 48s)
 ```

## LICENSE

[MIT License](LICENSE) - Copyright (c) 2019 David Newhall II
