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
//
// The systemInfo request and all the wiring happen off to the side, so readers
// keep serving the previous snapshot for the length of the round trip and only
// block for the swap at the end. A refresh that times out never stalls them.
func (s *Server) RefreshContext(ctx context.Context) error {
	// refreshMu serializes refreshes with each other, mu only guards the swap.
	s.refreshMu.Lock()
	defer s.refreshMu.Unlock()

	var sysInfo systemInfo

	err := s.GetXMLContext(ctx, "++systemInfo", nil, &sysInfo)
	if err != nil {
		return fmt.Errorf("getting systemInfo: %w", err)
	}

	info := sysInfo.Server
	if info == nil {
		info = &ServerInfo{}
	}

	info.Refreshed = time.Now()
	// Point all the unmarshalled data into an exported struct. Better-formatted data.
	info.ServerSchedules = sysInfo.schedules()
	info.SchedulePresets = sysInfo.schedulePresets()
	info.ScheduleOverrides = sysInfo.scheduleOverrides()

	cameras := &Cameras{cameras: sysInfo.cameras(), server: s}
	s.wireCameras(cameras.cameras, info)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.Info = info
	s.Cameras = cameras
	s.Groups = sysInfo.GroupList.Groups

	return nil
}

// wireCameras points every camera back at the server and fills in the schedule
// names, which systemInfo only provides as IDs. Runs on a new camera list before
// it is published, so no lock is needed.
func (s *Server) wireCameras(cameras []*Camera, info *ServerInfo) {
	for idx, cam := range cameras {
		cameras[idx].server = s
		if cameras[idx].PTZ != nil {
			cameras[idx].PTZ.camera = cameras[idx]
		}

		if name, ok := info.ServerSchedules[cam.ScheduleIDA.ID]; ok {
			cameras[idx].ScheduleIDA.Name = name
		}

		if name, ok := info.ServerSchedules[cam.ScheduleIDCC.ID]; ok {
			cameras[idx].ScheduleIDCC.Name = name
		}

		if name, ok := info.ServerSchedules[cam.ScheduleIDMC.ID]; ok {
			cameras[idx].ScheduleIDMC.Name = name
		}

		if name, ok := info.ScheduleOverrides[cam.ScheduleOverrideA.ID]; ok {
			cameras[idx].ScheduleOverrideA.Name = name
		}

		if name, ok := info.ScheduleOverrides[cam.ScheduleOverrideCC.ID]; ok {
			cameras[idx].ScheduleOverrideCC.Name = name
		}

		if name, ok := info.ScheduleOverrides[cam.ScheduleOverrideMC.ID]; ok {
			cameras[idx].ScheduleOverrideMC.Name = name
		}
	}
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
