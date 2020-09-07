package proxy

import (
	"github.com/yeqown/log"
)

var (
	debug     = true
	logger, _ = log.NewLogger()
)

// SetProduction .
func SetProduction() {
	debug = false
	logger.SetLogLevel(log.LevelInfo)
}
