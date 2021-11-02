package sinks

import (
	"github.com/distatus/battery"
	"github.com/sirupsen/logrus"
)

func init() {
	if err := registerSinkFactory("dummy", func(config map[string]interface{}) (Sinker, error) { return &DummySink{}, nil }); err != nil {
		logrus.WithError(err).Fatal("Failed registering log sink factory")
	}
}

type DummySink struct {
}

func (d *DummySink) Write(bat *battery.Battery) error {
	return nil
}

func (d *DummySink) Name() string {
	return "DummySink"
}
