package securityspy

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// GetServer returns an iterface to interact with SecuritySpy.
func GetServer(user, pass, url string, verifySSL bool) (Server, error) {
	c := &concourse{
		SystemInfo: new(systemInfo),
		EventBinds: make(map[EventName][]func(Event)),
		Config: &config{
			BaseURL:   url,
			AuthB64:   base64.URLEncoding.EncodeToString([]byte(user + ":" + pass)),
			Username:  user,
			VerifySSL: verifySSL,
		},
	}
	if err := c.Refresh(); err != nil {
		return c, err
	} else if err := c.RefreshScripts(); err != nil {
		return c, err
	}
	return c, c.RefreshSounds()
}

// Refresh gets fresh camera data from SecuritySpy, maybe run this after every action.
func (c *concourse) Refresh() error {
	if xmldata, err := c.secReqXML("/++systemInfo", nil); err != nil {
		return err
	} else if err := xml.Unmarshal(xmldata, c.SystemInfo); err != nil {
		return errors.Wrap(err, "xml.Unmarshal(++systemInfo)")
	}
	// Add the name to each assigned Camera Schedule.
	for i, cam := range c.SystemInfo.CameraList.Cameras {
		c.SystemInfo.CameraList.Cameras[i].ScheduleIDA.Name = strings.TrimSpace(c.getScheduleName(cam.ScheduleIDA.ID))
		c.SystemInfo.CameraList.Cameras[i].ScheduleIDCC.Name = strings.TrimSpace(c.getScheduleName(cam.ScheduleIDCC.ID))
		c.SystemInfo.CameraList.Cameras[i].ScheduleIDMC.Name = strings.TrimSpace(c.getScheduleName(cam.ScheduleIDMC.ID))
		c.SystemInfo.CameraList.Cameras[i].ScheduleOverrideA.Name = strings.TrimSpace(c.getScheduleName(cam.ScheduleOverrideA.ID))
		c.SystemInfo.CameraList.Cameras[i].ScheduleOverrideCC.Name = strings.TrimSpace(c.getScheduleName(cam.ScheduleOverrideCC.ID))
		c.SystemInfo.CameraList.Cameras[i].ScheduleOverrideMC.Name = strings.TrimSpace(c.getScheduleName(cam.ScheduleOverrideMC.ID))
	}
	return nil
}

// Info returns the server name and version.
func (c *concourse) Info() ServerInfo {
	return c.SystemInfo.Server
}

// RefreshScripts refreshes the list of script files. Probably doesn't change much.
// Retreivable as server.Info().Scripts.Names
func (c *concourse) RefreshScripts() error {
	if xmldata, err := c.secReqXML("/++scripts", nil); err != nil {
		return err
	} else if err := xml.Unmarshal(xmldata, &c.SystemInfo.Server.Scripts); err != nil {
		return errors.Wrap(err, "xml.Unmarshal(++scripts)")
	}
	return nil
}

// RefreshSounds refreshes the list of sound files. Probably doesn't change much.
// Retreivable as server.Info().Sounds.Names
func (c *concourse) RefreshSounds() error {
	if xmldata, err := c.secReqXML("/++sounds", nil); err != nil {
		return err
	} else if err := xml.Unmarshal(xmldata, &c.SystemInfo.Server.Sounds); err != nil {
		return errors.Wrap(err, "xml.Unmarshal(++sounds)")
	}
	return nil
}

/* INTERFACE HELPER METHODS FOLLOW */

// secReq is a helper function that formats the http request to SecuritySpy
func (c *concourse) secReq(apiPath string, params url.Values, timeout time.Duration) (resp *http.Response, err error) {
	if params == nil {
		params = make(url.Values)
	}
	a := &http.Client{Timeout: timeout, Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: !c.Config.VerifySSL}}}
	req, err := http.NewRequest("GET", c.Config.BaseURL+apiPath, nil)
	if err != nil {
		return resp, errors.Wrap(err, "http.NewRequest()")
	}
	params.Set("auth", c.Config.AuthB64)
	params.Set("format", "xml")
	req.URL.RawQuery = params.Encode()
	req.Header.Add("Accept", "application/xml")
	resp, err = a.Do(req)
	if err != nil {
		return resp, errors.Wrap(err, "http.Do(req)")
	}
	return resp, nil
}

func (c *concourse) secReqXML(apiPath string, params url.Values) (xmldata []byte, err error) {
	resp, err := c.secReq(apiPath, params, 15*time.Second)
	if err != nil {
		return xmldata, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return xmldata, errors.Errorf("authentication failed (%v): %v (status: %v/%v)",
			c.Config.Username, c.Config.BaseURL+apiPath, resp.StatusCode, resp.Status)
	} else if xmldata, err = ioutil.ReadAll(resp.Body); err == nil {
		return xmldata, errors.Wrap(err, "ioutil.ReadAll(resp.Body)")
	}
	return xmldata, nil
}

func (c *concourse) getScheduleName(id int) string {
	for _, schedule := range c.SystemInfo.ScheduleList.Schedules {
		if schedule.ID == id {
			return schedule.Name
		}
	}
	return "Unknown Schedule"
}
