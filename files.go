package securityspy

/* This file handles all file transfers for media saved by security spy. */
// This file and the methods herein: incomplete

import (
	"encoding/xml"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// fileDateFormat is the format the SecuritySpy ++download method accepts.
var fileDateFormat = "01/02/06"

// Files interface allows searching for saved files.
type Files interface {
	GetAll(cameraNums []int, from, to time.Time) ([]File, error)
	GetPhotos(cameraNums []int, from, to time.Time) ([]File, error)
	GetVideos(cameraNums []int, from, to time.Time) ([]File, error)
	GetFile(name string) (file io.ReadCloser, err error)
}

// filesData powers the Files interface.
// It's really an extension of the concourse interface.
type filesData struct {
	*concourse
}

// fileFeed represents the XML data from ++download
type fileFeed struct {
	XMLName      xml.Name     `xml:"feed"`
	BSL          string       `xml:"bsl,attr"`     // http://www.bensoftware.com/
	Title        string       `xml:"title"`        // Downloads
	GmtOffset    string       `xml:"gmt-offset"`   // -28800
	Continuation string       `xml:"continuation"` // 0007E3010C0E1D3A
	Entry        []filesEntry `xml:"entry"`
}

// filesEntry represents a saved media file.
type filesEntry struct {
	Title string `xml:"title"` // 01-12-2019 M Gate.m4v, 01...
	Link  struct {
		Rel    string `xml:"rel,attr"`    // alternate, alternate, alternate
		Type   string `xml:"type,attr"`   // video/quicktime, video/quicktime
		Length int64  `xml:"length,attr"` // 358472320, 483306152, 900789978,
		HREF   string `xml:"href,attr"`   // ++getfile/4/2018-10-17/10-17-2018+M+Gate.m4v
	} `xml:"link"`
	Updated   time.Time `xml:"updated"`   // 2019-01-12T08:57:58Z, 201...
	CameraNum int       `xml:"cameraNum"` // 0, 1, 2, 4, 5, 7, 9, 10, 11, 12, 13
	GmtOffset int
	server    *concourse
	camera    Camera
}

// File is used to do something with a filesEntry.
type File interface {
	Name() string
	Size() int64
	Type() string
	Date() time.Time
	Offset() int
	Save(path string) error
	Camera() Camera
	Stream() (io.ReadCloser, error)
}

/* Files-specific concourse methods are at the top. */

// Files returns a Files interface, used to retreive file listings.
func (c *concourse) Files() Files {
	return filesData{c}
}

/* FileEntry interface for FileInterface follows */

// Name returns a file name.
func (f *filesEntry) Name() string {
	return f.Title
}

// Size returns a file size in bytes.
func (f *filesEntry) Size() int64 {
	return f.Link.Length
}

// Type returns the file type. video or photo.
func (f *filesEntry) Type() string {
	return f.Link.Type
}

// Date returns the timestamp for a file.
func (f *filesEntry) Date() time.Time {
	return f.Updated
}

// Date returns the GMT offset for a file's date.
func (f *filesEntry) Offset() int {
	return f.GmtOffset
}

// Camera returns the Camera interface for a camera.
func (f *filesEntry) Camera() Camera {
	return f.camera
}

// Save downloads a link from SecuritySpy and saves it to a file.
func (f *filesEntry) Save(path string) error {
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
func (f *filesEntry) Stream() (io.ReadCloser, error) {
	resp, err := f.server.secReq(f.Link.HREF, make(url.Values), 10*time.Second)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

/* FilesData interface for Files follows */

// GetPhotos returns a list of links to captured images.
func (f filesData) GetPhotos(cameraNums []int, from, to time.Time) ([]File, error) {
	return f.getFiles(cameraNums, from, to, "i", "")
}

// GetAll returns a list of links to captured videos and images.
func (f filesData) GetAll(cameraNums []int, from, to time.Time) ([]File, error) {
	return f.getFiles(cameraNums, from, to, "b", "")
}

// GetVideos returns a list of links to captured videos.
func (f filesData) GetVideos(cameraNums []int, from, to time.Time) ([]File, error) {
	return f.getFiles(cameraNums, from, to, "m", "")
}

// GetFile returns a file based on the name. It makes a lot of assumptions about file paths.
func (f filesData) GetFile(name string) (file io.ReadCloser, err error) {
	//	++getfile/0/2019-01-18/01-18-2019+10-17-53 M Porch.m4v
	//  ++getfile/5/2019-01-18/01-18-2019 M Pool.m4v
	nameOnly := strings.Split(name, ".")[0]
	splitName := strings.Split(nameOnly, " ")
	if len(splitName) < 2 {
		return nil, ErrorCAMMissing
	}
	dateOnly := splitName[0]
	camName := splitName[len(splitName)-1]
	dateParts := strings.Split(dateOnly, "-")
	if len(dateParts) != 3 {
		return nil, ErrorCAMMissing
	}
	pathDate := dateParts[2] + "-" + dateParts[0] + "-" + dateParts[1]
	camera := f.GetCameraByName(camName)
	if camera == nil {
		return nil, ErrorCAMMissing
	}
	filePath := "++getfile/" + camera.Num() + "/" + pathDate + "/" + url.QueryEscape(name)
	resp, err := f.secReq(filePath, nil, 10*time.Second)
	return resp.Body, err
}

/* INTERFACE HELPER METHODS FOLLOW */

// getFiles is a helper function to do all the work for GetVideos, GetPhotos & GetAll.
func (f filesData) getFiles(cameraNums []int, from, to time.Time, fileType, continuation string) ([]File, error) {
	var files []File
	var feed fileFeed
	params := makeFilesParams(cameraNums, from, to, fileType, continuation)
	if xmldata, err := f.secReqXML("++download", params); err != nil {
		return nil, err
	} else if err := xml.Unmarshal(xmldata, &feed); err != nil {
		return nil, errors.Wrap(err, "xml.Unmarshal(++download)")
	}
	for i := range feed.Entry {
		// Add the camera and server interfaces to every file struct/interface.
		feed.Entry[i].camera = f.GetCamera(feed.Entry[i].CameraNum)
		feed.Entry[i].server = f.concourse
		feed.Entry[i].GmtOffset, _ = strconv.Atoi(feed.GmtOffset)
		files = append(files, &feed.Entry[i])
	}
	// ++download automatically paginates. Follow the continuation.
	if feed.Continuation != "" && feed.Continuation != "FFFFFFFFFFFFFFFF" {
		moreFiles, err := f.getFiles(cameraNums, from, to, fileType, feed.Continuation)
		if files = append(files, moreFiles...); err != nil {
			// We got some files, but one of the pages returned an error.
			return files, err
		}
	}
	return files, nil
}

// makeFilesParams makes the url Values for a file retreival.
func makeFilesParams(cameraNums []int, from time.Time, to time.Time, fileType string, continuation string) url.Values {
	params := make(url.Values)
	params.Set("date1Text", from.Format(fileDateFormat))
	params.Set("date2Text", to.Format(fileDateFormat))
	params.Set("fileTypeMenu", fileType)
	for _, num := range cameraNums {
		params.Add("cameraNum", strconv.Itoa(num))
	}
	if continuation != "" {
		params.Set("continuation", continuation)
	}
	return params
}
