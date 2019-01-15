package securityspy

/* This file handles all file transfers for media saved by security spy. */
// This file and the methods herein: incomplete

import (
	"encoding/xml"
	"io"
	"time"
)

// Files interface allows searching for saved files.
type Files interface {
	GetSavePhotos([]int, time.Time, time.Time) []FileEntry
	GetSaveMedias([]int, time.Time, time.Time) []FileEntry
	GetSaveVideos([]int, time.Time, time.Time) []FileEntry
}

// FilesData powers the Files interface.
type FilesData struct {
	config *Config
}

// Feed represents the XML data from /++download
type Feed struct {
	XMLName      xml.Name    `xml:"feed"`
	BSL          string      `xml:"bsl,attr"`     // http://www.bensoftware.com/
	Title        string      `xml:"title"`        // Downloads
	GmtOffset    string      `xml:"gmt-offset"`   // -28800
	Continuation string      `xml:"continuation"` // 0007E3010C0E1D3A
	Entry        []FileEntry `xml:"entry"`
}

// FileEntry represents a saved media file.
type FileEntry struct {
	Title string `xml:"title"` // 01-12-2019 M Gate.m4v, 01...
	Link  struct {
		Rel    string  `xml:"rel,attr"`    // alternate, alternate, alternate
		Type   string  `xml:"type,attr"`   // video/quicktime, video/quicktime
		Length float64 `xml:"length,attr"` // 358472320, 483306152, 900789978,
		Href   string  `xml:"href,attr"`   // ++getfile/4/2018-10-17/10-17-2018+M+Gate.m4v
	} `xml:"link"`
	Updated   time.Time `xml:"updated"`   // 2019-01-12T08:57:58Z, 201...
	CameraNum int       `xml:"cameraNum"` // 0, 1, 2, 4, 5, 7, 9, 10, 11, 12, 13
	config    *Config
}

// FileInterface is used to do something with a FileEntry.
type FileInterface interface {
	StreamFile(file string) (io.Reader, error)
	SaveFile(file, path string) error
}

// Files returns a Files interface, used to retreive file listings.
func (c *concourse) Files() Files {
	return FilesData{config: c.Config}
}

// GetSavePhotos returns a list of links to captured images.
func (f FilesData) GetSavePhotos([]int, time.Time, time.Time) []FileEntry {
	return nil
}

// GetSaveMedias returns a list of links to captured videos and images.
func (f FilesData) GetSaveMedias([]int, time.Time, time.Time) []FileEntry {
	return nil
}

// GetSaveVideos returns a list of links to captured videos.
func (f FilesData) GetSaveVideos([]int, time.Time, time.Time) []FileEntry {
	return nil
}
