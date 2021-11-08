package sinks

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/hemarkus/batwatch/source"
)

func init() {
	if err := registerSinkFactory("log",
		func(config map[string]interface{}) (Sinker, error) {
			sev, ok := config["severity"]
			if !ok {
				return nil, fmt.Errorf("invalid log sink config; missing severity")
			}
			severity := sev.(string)
			return &LogSink{LogLevel: severity}, nil
		},
	); err != nil {
		logrus.WithError(err).Fatal("Failed registering log sink factory")
	}
}

var logLevelMap = map[string]logrus.Level{
	"info":  logrus.InfoLevel,
	"debug": logrus.DebugLevel,
}

type LogSink struct {
	LogLevel string
}

func (d *LogSink) Write(bat *source.Battery) error {
	f, ok := logLevelMap[d.LogLevel]
	if !ok {
		return fmt.Errorf("unsupported severity %v", d.LogLevel)
	}

	logrus.WithFields(logrus.Fields{
		"state":      bat.State,
		"percentage": fmt.Sprintf("%.2f", bat.Percentage()),
		"current":    bat.Current,
		"design":     bat.Design,
		"full":       bat.Full,
	}).Log(f, "Battery status")

	return nil
}

func (d *LogSink) Name() string {
	return "LogSink"
}
