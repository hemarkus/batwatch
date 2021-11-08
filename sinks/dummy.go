package sinks

import (
	"github.com/sirupsen/logrus"

	"github.com/hemarkus/batwatch/source"
)

func init() {
	if err := registerSinkFactory("dummy", func(config map[string]interface{}) (Sinker, error) { return &DummySink{}, nil }); err != nil {
		logrus.WithError(err).Fatal("Failed registering log sink factory")
	}
}

type DummySink struct {
}

func (d *DummySink) Write(bat *source.Battery) error {
	return nil
}

func (d *DummySink) Name() string {
	return "DummySink"
}
