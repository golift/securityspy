package securityspy

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

// ErrCameraNumRequired is returned when a settings call needs cameraNum.
var ErrCameraNumRequired = errors.New("cameraNum required")

// GetGeneralSettings fetches ++settings-general.
func (s *Server) GetGeneralSettings() (*GeneralSettings, error) {
	var val GeneralSettings
	if err := s.GetXML("++settings-general", nil, &val); err != nil {
		return nil, fmt.Errorf("getting general settings: %w", err)
	}

	return &val, nil
}

// SetGeneralSettings posts form fields to ++settings-general (partial update).
func (s *Server) SetGeneralSettings(form url.Values) error {
	if err := s.PostForm("++settings-general", form); err != nil {
		return fmt.Errorf("setting general settings: %w", err)
	}

	return nil
}

// GetDisplaySettings fetches ++settings-display.
func (s *Server) GetDisplaySettings() (*DisplaySettings, error) {
	var val DisplaySettings
	if err := s.GetXML("++settings-display", nil, &val); err != nil {
		return nil, fmt.Errorf("getting display settings: %w", err)
	}

	return &val, nil
}

// SetDisplaySettings posts form fields to ++settings-display (partial update).
func (s *Server) SetDisplaySettings(form url.Values) error {
	if err := s.PostForm("++settings-display", form); err != nil {
		return fmt.Errorf("setting display settings: %w", err)
	}

	return nil
}

// GetStorageSettings fetches ++settings-storage.
func (s *Server) GetStorageSettings() (*StorageSettings, error) {
	var val StorageSettings
	if err := s.GetXML("++settings-storage", nil, &val); err != nil {
		return nil, fmt.Errorf("getting storage settings: %w", err)
	}

	return &val, nil
}

// SetStorageSettings posts form fields to ++settings-storage (partial update).
func (s *Server) SetStorageSettings(form url.Values) error {
	if err := s.PostForm("++settings-storage", form); err != nil {
		return fmt.Errorf("setting storage settings: %w", err)
	}

	return nil
}

// GetCompressionSettings fetches ++settings-compression.
func (s *Server) GetCompressionSettings() (*CompressionSettings, error) {
	var val CompressionSettings
	if err := s.GetXML("++settings-compression", nil, &val); err != nil {
		return nil, fmt.Errorf("getting compression settings: %w", err)
	}

	return &val, nil
}

// SetCompressionSettings posts form fields to ++settings-compression (partial update).
func (s *Server) SetCompressionSettings(form url.Values) error {
	if err := s.PostForm("++settings-compression", form); err != nil {
		return fmt.Errorf("setting compression settings: %w", err)
	}

	return nil
}

// GetEmailSettings fetches ++settings-email.
func (s *Server) GetEmailSettings() (*EmailSettings, error) {
	var val EmailSettings
	if err := s.GetXML("++settings-email", nil, &val); err != nil {
		return nil, fmt.Errorf("getting email settings: %w", err)
	}

	return &val, nil
}

// SetEmailSettings posts form fields to ++settings-email (partial update).
func (s *Server) SetEmailSettings(form url.Values) error {
	if err := s.PostForm("++settings-email", form); err != nil {
		return fmt.Errorf("setting email settings: %w", err)
	}

	return nil
}

// GetWebSettings fetches ++settings-web.
func (s *Server) GetWebSettings() (*WebSettings, error) {
	var val WebSettings
	if err := s.GetXML("++settings-web", nil, &val); err != nil {
		return nil, fmt.Errorf("getting web settings: %w", err)
	}

	return &val, nil
}

// SetWebSettings posts form fields to ++settings-web (partial update).
func (s *Server) SetWebSettings(form url.Values) error {
	if err := s.PostForm("++settings-web", form); err != nil {
		return fmt.Errorf("setting web settings: %w", err)
	}

	return nil
}

// GetCameraSettings fetches ++settings-cameras for a camera number.
func (s *Server) GetCameraSettings(cameraNum int) (*CameraSettings, error) {
	params := make(url.Values)
	params.Set("cameraNum", strconv.Itoa(cameraNum))

	var val CameraSettings
	if err := s.GetXML("++settings-cameras", params, &val); err != nil {
		return nil, fmt.Errorf("getting camera settings: %w", err)
	}

	return &val, nil
}

// SetCameraSettings posts form fields to ++settings-cameras (partial update).
// form must include cameraNum.
func (s *Server) SetCameraSettings(form url.Values) error {
	if form == nil || form.Get("cameraNum") == "" {
		return ErrCameraNumRequired
	}

	if err := s.PostForm("++settings-cameras", form); err != nil {
		return fmt.Errorf("setting camera settings: %w", err)
	}

	return nil
}
