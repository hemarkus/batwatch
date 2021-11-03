package sinks

import (
	"fmt"

	"github.com/distatus/battery"
	"github.com/sirupsen/logrus"
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

func (d *LogSink) Write(bat *battery.Battery) error {
	f, ok := logLevelMap[d.LogLevel]
	if !ok {
		return fmt.Errorf("unsupported severity %v", d.LogLevel)
	}

	percentage := bat.Current * 100 / bat.Design

	logrus.WithFields(logrus.Fields{
		"state":      bat.State,
		"percentage": fmt.Sprintf("%.2f", percentage),
		"current":    bat.Current,
		"design":     bat.Design,
		"full":       bat.Full,
	}).Log(f, "Battery status")

	return nil
}

func (d *LogSink) Name() string {
	return "LogSink"
}
