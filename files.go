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

// GetImages returns a list of links to captured images.
func (f *Files) GetImages(cameraNums []int, from, to time.Time) ([]*File, error) {
	return f.getFiles(cameraNums, from, to, "imageFilesCheck", "")
}

// GetAll returns a list of links to all captured videos and images.
func (f *Files) GetAll(cameraNums []int, from, to time.Time) ([]*File, error) {
	return f.getFiles(cameraNums, from, to, "ccFilesCheck&mcFilesCheck&imageFilesCheck", "")
}

// GetMCVideos returns a list of links to motion-captured videos.
func (f *Files) GetMCVideos(cameraNums []int, from, to time.Time) ([]*File, error) {
	return f.getFiles(cameraNums, from, to, "mcFilesCheck", "")
}

// GetCCVideos returns a list of links to continuous-captured videos.
func (f *Files) GetCCVideos(cameraNums []int, from, to time.Time) ([]*File, error) {
	return f.getFiles(cameraNums, from, to, "ccFilesCheck", "")
}

// GetFile returns a file based on the name. It makes a lot of assumptions about file paths.
// Not all methods work with this. Avoid it if possible. This allows Get() and Save() to work.
func (f *Files) GetFile(name string) (*File, error) {
	//	01-18-2019 10-17-53 M Porch.m4v => ++getfile/0/2019-01-18/01-18-2019+10-17-53+M+Porch.m4v
	var err error
	file := new(File)
	if fileExtSplit := strings.Split(name, "."); len(fileExtSplit) != 2 {
		return file, ErrorNoExtension
	} else if nameDateSplit := strings.Split(fileExtSplit[0], " "); len(fileExtSplit) < 2 {
		return file, ErrorInvalidName
	} else if file.Updated, err = time.Parse(fileDateFormat, nameDateSplit[0]); err != nil {
		return file, ErrorInvalidName
	} else if file.Camera = f.server.Cameras.ByName(nameDateSplit[len(nameDateSplit)-1]); file.Camera == nil {
		return file, ErrorCAMMissing
	} else if file.Link.Type = "video/quicktime"; fileExtSplit[1] == "jpg" {
		file.Link.Type = "image/jpeg"
	}
	file.Title = name
	file.server = f.server
	file.CameraNum = file.Camera.Number
	file.GmtOffset = f.server.Info.GmtOffset
	file.Link.HREF = "++getfile/" + strconv.Itoa(file.CameraNum) + "/" +
		file.Updated.Format(downloadDateFormat) + "/" + url.QueryEscape(name)
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
	size, err := io.Copy(newFile, body)
	if err != nil {
		_ = newFile.Close()
		return size, nil
	}
	return size, newFile.Close()
}

// Get opens a file from a SecuritySpy link and returns the http.Body io.ReadCloser.
func (f *File) Get() (io.ReadCloser, error) {
	// use high bandwidth (full size) file download.
	uri := strings.Replace(f.Link.HREF, "++getfile", "++getfilehb", 1)
	resp, err := f.server.secReq(uri, make(url.Values), DefaultTimeout)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

/* INTERFACE HELPER METHODS FOLLOW */

// getFiles is a helper function to do all the work for GetVideos, GetPhotos & GetAll.
func (f *Files) getFiles(cameraNums []int, from, to time.Time, fileTypes, continuation string) ([]*File, error) {
	var entries []*File
	var feed fileFeed
	params := makeFilesParams(cameraNums, from, to, fileTypes, continuation)
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
		moreFiles, err := f.getFiles(cameraNums, from, to, fileTypes, feed.Continuation)
		if entries = append(entries, moreFiles...); err != nil {
			// We got some files, but one of the pages returned an error.
			return entries, err
		}
	}
	return entries, nil
}

// makeFilesParams makes the url Values for a file retreival.
func makeFilesParams(cameraNums []int, from time.Time, to time.Time, fileTypes string, continuation string) url.Values {
	params := make(url.Values)
	params.Set("results", "1000")
	params.Set("date1", from.Format(downloadDateFormat))
	params.Set("date2", to.Format(downloadDateFormat))
	for _, fileType := range strings.Split(fileTypes, "&") {
		params.Set(fileType, "1")
	}
	for _, num := range cameraNums {
		params.Add("cameraNum", strconv.Itoa(num))
	}
	if continuation != "" {
		params.Set("continuation", continuation)
	}
	return params
}
