package httphandler

import (
	"html/template"

	"github.com/dhcgn/iot-ephemeral-value-store/stats"
	"github.com/dhcgn/iot-ephemeral-value-store/storage"
)

type Config struct {
	StorageInstance  storage.Storage
	StatsInstance    *stats.Stats
	DownloadTemplate *template.Template
}
