// Package securityspy is a full featured SDK library for interacting with the
// SecuritySpy API: https://www.bensoftware.com/securityspy/web-server-spec.html
package securityspy

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// GetServer returns an iterface to interact with SecuritySpy.
// This is the only exportred function in the library.
// All of the other interfaces are accessed through this interface.
func GetServer(c *Config) (*Server, error) {
	server := &Server{
		Config:     c,
		systemInfo: &systemInfo{Server: &ServerInfo{}},
	}
	if !strings.HasSuffix(server.URL, "/") {
		server.URL += "/"
	}
	if server.Username != "" && server.Password != "" {
		server.Password = base64.URLEncoding.EncodeToString([]byte(server.Username + ":" + server.Password))
	}

	// Assign all the sub-interface structs.
	server.api = server
	server.Info = server.systemInfo.Server
	server.Files = &Files{server: server}
	server.Events = &Events{server: server,
		eventBinds: make(map[EventType][]func(Event)),
		eventChans: make(map[EventType][]chan Event),
	}
	return server, server.Refresh()
}

// Refresh gets fresh camera and serverInfo data from SecuritySpy,
// run this after every action to keep the data pool up to date.
func (s *Server) Refresh() error {
	s.Info.Lock()
	defer s.Info.Unlock()
	if xmldata, err := s.api.secReqXML("++systemInfo", nil); err != nil {
		return err
	} else if err := xml.Unmarshal(xmldata, s.systemInfo); err != nil {
		return errors.Wrap(err, "xml.Unmarshal(++systemInfo)")
	}

	s.Info.Refreshed = time.Now()
	// Point all the unmarshalled data into an exported struct. Better-formatted data.
	s.Info.ServerSchedules = s.systemInfo.Schedules
	s.Info.SchedulePresets = s.systemInfo.SchedulePresets
	s.Info.ScheduleOverrides = s.systemInfo.ScheduleOverrides
	s.Cameras = &Cameras{server: s}
	// Collect the camera names and numbers for user convenience.
	for _, cam := range s.systemInfo.CameraList.Cameras {
		s.Cameras.Names = append(s.Cameras.Names, cam.Name)
		s.Cameras.Numbers = append(s.Cameras.Numbers, cam.Number)
	}
	return nil
}

// GetScripts fetches and returns the list of script files.
// You can't do much with these.
func (s *Server) GetScripts() ([]string, error) {
	var val struct {
		Names []string `xml:"name"`
	}
	if xmldata, err := s.api.secReqXML("++scripts", nil); err != nil {
		return nil, err
	} else if err := xml.Unmarshal(xmldata, &val); err != nil {
		return nil, errors.Wrap(err, "xml.Unmarshal(++scripts)")
	}
	return val.Names, nil
}

// GetSounds fetches and returns the list of sound files.
// You can't do much with these.
func (s *Server) GetSounds() ([]string, error) {
	var val struct {
		Names []string `xml:"name"`
	}
	if xmldata, err := s.api.secReqXML("++sounds", nil); err != nil {
		return nil, err
	} else if err := xml.Unmarshal(xmldata, &val); err != nil {
		return nil, errors.Wrap(err, "xml.Unmarshal(++sounds)")
	}
	return val.Names, nil
}

/* INTERFACE HELPER METHODS FOLLOW */

func (s *Server) getClient(timeout time.Duration) (httpClient *http.Client) {
	return &http.Client{
		Timeout:   timeout,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: !s.VerifySSL}},
	}
}

// secReq is a helper function that formats the http request to SecuritySpy
func (s *Server) secReq(apiPath string, params url.Values, httpClient *http.Client) (*http.Response, error) {
	if params == nil {
		params = make(url.Values)
	}
	if s.Password != "" {
		params.Set("auth", s.Password)
	}
	req, err := http.NewRequest("GET", s.URL+apiPath, nil)
	if err != nil {
		return nil, errors.Wrap(err, "http.NewRequest()")
	}
	if a := apiPath; !strings.HasPrefix(a, "++getfile") && !strings.HasPrefix(a, "++event") &&
		!strings.HasPrefix(a, "++image") && !strings.HasPrefix(a, "++audio") &&
		!strings.HasPrefix(a, "++stream") && !strings.HasPrefix(a, "++video") {
		params.Set("format", "xml")
		req.Header.Add("Accept", "application/xml")
	}
	req.URL.RawQuery = params.Encode()
	return httpClient.Do(req)
}

// secReqXML returns raw http body, so it can be unmarshaled into an xml struct.
func (s *Server) secReqXML(apiPath string, params url.Values) ([]byte, error) {
	resp, err := s.api.secReq(apiPath, params, s.getClient(DefaultTimeout))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("request failed (%v): %v (status: %v/%v)",
			s.Username, s.URL+apiPath, resp.StatusCode, resp.Status)
	}
	return ioutil.ReadAll(resp.Body)
}

// simpleReq performes HTTP req, checks for OK at end of output.
func (s *Server) simpleReq(apiURI string, params url.Values, cameraNum int) error {
	if cameraNum != -1 {
		params.Set("cameraNum", strconv.Itoa(cameraNum))
	}
	resp, err := s.api.secReq(apiURI, params, s.getClient(DefaultTimeout))
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if body, err := ioutil.ReadAll(resp.Body); err != nil || !strings.HasSuffix(string(body), "OK") {
		return ErrorCmdNotOK
	}
	return nil
}
