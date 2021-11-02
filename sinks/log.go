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
			return &LogSink{Severity: severity}, nil
		},
	); err != nil {
		logrus.WithError(err).Fatal("Failed registering log sink factory")
	}
}

var severityFuncMap = map[string]func(args ...interface{}){
	"info":  logrus.Info,
	"debug": logrus.Debug,
}

type LogSink struct {
	Severity string
}

func (d *LogSink) Write(bat *battery.Battery) error {
	f, ok := severityFuncMap[d.Severity]
	if !ok {
		return fmt.Errorf("unsupported severity %v", d.Severity)
	}

	percentage := bat.Current * 100 / bat.Design

	f(fmt.Sprintf("Battery is %v at %f %%", bat.State, percentage))

	return nil
}

func (d *LogSink) Name() string {
	return "LogSink"
}
