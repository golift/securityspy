// Package securityspy is a full featured SDK library for interacting with the
// SecuritySpy API: https://www.bensoftware.com/securityspy/web-server-spec.html
package securityspy

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"golift.io/securityspy/server"
)

// New returns an iterface to interact with SecuritySpy.
func New(c *server.Config) (*Server, error) {
	s := NewMust(c)

	return s, s.Refresh()
}

// NewMust returns an iterface to interact with SecuritySpy.
// This does not attempt to connect to SecuritySpy first.
// You must call s.Refresh() before attempting to access other datas.
func NewMust(config *server.Config) *Server {
	if !strings.HasSuffix(config.URL, "/") {
		config.URL += "/"
	}

	if config.Username != "" && config.Password != "" {
		config.Password = base64.URLEncoding.EncodeToString([]byte(config.Username + ":" + config.Password))
	}

	// Assign all the sub-interface structs.
	server := &Server{API: config, Encoder: DefaultEncoder}
	server.Files = &Files{server: server}
	server.Events = &Events{
		server:     server,
		eventBinds: make(map[EventType][]func(Event)),
		eventChans: make(map[EventType][]chan Event),
	}

	return server
}

// Refresh gets fresh camera and serverInfo data from SecuritySpy,
// run this after every action to keep the data pool up to date.
// This is not at all thread safe. Do not run this if other methods
// may run in a different go routine.
func (s *Server) Refresh() error {
	s.Lock()
	defer s.Unlock()

	var sysInfo systemInfo

	if err := s.GetXML("++systemInfo", nil, &sysInfo); err != nil {
		return fmt.Errorf("getting systemInfo: %w", err)
	}

	s.Info = sysInfo.Server
	s.Cameras = &Cameras{cameras: sysInfo.CameraList.Cameras, server: s}
	s.Info.Refreshed = time.Now()
	// Point all the unmarshalled data into an exported struct. Better-formatted data.
	s.Info.ServerSchedules = sysInfo.Schedules
	s.Info.SchedulePresets = sysInfo.SchedulePresets
	s.Info.ScheduleOverrides = sysInfo.ScheduleOverrides

	for idx, cam := range s.Cameras.cameras {
		s.Cameras.cameras[idx].server = s
		s.Cameras.cameras[idx].PTZ.camera = s.Cameras.cameras[idx]
		// Fill in the missing schedule names (all we have are IDs, so fetch the names from systemInfo)
		s.Cameras.cameras[idx].ScheduleIDA.Name = s.Info.ServerSchedules[cam.ScheduleIDA.ID]
		s.Cameras.cameras[idx].ScheduleIDCC.Name = s.Info.ServerSchedules[cam.ScheduleIDCC.ID]
		s.Cameras.cameras[idx].ScheduleIDMC.Name = s.Info.ServerSchedules[cam.ScheduleIDMC.ID]
		s.Cameras.cameras[idx].ScheduleOverrideA.Name = s.Info.ScheduleOverrides[cam.ScheduleOverrideA.ID]
		s.Cameras.cameras[idx].ScheduleOverrideCC.Name = s.Info.ScheduleOverrides[cam.ScheduleOverrideCC.ID]
		s.Cameras.cameras[idx].ScheduleOverrideMC.Name = s.Info.ScheduleOverrides[cam.ScheduleOverrideMC.ID]
	}

	return nil
}

// GetScripts fetches and returns the list of script files.
// You can't do much with these.
func (s *Server) GetScripts() ([]string, error) {
	var val struct {
		Names []string `xml:"name"`
	}

	if err := s.API.GetXML("++scripts", nil, &val); err != nil {
		return nil, fmt.Errorf("getting scripts: %w", err)
	}

	return val.Names, nil
}

// GetSounds fetches and returns the list of sound files.
// You can't do much with these.
func (s *Server) GetSounds() ([]string, error) {
	var val struct {
		Names []string `xml:"name"`
	}

	if err := s.GetXML("++sounds", nil, &val); err != nil {
		return nil, fmt.Errorf("getting sounds: %w", err)
	}

	return val.Names, nil
}
