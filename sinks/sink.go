package sinks

import (
	"fmt"

	"github.com/distatus/battery"
)

type Sinker interface {
	Name() string
	Write(bat *battery.Battery) error
}

type sinkFactory func(config map[string]interface{}) (Sinker, error)

var sinkFactories = map[string]sinkFactory{}

func registerSinkFactory(name string, snkf sinkFactory) error {
	if _, ok := sinkFactories[name]; ok {
		return fmt.Errorf("duplicate sink factory [%s]", name)
	}

	sinkFactories[name] = snkf

	return nil
}

func Get(name string, config map[string]interface{}) (Sinker, error) {
	sinkF, ok := sinkFactories[name]
	if !ok {
		return nil, fmt.Errorf("no such sink factory %s", name)
	}
	return sinkF(config)
}
