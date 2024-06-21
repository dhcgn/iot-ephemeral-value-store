package httphandler

import (
	"github.com/dhcgn/iot-ephemeral-value-store/stats"
	"github.com/dhcgn/iot-ephemeral-value-store/storage"
)

type Config struct {
	StorageInstance storage.Storage
	StatsInstance   *stats.Stats
}
