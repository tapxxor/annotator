package lib

import (
	"annotator/types"
	"database/sql"
	"sync"
)

const (
	// AnnotationsAPI the end point of annotations api in grafana
	AnnotationsAPI string = "/api/annotations"
)

var (
	// Config the servers struct for configuration
	Config types.ServerConf
	// GrafanaAPIAnnotations the struct for annotations response
	GrafanaAPIAnnotations types.AnnotationsResponse
	// Regions struct with ids for starting and ending annotation
	Regions types.Regions
	// ConfigFile the path of the configuration file
	ConfigFile *string
	// GrafanaAnnotationsURL absolute url of grafana annotations api
	GrafanaAnnotationsURL string
	// SqliteHome the folder of the Sqlite DB file
	SqliteHome string
	// SqliteDB the name of the sqlite db file
	SqliteDB string
	// AlertsPath the directory where alerts are stored from receiver
	AlertsPath string
	// AlertsFiringPath the absolute path of firing alerts
	AlertsFiringPath string
	// AlertsResolvedPath the absolute path of resolved alerts
	AlertsResolvedPath string
	// AlertsPaths a map containing AlertsFiringPath and AlertsResolvedPath
	AlertsPaths map[string]string
	// C the aqlite db connection
	C *sql.DB
	// Mu the database mutex
	Mu sync.Mutex
	// Ch the channel used for getting messages from routines
	Ch = make(chan string)
)
