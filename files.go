package securityspy

/* This file handles all file transfers for media saved by security spy. */
// This file and the methods herein: incomplete

import (
	"encoding/xml"
	"io"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

// fileDateFormat is the format the SecuritySpy ++downloads method accepts.
var fileDateFormat = "02/01/06"

// Files interface allows searching for saved files.
type Files interface {
	GetAll(cameraNums []int, from, to time.Time) ([]FileInterface, error)
	GetPhotos(cameraNums []int, from, to time.Time) ([]FileInterface, error)
	GetVideos(cameraNums []int, from, to time.Time) ([]FileInterface, error)
}

// FilesData powers the Files interface.
// It's really an extension of the concourse interface.
type FilesData struct {
	Server *concourse
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
		Rel    string `xml:"rel,attr"`    // alternate, alternate, alternate
		Type   string `xml:"type,attr"`   // video/quicktime, video/quicktime
		Length int64  `xml:"length,attr"` // 358472320, 483306152, 900789978,
		HREF   string `xml:"href,attr"`   // ++getfile/4/2018-10-17/10-17-2018+M+Gate.m4v
	} `xml:"link"`
	Updated   time.Time `xml:"updated"`   // 2019-01-12T08:57:58Z, 201...
	CameraNum int       `xml:"cameraNum"` // 0, 1, 2, 4, 5, 7, 9, 10, 11, 12, 13
	server    *concourse
	camera    Camera
}

// FileInterface is used to do something with a FileEntry.
type FileInterface interface {
	Name() string
	Size() int64
	Type() string
	Date() time.Time
	Save(path string) error
	Camera() Camera
	Stream() (io.ReadCloser, error)
}

/* Files-specific concourse methods are at the top. */

// Files returns a Files interface, used to retreive file listings.
func (c *concourse) Files() Files {
	return FilesData{Server: c}
}

/* FileEntry interface for FileInterface follows */

// Name returns a file name.
func (f *FileEntry) Name() string {
	return f.Title
}

// Size returns a file size in bytes.
func (f *FileEntry) Size() int64 {
	return f.Link.Length
}

// Type returns the file type. video or photo.
func (f *FileEntry) Type() string {
	return f.Link.Type
}

// Date returns the timestamp for a file.
func (f *FileEntry) Date() time.Time {
	return f.Updated
}

// Camera returns the Camera interface for a camera.
func (f *FileEntry) Camera() Camera {
	return f.camera
}

// Save downloads a link from SecuritySpy and saves it to a file.
func (f *FileEntry) Save(path string) error {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return ErrorPathExists
	}
	resp, err := f.server.secReq(f.Link.HREF, make(url.Values), 10*time.Second)
	if err != nil {
		return err
	}
	newFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = newFile.Close()
	}()
	return resp.Write(newFile)
}

// Stream opens a file from a SecuritySpy link and returns the http.Body io.ReadCloser.
func (f *FileEntry) Stream() (io.ReadCloser, error) {
	resp, err := f.server.secReq(f.Link.HREF, make(url.Values), 10*time.Second)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

/* FilesData interface for Files follows */

// GetPhotos returns a list of links to captured images.
func (f FilesData) GetPhotos(cameraNums []int, from, to time.Time) ([]FileInterface, error) {
	return f.getFiles(cameraNums, from, to, "i", "")
}

// GetAll returns a list of links to captured videos and images.
func (f FilesData) GetAll(cameraNums []int, from, to time.Time) ([]FileInterface, error) {
	return f.getFiles(cameraNums, from, to, "b", "")
}

// GetVideos returns a list of links to captured videos.
func (f FilesData) GetVideos(cameraNums []int, from, to time.Time) ([]FileInterface, error) {
	return f.getFiles(cameraNums, from, to, "m", "")
}

/* INTERFACE HELPER METHODS FOLLOW */

// getFiles is a helper function to do all the work for GetVideos, GetPhotos & GetAll.
func (f FilesData) getFiles(cameraNums []int, from, to time.Time, fileType, continuation string) ([]FileInterface, error) {
	var files []FileInterface
	var feed Feed
	params := MakeFilesParams(cameraNums, from, to, fileType, continuation)
	if xmldata, err := f.Server.secReqXML("/++downloads", params); err != nil {
		return nil, err
	} else if err := xml.Unmarshal(xmldata, &feed); err != nil {
		return nil, errors.Wrap(err, "xml.Unmarshal(++downloads)")
	}
	for i, file := range feed.Entry {
		// Add the camera and server interfaces to every file struct/interface.
		file.camera = f.Server.GetCamera(file.CameraNum)
		file.server = f.Server
		files = append(files, &feed.Entry[i])
	}
	// ++downloads automatically paginates. Follow the continuation.
	if feed.Continuation != "" {
		moreFiles, err := f.getFiles(cameraNums, from, to, fileType, feed.Continuation)
		if err != nil {
			// We got some files, but one of the pages returned an error.
			return files, err
		}
		files = append(files, moreFiles...)
	}
	return files, nil
}

// MakeFilesParams makes the url Values for a file retreival.
func MakeFilesParams(cameraNums []int, from time.Time, to time.Time, fileType string, continuation string) url.Values {
	params := make(url.Values)
	params.Set("date2Text", to.Format(fileDateFormat))
	params.Set("date1Text", from.Format(fileDateFormat))
	params.Set("fileTypeMenu", fileType)
	for _, num := range cameraNums {
		params.Add("cameraNum", strconv.Itoa(num))
	}
	if continuation != "" {
		params.Set("continuation", continuation)
	}
	return params
}
