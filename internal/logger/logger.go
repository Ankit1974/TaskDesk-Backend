// Package logger provides a global structured logger using Uber's Zap library.
// It supports environment-aware configuration: development mode enables
// human-readable output with debug level, while production mode uses
// JSON-formatted output optimized for log aggregation services.
//
// Usage from any package: logger.Log.Info("message"), logger.Log.Error("err"), etc.
package logger

import (
	"go.uber.org/zap"
)

// Log is the global logger instance accessible from anywhere as logger.Log.
// Must be initialized via InitLogger() before use.
var Log *zap.Logger

// InitLogger creates and configures the global logger based on the environment.
//   - "production": JSON output, info level, optimized for performance.
//   - any other value (e.g., "development"): console output, debug level, with stack traces.
func InitLogger(env string) {
	var config zap.Config
	if env == "production" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}

	var err error
	Log, err = config.Build()
	if err != nil {
		panic(err)
	}
}
