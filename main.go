package main

import (
	"time"

	"github.com/distatus/battery"
	"github.com/sirupsen/logrus"

	"github.com/hemarkus/batwatch/sinks"
)

const updateInterval time.Duration = 1

var sinksReg = []sinks.Sinker{}
var batNumber = 0
var bats = make(chan *battery.Battery)

func init() {
	sinksReg = append(sinksReg, &sinks.DummySink{}, &sinks.LogSink{Severity: "info"})
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
				for _, sink := range sinksReg {
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
