package types

import (
	"encoding/json"
	"log"
	"net/http"
)

// AnnotationsResponse is a list of Annotation objects.
type AnnotationsResponse []Annotation

// Annotation item
type Annotation struct {
	ID             int64    `json:"id"`
	AlertID        int64    `json:"alertId"`
	AlertName      string   `json:"alertName"`
	DashboardID    int64    `json:"dashboardId"`
	PanelID        int64    `json:"panelId"`
	UserID         int64    `json:"userId"`
	NewState       string   `json:"newState"`
	PrevState      string   `json:"prevState"`
	Created        int64    `json:"created"`
	Updated        int64    `json:"updated"`
	Time           int64    `json:"time"`
	Text           string   `json:"text"`
	RegionID       int64    `json:"regionId"`
	Tags           Tags     `json:"tags"`
	Login          string   `json:"login"`
	Email          string   `json:"email"`
	AvatarURL      string   `json:"avatarUrl"`
	AnnotationData JSONData `json:"data"`
}

// JSONData can match any type of data
type JSONData struct {
	data interface{}
}

// Tags is slice of string representing tags
type Tags []string

// Load gets annotations from grafana with url u , using the APIKey k
func (r *AnnotationsResponse) Load(u string, k string) (err error) {

	req, err := http.NewRequest("GET", u, nil)
	req.Header.Add("Authorization", "Bearer "+k)

	// Send req using http Client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error in GET %s: %v", u, err)
	}

	// Decode the JSON in the body
	d := json.NewDecoder(resp.Body)
	defer resp.Body.Close()

	err = d.Decode(r)
	if err != nil {
		log.Fatalf("Error in decoding grafana annotations response: %v", err)
	}

	d = nil
	return
}

// AnnotationsPost is the paylod used to post an annotation
type AnnotationsPost struct {
	Time     int64  `json:"time"`
	IsRegion bool   `json:"isRegion"`
	TimeEnd  int64  `json:"timeEnd"`
	Tags     Tags   `json:"tags"`
	Text     string `json:"text"`
}

// AnnotationsPatch is the paylod used to patch an annotation
type AnnotationsPatch struct {
	Time int64 `json:"time"`
}

// AnnotationsMethodResponse is the payload returned after posting/patching an annotation
type AnnotationsMethodResponse struct {
	Message string `json:"message"`
	ID      int64  `json:"id,omitempty"`
	EndID   int64  `json:"endId,omitempty"`
}
