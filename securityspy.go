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
func GetServer(user, pass, url string, verifySSL bool) (*Server, error) {
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	authB64 := ""
	if user != "" && pass != "" {
		authB64 = base64.URLEncoding.EncodeToString([]byte(user + ":" + pass))
	}
	server := &Server{
		systemInfo: &systemInfo{Server: &ServerInfo{}},
		baseURL:    url,
		authB64:    authB64,
		username:   user,
		verifySSL:  verifySSL,
	}
	// Assign all the sub-interface structs.
	server.Info = server.systemInfo.Server
	server.Files = &Files{server: server}
	server.Events = &Events{server: server,
		eventBinds: make(map[EventType][]func(Event)),
		eventChans: make(map[EventType][]chan Event),
	}

	// Run three API methods to fill in the Server data
	// structure when a new server is created. Return any error.
	if err := server.Refresh(); err != nil {
		return server, err
	} else if err := server.RefreshScripts(); err != nil {
		return server, err
	} else if err := server.RefreshSounds(); err != nil {
		return server, err
	}
	return server, nil
}

// Refresh gets fresh camera and serverInfo data from SecuritySpy,
// run this after every action to keep the data pool up to date.
func (s *Server) Refresh() error {
	s.Info.Lock()
	defer s.Info.Unlock()
	if xmldata, err := s.secReqXML("++systemInfo", nil); err != nil {
		return err
	} else if err := xml.Unmarshal(xmldata, s.systemInfo); err != nil {
		return errors.Wrap(err, "xml.Unmarshal(++systemInfo)")
	}

	// Point all the unmarshalled data into an exported struct. Better-formatted data.
	s.Info.Refreshed = time.Now()
	s.Info.ServerSchedules = s.systemInfo.Schedules
	s.Info.SchedulePresets = s.systemInfo.SchedulePresets
	s.Info.ScheduleOverrides = s.systemInfo.ScheduleOverrides
	s.Cameras = &Cameras{
		server: s,
		Count:  len(s.systemInfo.CameraList.Cameras),
	}
	// Add the name to each assigned Camera Schedule and Schedule Override.
	for i, cam := range s.systemInfo.CameraList.Cameras {
		s.systemInfo.CameraList.Cameras[i].ScheduleIDA.Name = s.Info.ServerSchedules[cam.ScheduleIDA.ID]
		s.systemInfo.CameraList.Cameras[i].ScheduleIDCC.Name = s.Info.ServerSchedules[cam.ScheduleIDCC.ID]
		s.systemInfo.CameraList.Cameras[i].ScheduleIDMC.Name = s.Info.ServerSchedules[cam.ScheduleIDMC.ID]
		s.systemInfo.CameraList.Cameras[i].ScheduleOverrideA.Name = s.Info.ScheduleOverrides[cam.ScheduleOverrideA.ID]
		s.systemInfo.CameraList.Cameras[i].ScheduleOverrideCC.Name = s.Info.ScheduleOverrides[cam.ScheduleOverrideCC.ID]
		s.systemInfo.CameraList.Cameras[i].ScheduleOverrideMC.Name = s.Info.ScheduleOverrides[cam.ScheduleOverrideMC.ID]
		s.Cameras.Names = append(s.Cameras.Names, cam.Name)
		s.Cameras.Numbers = append(s.Cameras.Numbers, cam.Number)
	}
	return nil
}

// RefreshScripts refreshes the list of script files. Probably doesn't change much.
// Retreivable as server.Info.ScriptsNames
func (s *Server) RefreshScripts() error {
	if xmldata, err := s.secReqXML("++scripts", nil); err != nil {
		return err
	} else if err := xml.Unmarshal(xmldata, &s.systemInfo.Scripts); err != nil {
		return errors.Wrap(err, "xml.Unmarshal(++scripts)")
	}
	s.Info.ScriptsNames = s.systemInfo.Scripts.Names
	return nil
}

// RefreshSounds refreshes the list of sound files. Probably doesn't change much.
// Retreivable as server.Info.SoundsNames
func (s *Server) RefreshSounds() error {
	if xmldata, err := s.secReqXML("++sounds", nil); err != nil {
		return err
	} else if err := xml.Unmarshal(xmldata, &s.systemInfo.Sounds); err != nil {
		return errors.Wrap(err, "xml.Unmarshal(++sounds)")
	}
	s.Info.SoundsNames = s.systemInfo.Sounds.Names
	return nil
}

/* INTERFACE HELPER METHODS FOLLOW */

// secReq is a helper function that formats the http request to SecuritySpy
func (s *Server) secReq(apiPath string, params url.Values, timeout time.Duration) (resp *http.Response, err error) {
	if params == nil {
		params = make(url.Values)
	}
	a := &http.Client{Timeout: timeout, Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: !s.verifySSL}}}
	req, err := http.NewRequest("GET", s.baseURL+apiPath, nil)
	if err != nil {
		return resp, errors.Wrap(err, "http.NewRequest()")
	}
	if s.authB64 != "" {
		params.Set("auth", s.authB64)
	}
	if a := apiPath; !strings.HasPrefix(a, "++getfile") && !strings.HasPrefix(a, "++event") &&
		!strings.HasPrefix(a, "++image") && !strings.HasPrefix(a, "++audio") &&
		!strings.HasPrefix(a, "++stream") && !strings.HasPrefix(a, "++video") {
		params.Set("format", "xml")
		req.Header.Add("Accept", "application/xml")
	}
	req.URL.RawQuery = params.Encode()
	resp, err = a.Do(req)
	if err != nil {
		return resp, errors.Wrap(err, "http.Do(req)")
	}
	return resp, nil
}

// secReqXML returns raw http body, so it can be unmarshaled into an xml struct.
func (s *Server) secReqXML(apiPath string, params url.Values) (body []byte, err error) {
	resp, err := s.secReq(apiPath, params, DefaultTimeout)
	if err != nil {
		return body, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return body, errors.Errorf("request failed (%v): %v (status: %v/%v)",
			s.username, s.baseURL+apiPath, resp.StatusCode, resp.Status)
	} else if body, err = ioutil.ReadAll(resp.Body); err == nil {
		return body, errors.Wrap(err, "ioutil.ReadAll(resp.Body)")
	}
	return body, nil
}

// simpleReq performes HTTP req, checks for OK at end of output.
func (s *Server) simpleReq(apiURI string, params url.Values, cameraNum int) error {
	if cameraNum != -1 {
		params.Set("cameraNum", strconv.Itoa(cameraNum))
	}
	resp, err := s.secReq(apiURI, params, DefaultTimeout)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if body, err := ioutil.ReadAll(resp.Body); err != nil {
		return err
	} else if !strings.HasSuffix(string(body), "OK") {
		return ErrorCmdNotOK
	}
	return nil
}
