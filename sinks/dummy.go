package sinks

import "github.com/distatus/battery"

type DummySink struct {
}

func (d *DummySink) Write(bat *battery.Battery) error {
	return nil
}

func (d *DummySink) Name() string {
	return "DummySink"
}
