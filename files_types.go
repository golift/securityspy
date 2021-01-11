package securityspy

import (
	"encoding/xml"
	"fmt"
	"time"
)

const (
	// DownloadDateFormat is the format the SecuritySpy ++download method accepts.
	// This matches the ++download inputs AND the folder names files are saved into.
	// The file1/file2 inputs this gets passed into are actually undocuemnted and were
	// created specifically for programmtic SDK access (ie. this library).
	DownloadDateFormat = "2006-01-02"
	// FileDateFormat is an arbitrary date format used for saved files; we hope doesn't change.
	// This is used in the actual name of files that are saved. No where else.
	// The GetFile() method uses this to construct arbitrary file download paths.
	FileDateFormat = "01-02-2006"
)

// Errors returned by the Files type methods.
var (
	// ErrorPathExists returns when a requested write path already exists.
	ErrorPathExists = fmt.Errorf("cannot overwrite existing path")

	// ErrorInvalidName returns when requesting a file download and the filename is invalid.
	ErrorInvalidName = fmt.Errorf("invalid file name")
)

// Files powers the Files interface.
// Use the bound methods to list and download saved media files.
type Files struct {
	server *Server
}

// fileFeed represents the XML data from ++download api path.
type fileFeed struct {
	XMLName      xml.Name `xml:"feed"`
	BSL          string   `xml:"bsl,attr"`     // http://www.bensoftware.com/
	Title        string   `xml:"title"`        // Downloads
	GmtOffset    Duration `xml:"gmt-offset"`   // -28800
	Continuation string   `xml:"continuation"` // 0007E3010C0E1D3A
	Entries      []*File  `xml:"entry"`        // List of File pointers
}

// File represents a saved media file. This is all the data retreived from
// the ++download method for a particular file. Contains a camera interface
// for the camera that created the file. All of the Files type methods return this type.
type File struct {
	Title string `xml:"title"` // 01-12-2019 M Gate.m4v, 01...
	Link  struct {
		Rel    string `xml:"rel,attr"`    // alternate, alternate, alternate
		Type   string `xml:"type,attr"`   // video/quicktime, video/quicktime
		Length int64  `xml:"length,attr"` // 358472320, 483306152, 900789978,
		HREF   string `xml:"href,attr"`   // ++getfile/4/2018-10-17/10-17-2018+M+Gate.m4v
	} `xml:"link"`
	Updated   time.Time     `xml:"updated"`   // 2019-01-12T08:57:58Z, 201...
	CameraNum int           `xml:"cameraNum"` // 0, 1, 2, 4, 5, 7, 9, 10, 11, 12, 13
	GmtOffset time.Duration // the rest are copied in per-file from fileFeed.
	Camera    *Camera
	server    *Server
}
