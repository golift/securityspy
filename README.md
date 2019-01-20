# go-securityspy

Full Featured Go Library for [SecuritySpy](https://www.bensoftware.com/securityspy/)

Feedback is welcomed! A versioned release will be created soon. I'd like to finish testing PTZ before a release.

`Events` and, `Camera` and `Cameras` interfaces are (mostly) tested and working.
`Files` and `File` interfaces are well tested and working.
`ptz` interface has received little to no testing, but probably works? :D

Still working on it. Lots of docs to come.

`ffmpeg` is also used if you want video snippets, but not required for most functions.

My server doesn't have auth enabled, so I really have no idea if this works with a password. lemme know?

A command line interface app that uses this library exists. Most of the testing is done with this app.
Find it here: https://github.com/davidnewhall/SecSpyCLI
