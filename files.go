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

/* Files interface methods follow. */

// GetPhotos returns a list of links to captured images.
func (f *Files) GetPhotos(cameraNums []int, from, to time.Time) ([]*File, error) {
	return f.getFiles(cameraNums, from, to, "i", "")
}

// GetAll returns a list of links to captured videos and images.
func (f *Files) GetAll(cameraNums []int, from, to time.Time) ([]*File, error) {
	return f.getFiles(cameraNums, from, to, "b", "")
}

// GetVideos returns a list of links to captured videos.
func (f *Files) GetVideos(cameraNums []int, from, to time.Time) ([]*File, error) {
	return f.getFiles(cameraNums, from, to, "m", "")
}

// GetFile returns a file based on the name. It makes a lot of assumptions about file paths.
// Not all methods work with this. Avoid it if possible. This allows Get() and Save() to work.
func (f *Files) GetFile(name string) (*File, error) {
	//	01-18-2019 10-17-53 M Porch.m4v => ++getfile/0/2019-01-18/01-18-2019+10-17-53+M+Porch.m4v
	file := new(File)
	file.Title = name
	file.server = f.server
	if strings.Count(name, ".") != 1 {
		return nil, ErrorNoExtension
	}
	split := strings.Split(name, ".")
	if file.Link.Type = "video/quicktime"; split[1] == "jpg" {
		file.Link.Type = "image/jpeg"
	}
	if split = strings.Split(split[0], " "); len(split) < 2 {
		return nil, ErrorInvalidName
	}
	camName := split[len(split)-1]
	if split = strings.Split(split[0], "-"); len(split) != 3 {
		return nil, ErrorInvalidName
	}
	pathDate := split[2] + "-" + split[0] + "-" + split[1]
	if file.Camera = f.server.Cameras.ByName(camName); file.Camera == nil {
		return nil, ErrorCAMMissing
	}
	file.CameraNum = file.Camera.Number
	file.Link.HREF = "++getfile/" + strconv.Itoa(file.CameraNum) + "/" + pathDate + "/" + url.QueryEscape(name)
	return file, nil
}

/* File interface methods follow. */

// Save downloads a link from SecuritySpy and saves it to a file.
func (f *File) Save(path string) (int64, error) {
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

// Get opens a file from a SecuritySpy link and returns the http.Body io.ReadCloser.
func (f *File) Get() (io.ReadCloser, error) {
	resp, err := f.server.secReq(f.Link.HREF, make(url.Values), 10*time.Second)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

/* INTERFACE HELPER METHODS FOLLOW */

// getFiles is a helper function to do all the work for GetVideos, GetPhotos & GetAll.
func (f *Files) getFiles(cameraNums []int, from, to time.Time, fileType, continuation string) ([]*File, error) {
	var entries []*File
	var feed fileFeed
	params := makeFilesParams(cameraNums, from, to, fileType, continuation)
	if xmldata, err := f.server.secReqXML("++download", params); err != nil {
		return nil, err
	} else if err := xml.Unmarshal(xmldata, &feed); err != nil {
		return nil, errors.Wrap(err, "xml.Unmarshal(++download)")
	}
	for i := range feed.Entries {
		// Add the camera, server and file interfaces to every file entry.
		feed.Entries[i].Camera = f.server.Cameras.ByNum(feed.Entries[i].CameraNum)
		feed.Entries[i].server = f.server
		feed.Entries[i].GmtOffset, _ = strconv.Atoi(feed.GmtOffset)
		entries = append(entries, feed.Entries[i])
	}
	// ++download automatically paginates. Follow the continuation.
	if feed.Continuation != "" && feed.Continuation != "FFFFFFFFFFFFFFFF" {
		moreFiles, err := f.getFiles(cameraNums, from, to, fileType, feed.Continuation)
		if entries = append(entries, moreFiles...); err != nil {
			// We got some files, but one of the pages returned an error.
			return entries, err
		}
	}
	return entries, nil
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
