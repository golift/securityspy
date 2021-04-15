package securityspy

/* This file handles all file transfers for media saved by security spy. */

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
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

/* Files interface methods follow. */

// GetImages returns a list of File interfaces to captured images.
// Takes in a list of Camera Numbers, as well as a start and stop time to filter results.
func (f *Files) GetImages(cameraNums []int, from, to time.Time) ([]*File, error) {
	return f.getFiles(cameraNums, from, to, "imageFilesCheck", "")
}

// GetAll returns a list of File interfaces to all captured videos and images.
// Takes in a list of Camera Numbers, as well as a start and stop time to filter results.
func (f *Files) GetAll(cameraNums []int, from, to time.Time) ([]*File, error) {
	return f.getFiles(cameraNums, from, to, "ccFilesCheck&mcFilesCheck&imageFilesCheck", "")
}

// GetMCVideos returns a list of File interfaces to motion-captured videos.
// Takes in a list of Camera Numbers, as well as a start and stop time to filter results.
func (f *Files) GetMCVideos(cameraNums []int, from, to time.Time) ([]*File, error) {
	return f.getFiles(cameraNums, from, to, "mcFilesCheck", "")
}

// GetCCVideos returns a list of File interfaces to continuous-captured videos.
// Takes in a list of Camera Numbers, as well as a start and stop time to filter results.
func (f *Files) GetCCVideos(cameraNums []int, from, to time.Time) ([]*File, error) {
	return f.getFiles(cameraNums, from, to, "ccFilesCheck", "")
}

const fileParts = 2

// GetFile returns a file based on the name. It makes a lot of assumptions about file paths.
// Not all methods work with this. Avoid it if possible. This allows Get() and Save() to work
// for an arbitrary file name.
func (f *Files) GetFile(name string) (*File, error) {
	//	01-18-2019 10-17-53 M Porch.m4v => ++getfile/0/2019-01-18/01-18-2019+10-17-53+M+Porch.m4v
	var err error

	file := &File{
		Title:     name,
		server:    f.server,
		GmtOffset: f.server.Info.GmtOffset.Duration,
	}

	if fileExtSplit := strings.Split(name, "."); len(fileExtSplit) != fileParts {
		return file, ErrorInvalidName
	} else if nameDateSplit := strings.Split(fileExtSplit[0], " "); len(fileExtSplit) < fileParts {
		return file, ErrorInvalidName
	} else if file.Updated, err = time.Parse(FileDateFormat, nameDateSplit[0]); err != nil {
		return file, ErrorInvalidName
	} else if file.Camera = f.server.Cameras.ByName(nameDateSplit[len(nameDateSplit)-1]); file.Camera == nil {
		return file, ErrorCAMMissing
	} else if file.Link.Type = "video/quicktime"; fileExtSplit[1] == "jpg" {
		file.Link.Type = "image/jpeg"
	}

	file.CameraNum = file.Camera.Number
	file.Link.HREF = "++getfile/" + strconv.Itoa(file.CameraNum) + "/" +
		file.Updated.Format(DownloadDateFormat) + "/" + url.QueryEscape(name)

	return file, nil
}

/* File interface methods follow. */

// Save downloads a saved media File from SecuritySpy and saves it to a local file.
// Returns an error if path exists.
func (f *File) Save(path string) (int64, error) {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return 0, ErrorPathExists
	}

	body, err := f.Get(true)
	if err != nil {
		return 0, err
	}
	defer body.Close()

	newFile, err := os.Create(path)
	if err != nil {
		return 0, fmt.Errorf("os.Create(): %w", err)
	}
	defer newFile.Close()

	size, err := io.Copy(newFile, body)
	if err != nil {
		return size, fmt.Errorf("io.Copy(): %w", err)
	}

	return size, nil
}

// Get opens a file from a SecuritySpy download href and returns the http.Body io.ReadCloser.
// Close() the Closer when finished. Pass true (for highBandwidth) will download
// the full size file. Passing false will download a smaller transcoded file.
func (f *File) Get(highBandwidth bool) (io.ReadCloser, error) {
	// use high bandwidth (full size) file download.
	uri := strings.Replace(f.Link.HREF, "++getfile/", "++getfilelb/", 1)

	if highBandwidth {
		uri = strings.Replace(f.Link.HREF, "++getfile/", "++getfilehb/", 1)
	}

	resp, err := f.server.Get(uri, make(url.Values))
	if err != nil {
		return nil, fmt.Errorf("getting file: %w", err)
	}

	return resp.Body, nil
}

/* INTERFACE HELPER METHODS FOLLOW */

// getFiles is a helper function to do all the work for GetVideos, GetPhotos & GetAll.
func (f *Files) getFiles(cameraNums []int, from, to time.Time, fileTypes, continuation string) ([]*File, error) {
	var (
		entries = []*File{}
		feed    fileFeed
		params  = makeFilesParams(cameraNums, from, to, fileTypes, continuation)
	)

	if err := f.server.GetXML("++download", params, &feed); err != nil {
		return nil, fmt.Errorf("getting download: %w", err)
	}

	for i := range feed.Entries {
		// Add the camera, server and file interfaces to every file entry.
		feed.Entries[i].Camera = f.server.Cameras.ByNum(feed.Entries[i].CameraNum)
		feed.Entries[i].server = f.server
		feed.Entries[i].GmtOffset = feed.GmtOffset.Duration
		entries = append(entries, feed.Entries[i])
	}

	// ++download automatically paginates. Follow the continuation.
	if feed.Continuation != "" && feed.Continuation != "FFFFFFFFFFFFFFFF" {
		moreFiles, err := f.getFiles(cameraNums, from, to, fileTypes, feed.Continuation)
		if err != nil { // We got some files, but one of the pages returned an error.
			return entries, err
		}

		entries = append(entries, moreFiles...)
	}

	return entries, nil
}

// makeFilesParams makes the url Values for a file retreival.
func makeFilesParams(cameraNums []int, from time.Time, to time.Time, fileTypes string, continuation string) url.Values {
	params := make(url.Values)
	params.Set("results", "1000")
	params.Set("date1", from.Format(DownloadDateFormat))
	params.Set("date2", to.Format(DownloadDateFormat))

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
