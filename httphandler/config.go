package httphandler

import (
	"html/template"

	"github.com/dhcgn/iot-ephemeral-value-store/data"
	"github.com/dhcgn/iot-ephemeral-value-store/stats"
)

type Config struct {
	DataService      *data.Service
	StatsInstance    *stats.Stats
	DownloadTemplate *template.Template
}
