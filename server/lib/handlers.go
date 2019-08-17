package lib

import (
	"log"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ServeMetrics the function handler that exposes golang metrics
func ServeMetrics() {
	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Start exposing prometheus metrics on port %d\n", Config.Server.Settings.Port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+strconv.Itoa(int(Config.Server.Settings.Port)), nil))
}
