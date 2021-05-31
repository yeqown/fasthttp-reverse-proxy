package proxy

import (
	"github.com/yeqown/log"
)

var (
	logger, _ = log.NewLogger()
)

// SetProduction .
func SetProduction() {
	logger.SetLogLevel(log.LevelInfo)
}
