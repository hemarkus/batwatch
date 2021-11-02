package main

import (
	"fmt"
	"time"

	"github.com/distatus/battery"
	"github.com/sirupsen/logrus"
)

type Sinker interface {
	Name() string
	Write(bat *battery.Battery) error
}

const updateInterval time.Duration = 1

var sinks = []Sinker{}
var batNumber = 0
var bats = make(chan *battery.Battery)

type DummySink struct {
}

func (d *DummySink) Write(bat *battery.Battery) error {
	return nil
}

func (d *DummySink) Name() string {
	return "DummySink"
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

	f(fmt.Sprintf("Battery state is %v", bat.State))

	return nil
}

func (d *LogSink) Name() string {
	return "LogSink"
}

func init() {
	sinks = append(sinks, &DummySink{}, &LogSink{Severity: "info"})
}

func main() {
	ticker := time.NewTicker(updateInterval * time.Second)
	done := make(chan struct{})

	// kick off source
	go func() {
		for {
			select {
			case <-done:
				done <- struct{}{}
				return
			case <-ticker.C:
				err := updateBatteries()
				if err != nil {
					logrus.WithError(err).Error("Failed sourcing battery state")
				}
			}
		}
	}()

	// kick off sink
	go func() {
		for {
			select {
			case <-done:
				done <- struct{}{}
				return
			case b := <-bats:
				for _, sink := range sinks {
					err := sink.Write(b)
					if err != nil {
						logrus.WithError(err).WithField("sink", sink.Name()).Error("Failed writing to sink")
					}
				}
			}
		}
	}()

	<-done
}

func updateBatteries() error {
	bat, err := battery.Get(batNumber)
	if err != nil {
		logrus.WithError(err).Error("Could not get battery info")
		return err
	}

	bats <- bat

	return nil
}
