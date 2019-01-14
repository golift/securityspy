package securityspy

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

// Handle returns an iterface to interact with SecuritySpy.
func Handle(user, pass, url string, verifySSL bool) (SecuritySpy, error) {
	c := &concourse{
		EventBinds: make(map[EventName][]func(Event)),
		Config: &Config{
			BaseURL:   url,
			AuthB64:   base64.URLEncoding.EncodeToString([]byte(user + ":" + pass)),
			Username:  user,
			VerifySSL: verifySSL,
		},
	}
	return c, c.Refresh()
}

// Refresh gets fresh camera data from SecuritySpy, maybe run this after every action.
func (c *concourse) Refresh() error {
	if xmldata, err := c.secReqXML("/++systemInfo", nil); err != nil {
		return err
	} else if err := xml.Unmarshal(xmldata, c.SystemInfo); err != nil {
		return errors.Wrap(err, "xml.Unmarshal(++systemInfo)")
	}
	return nil
}

// ServerInfo returns the server name and version.
func (c *concourse) ServerInfo() Server {
	return c.SystemInfo.Server
}

// Scripts returns a list of scripts.
func (c *concourse) Scripts() ([]string, error) {
	type Scripts struct {
		XMLName xml.Name `xml:"scripts"`
		Names   []string `xml:"name"`
	}
	var s Scripts
	if xmldata, err := c.secReqXML("/++scripts", nil); err != nil {
		return nil, err
	} else if err := xml.Unmarshal(xmldata, s); err != nil {
		return nil, errors.Wrap(err, "xml.Unmarshal(++scripts)")
	}
	return s.Names, nil
}

// Sounds returns a list of sounds.
func (c *concourse) Sounds() ([]string, error) {
	type Sounds struct {
		XMLName xml.Name `xml:"sounds"`
		Names   []string `xml:"name"`
	}
	var s Sounds
	if xmldata, err := c.secReqXML("/++sounds", nil); err != nil {
		return nil, err
	} else if err := xml.Unmarshal(xmldata, s); err != nil {
		return nil, errors.Wrap(err, "xml.Unmarshal(++sounds)")
	}
	return s.Names, nil
}

/* INTERFACE HELPER METHODS FOLLOW */

// secReq is a helper function that formats the http request to SecuritySpy
func (c *concourse) secReq(apiPath string, params url.Values) (resp *http.Response, err error) {
	a := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: !c.Config.VerifySSL}}}
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
	resp, err := c.secReq(apiPath, params)
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
