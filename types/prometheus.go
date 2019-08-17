package types

import (
	"encoding/json"
	"fmt"
	"time"
)

// KV is a set of key/value string pairs.
type KV map[string]string

// AlertState is used as part of AlertStatus.
type AlertState string

// Message holds the JSON object sent from the alertmanager
type Message struct {
	*Data
	// The protocol version
	Version  string `json:"version"`
	GroupKey string `json:"groupKey"`
}

// Data is the data passed to notification templates and webhook pushes.
type Data struct {
	Receiver          string `json:"receiver"`
	Status            string `json:"status"`
	Alerts            Alerts `json:"alerts"`
	GroupLabels       KV     `json:"groupLabels"`
	CommonLabels      KV     `json:"commonLabels"`
	CommonAnnotations KV     `json:"commonAnnotations"`
	ExternalURL       string `json:"externalURL"`
}

// Alerts is a list of Alert objects.
type Alerts []Alert

// Alert holds one alert for notification templates.
type Alert struct {
	Status       string    `json:"status"`
	Labels       KV        `json:"labels"`
	Annotations  KV        `json:"annotations"`
	StartsAt     time.Time `json:"startsAt"`
	EndsAt       time.Time `json:"endsAt"`
	GeneratorURL string    `json:"generatorURL"`
	Fingerprint  string    `json:"fingerprint,omitempty"`
}

// Possible values for AlertState.
const (
	AlertStateFiring   AlertState = "firing"
	AlertStateResolved AlertState = "resolved"
	AlertStatepending  AlertState = "pending"
)

// String returns a pretty formatted Json string
func (m *Message) String() string {
	data, _ := json.MarshalIndent(m, "", "  ")
	return fmt.Sprintf("%s", data)
}
