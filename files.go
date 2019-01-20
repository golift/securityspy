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
func (f *filesEntry) Save(path string) (int64, error) {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return 0, ErrorPathExists
	}
	body, err := f.Get()
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = body.Close()
	}()
	newFile, err := os.Create(path)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = newFile.Close()
	}()
	return io.Copy(newFile, body)
}

// Stream opens a file from a SecuritySpy link and returns the http.Body io.ReadCloser.
func (f *filesEntry) Get() (io.ReadCloser, error) {
	resp, err := f.server.secReq(f.Link.HREF, make(url.Values), 10*time.Second)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

/* FilesData interface for Files follows */

// GetPhotos returns a list of links to captured images.
func (f files) GetPhotos(cameraNums []int, from, to time.Time) ([]File, error) {
	return f.getFiles(cameraNums, from, to, "i", "")
}

// GetAll returns a list of links to captured videos and images.
func (f files) GetAll(cameraNums []int, from, to time.Time) ([]File, error) {
	return f.getFiles(cameraNums, from, to, "b", "")
}

// GetVideos returns a list of links to captured videos.
func (f files) GetVideos(cameraNums []int, from, to time.Time) ([]File, error) {
	return f.getFiles(cameraNums, from, to, "m", "")
}

// GetFile returns a file based on the name. It makes a lot of assumptions about file paths.
// Not all methods work with this. Avoid it if possible. This allows Get() and Save() to work.
func (f files) GetFile(name string) (File, error) {
	//	01-18-2019 10-17-53 M Porch.m4v => ++getfile/0/2019-01-18/01-18-2019+10-17-53+M+Porch.m4v
	newFile := &filesEntry{
		Title: name,
		Link: linkInfo{
			Type: "video/quicktime",
			HREF: "++getfile/",
		},
		server: f.Server,
	}
	if strings.Count(name, ".") != 1 {
		return nil, ErrorNoExtension
	}
	split := strings.Split(name, ".")
	if split[1] == "jpg" {
		newFile.Link.Type = "image/jpeg"
	}
	nameOnly := split[0]
	splitName := strings.Split(nameOnly, " ")
	if len(splitName) < 2 {
		return nil, ErrorInvalidName
	}
	dateOnly := splitName[0]
	camName := splitName[len(splitName)-1]
	dateParts := strings.Split(dateOnly, "-")
	if len(dateParts) != 3 {
		return nil, ErrorInvalidName
	}
	pathDate := dateParts[2] + "-" + dateParts[0] + "-" + dateParts[1]
	newFile.camera = f.Cameras.ByName(camName)
	if newFile.camera == nil {
		return nil, ErrorCAMMissing
	}
	newFile.Link.HREF += newFile.camera.Num() + "/" + pathDate + "/" + url.QueryEscape(name)
	return newFile, nil
}

/* INTERFACE HELPER METHODS FOLLOW */

// getFiles is a helper function to do all the work for GetVideos, GetPhotos & GetAll.
func (f files) getFiles(cameraNums []int, from, to time.Time, fileType, continuation string) ([]File, error) {
	var allFiles []File
	var feed fileFeed
	params := makeFilesParams(cameraNums, from, to, fileType, continuation)
	if xmldata, err := f.secReqXML("++download", params); err != nil {
		return nil, err
	} else if err := xml.Unmarshal(xmldata, &feed); err != nil {
		return nil, errors.Wrap(err, "xml.Unmarshal(++download)")
	}
	for i := range feed.Entry {
		// Add the camera and server interfaces to every file struct/interface.
		feed.Entry[i].camera = f.Cameras.ByNum(feed.Entry[i].CameraNum)
		feed.Entry[i].server = f.Server
		feed.Entry[i].GmtOffset, _ = strconv.Atoi(feed.GmtOffset)
		allFiles = append(allFiles, &feed.Entry[i])
	}
	// ++download automatically paginates. Follow the continuation.
	if feed.Continuation != "" && feed.Continuation != "FFFFFFFFFFFFFFFF" {
		moreFiles, err := f.getFiles(cameraNums, from, to, fileType, feed.Continuation)
		if allFiles = append(allFiles, moreFiles...); err != nil {
			// We got some files, but one of the pages returned an error.
			return allFiles, err
		}
	}
	return allFiles, nil
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
