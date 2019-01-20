package securityspy

import (
	"encoding/xml"
	"io"
	"time"
)

// fileDateFormat is the format the SecuritySpy ++download method accepts.
var fileDateFormat = "01/02/06"

// Errors returned by this file.
const (
	ErrorPathExists  = Error("cannot overwrite existing path")
	ErrorNoExtension = Error("missing file extension")
	ErrorInvalidName = Error("invalid file name")
)

// Files interface allows searching for saved files.
type Files interface {
	GetAll(cameraNums []int, from, to time.Time) ([]*File, error)
	GetPhotos(cameraNums []int, from, to time.Time) ([]*File, error)
	GetVideos(cameraNums []int, from, to time.Time) ([]*File, error)
	GetFile(name string) (file *File, err error)
}

// files powers the Files interface.
// It's really an extension of the Server interface.
type files struct {
	*Server
}

// fileFeed represents the XML data from ++download
type fileFeed struct {
	XMLName      xml.Name `xml:"feed"`
	BSL          string   `xml:"bsl,attr"`     // http://www.bensoftware.com/
	Title        string   `xml:"title"`        // Downloads
	GmtOffset    string   `xml:"gmt-offset"`   // -28800
	Continuation string   `xml:"continuation"` // 0007E3010C0E1D3A
	Entries      []File   `xml:"entry"`
}

// File represents a saved media file.
type File struct {
	Title     string    `xml:"title"` // 01-12-2019 M Gate.m4v, 01...
	Link      LinkInfo  `xml:"link"`
	Updated   time.Time `xml:"updated"`   // 2019-01-12T08:57:58Z, 201...
	CameraNum int       `xml:"cameraNum"` // 0, 1, 2, 4, 5, 7, 9, 10, 11, 12, 13
	GmtOffset int
	Camera    Camera
	server    *Server
	fileEntry
}

// LinkInfo is part of filesEntry
type LinkInfo struct {
	Rel    string `xml:"rel,attr"`    // alternate, alternate, alternate
	Type   string `xml:"type,attr"`   // video/quicktime, video/quicktime
	Length int64  `xml:"length,attr"` // 358472320, 483306152, 900789978,
	HREF   string `xml:"href,attr"`   // ++getfile/4/2018-10-17/10-17-2018+M+Gate.m4v
}

// fileEntry is the interface to do something with a File.
type fileEntry interface {
	Save(path string) (size int64, err error)
	Get() (io.ReadCloser, error)
}
