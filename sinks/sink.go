package sinks

import "github.com/distatus/battery"

type Sinker interface {
	Name() string
	Write(bat *battery.Battery) error
}
