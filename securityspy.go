// Package securityspy is a full featured SDK library for interacting with the
// SecuritySpy API: https://www.bensoftware.com/securityspy/web-server-spec.html
package securityspy

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"golift.io/securityspy/v2/server"
)

const maxCallbackWorkers = 32

// New returns an interface to interact with SecuritySpy.
func New(c *server.Config) (*Server, error) {
	s := NewMust(c)

	return s, s.RefreshContext(context.Background()) //nolint:gocritic
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
	secspyServer := &Server{Config: config, Encoder: DefaultEncoder, Info: &ServerInfo{}}
	secspyServer.Files = &Files{server: secspyServer}
	secspyServer.Cameras = &Cameras{server: secspyServer}
	secspyServer.Events = &Events{
		server:      secspyServer,
		eventBinds:  make(map[EventType][]func(Event)),
		eventChans:  make(map[EventType][]chan Event),
		callbackSem: make(chan struct{}, maxCallbackWorkers),
	}

	return secspyServer
}

// Refresh gets fresh camera and serverInfo data from SecuritySpy,
// run this after every action to keep the data pool up to date.
// It replaces the Cameras, Groups and Info fields, so other goroutines must
// read those through GetCameras(), GetGroups() and GetInfo() while this can run.
func (s *Server) Refresh() error {
	return s.RefreshContext(context.Background())
}

// GetCameras returns the camera list. Use this instead of the Cameras field
// when another goroutine may call Refresh(), which replaces it.
// The returned *Cameras is a snapshot: a later refresh builds a new one.
func (s *Server) GetCameras() *Cameras {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.Cameras
}

// GetInfo returns the server info. Use this instead of the Info field when
// another goroutine may call Refresh(), which replaces it.
// The returned *ServerInfo is a snapshot: a later refresh builds a new one.
func (s *Server) GetInfo() *ServerInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.Info
}

// GetGroups returns the camera groups. Use this instead of the Groups field
// when another goroutine may call Refresh(), which replaces it.
// The returned slice is a snapshot: a later refresh builds a new one.
func (s *Server) GetGroups() []*Group {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.Groups
}

// RefreshContext gets fresh camera and serverInfo data from SecuritySpy with context support.
func (s *Server) RefreshContext(ctx context.Context) error { //nolint:cyclop // schedule name wiring
	s.mu.Lock()
	defer s.mu.Unlock()

	var sysInfo systemInfo

	if err := s.GetXMLContext(ctx, "++systemInfo", nil, &sysInfo); err != nil {
		return fmt.Errorf("getting systemInfo: %w", err)
	}

	s.Info = sysInfo.Server
	if s.Info == nil {
		s.Info = &ServerInfo{}
	}

	s.Cameras = &Cameras{cameras: sysInfo.cameras(), server: s}
	s.Groups = sysInfo.GroupList.Groups
	s.Info.Refreshed = time.Now()
	// Point all the unmarshalled data into an exported struct. Better-formatted data.
	s.Info.ServerSchedules = sysInfo.schedules()
	s.Info.SchedulePresets = sysInfo.schedulePresets()
	s.Info.ScheduleOverrides = sysInfo.scheduleOverrides()

	for idx, cam := range s.Cameras.cameras {
		s.Cameras.cameras[idx].server = s
		if s.Cameras.cameras[idx].PTZ != nil {
			s.Cameras.cameras[idx].PTZ.camera = s.Cameras.cameras[idx]
		}
		// Fill in the missing schedule names (all we have are IDs, so fetch the names from systemInfo)
		if name, ok := s.Info.ServerSchedules[cam.ScheduleIDA.ID]; ok {
			s.Cameras.cameras[idx].ScheduleIDA.Name = name
		}

		if name, ok := s.Info.ServerSchedules[cam.ScheduleIDCC.ID]; ok {
			s.Cameras.cameras[idx].ScheduleIDCC.Name = name
		}

		if name, ok := s.Info.ServerSchedules[cam.ScheduleIDMC.ID]; ok {
			s.Cameras.cameras[idx].ScheduleIDMC.Name = name
		}

		if name, ok := s.Info.ScheduleOverrides[cam.ScheduleOverrideA.ID]; ok {
			s.Cameras.cameras[idx].ScheduleOverrideA.Name = name
		}

		if name, ok := s.Info.ScheduleOverrides[cam.ScheduleOverrideCC.ID]; ok {
			s.Cameras.cameras[idx].ScheduleOverrideCC.Name = name
		}

		if name, ok := s.Info.ScheduleOverrides[cam.ScheduleOverrideMC.ID]; ok {
			s.Cameras.cameras[idx].ScheduleOverrideMC.Name = name
		}
	}

	return nil
}

// GetScripts fetches and returns the list of script files.
// You can't do much with these.
func (s *Server) GetScripts() ([]string, error) {
	var val struct {
		Names []string `xml:"name"`
	}

	if err := s.GetXML("++scripts", nil, &val); err != nil {
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
