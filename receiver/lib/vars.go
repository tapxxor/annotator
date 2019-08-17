package receiver

import "annotator/types"

// Configuration is a map tha holds application configuration
type Configuration map[string]string

// config variable for application configuration
var (
	Config      types.ReceiverConf
	ConfigFile  *string
	AlertsPath  string
	AlertsPaths map[string]string
)
