package lib

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

const (
	receiverDefaultPort int32 = 5000
	serverDefaultPort   int32 = 5001
)

// ServerConf is the server block
type ServerConf struct {
	Server ServerSettings `yaml:"server"`
}

// ServerSettings has server settings and annotation settings configuration
type ServerSettings struct {
	Settings    ServerSettingsData       `yaml:"settings"`
	Annotations []ServerAnnotationConfig `yaml:"annotations"`
}

// ServerSettingsData All possible server settings
type ServerSettingsData struct {
	Metrics    bool   `yaml:"metrics"`
	Port       int32  `yaml:"port"`
	Path       string `yaml:"path"`
	GrafanaURL string `yaml:"grafanaURL"`
	Apikey     string `yaml:"apiKey"`
	AlertsPath string `yaml:"alertsPath"`
	SqliteHome string `yaml:"sqliteHome"`
}

// ServerAnnotationConfig configuration struct for alert
type ServerAnnotationConfig struct {
	Name string   `yaml:"name"`
	Tags []string `yaml:"tags"`
}

// ReceiverConf is the server block
type ReceiverConf struct {
	Receiver ReceiverSettings `yaml:"receiver"`
}

// ReceiverSettings has receiver settings and annotation settings configuration
type ReceiverSettings struct {
	Settings ReceiverSettingsData `yaml:"settings"`
}

// ReceiverSettingsData All possible receiver settings
type ReceiverSettingsData struct {
	Metrics     bool   `yaml:"metrics"`
	Port        int32  `yaml:"port"`
	MetricsPath string `yaml:"path"`
	AlertsPath  string `yaml:"data"`
}

// Fread reads a server configuration from confPath
func (c *ServerConf) Fread(confPath *string) (err error) {
	yamlFile, err := ioutil.ReadFile(*confPath)
	err = yaml.Unmarshal(yamlFile, c)

	return
}

// Fread reads a receiver configuration from confPath
func (c *ReceiverConf) Fread(confPath *string) (err error) {
	yamlFile, err := ioutil.ReadFile(*confPath)
	err = yaml.Unmarshal(yamlFile, c)

	return
}

// Validate validates server yaml values
func (c *ServerConf) Validate() (err error) {

	if c.Server.Settings.Port == 0 {
		c.Server.Settings.Port = serverDefaultPort
	}

	if c.Server.Settings.Path == "" {
		err = errors.New("path is empty")
	}

	if c.Server.Settings.GrafanaURL == "" {
		err = errors.New("grafanaURL is empty")
	}

	if c.Server.Settings.Apikey == "" {
		err = errors.New("grafanaURL is empty")
	}

	_, err = os.Stat(c.Server.Settings.AlertsPath)

	return
}

// Validate validates server yaml values
func (c *ReceiverConf) Validate() (err error) {

	if c.Receiver.Settings.Port == 0 {
		c.Receiver.Settings.Port = receiverDefaultPort
	}

	if c.Receiver.Settings.MetricsPath == "" {
		err = errors.New("path is empty")
	}

	if c.Receiver.Settings.AlertsPath == "" {
		err = errors.New("data is empty")
	}

	return
}

// Region describes a region annotation in grafana
type Region struct {
	ID       int64
	StartID  int64
	StartsAt int64
	EndID    int64
	EndsAt   int64
}

// Regions is a map of regions on region ID
type Regions map[int64]*Region

// Load consturcts the map[int64]Region based on AnnotationsResponse contents
func (r *Regions) Load(a *AnnotationsResponse) (err error) {

	if *r == nil {
		*r = make(map[int64]*Region)
	}

	for _, ann := range *a {
		// fmt.Printf("regionId %d: id %d\n", ann.RegionID, ann.ID)
		if _, ok := (*r)[ann.RegionID]; !ok {
			(*r)[ann.RegionID] = &Region{ID: ann.RegionID,
				StartID:  ann.ID,
				StartsAt: ann.Time,
			}
		} else {
			// regionID exists fix start and end fields
			if ann.Time > (*r)[ann.RegionID].StartsAt {
				// this annotations is the end
				(*r)[ann.RegionID].EndID = ann.ID
				(*r)[ann.RegionID].EndsAt = ann.Time
			} else {
				// this annotation is the start
				(*r)[ann.RegionID].EndID, (*r)[ann.RegionID].StartID = (*r)[ann.RegionID].StartID, ann.ID
				(*r)[ann.RegionID].EndsAt, (*r)[ann.RegionID].StartsAt = (*r)[ann.RegionID].StartsAt, ann.Time
			}
		}
	}
	return
}

// String method prints Regions in pretty format
func (r *Regions) String() string {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for k, v := range *r {
		buf.WriteByte('\n')
		buf.WriteString(fmt.Sprintf("\t{ Region %d: [Start: %d, End: %d] }", k, v.StartID, v.EndID))
	}
	buf.WriteByte('\n')
	buf.WriteByte('}')

	return buf.String()
}
