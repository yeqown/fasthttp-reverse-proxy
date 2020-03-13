package proxy

import (
	"github.com/yeqown/log"
)

var (
	debug  bool = true
	logger      = log.NewLogger()
)

// SetProduction .
func SetProduction() {
	debug = false
	logger.SetLogLevel(log.LevelInfo)
}
