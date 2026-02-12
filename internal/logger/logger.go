package logger

import (
	"go.uber.org/zap"
)

var Log *zap.Logger

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
