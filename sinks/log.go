package sinks

import (
	"fmt"

	"github.com/distatus/battery"
	"github.com/sirupsen/logrus"
)

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
