package main

import (
	"os"
	"os/signal"
	"sync"
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

	var wg sync.WaitGroup

	// kick off source
	wg.Add(1)
	go func() {
		defer wg.Done()
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
	wg.Add(1)
	go func() {
		defer wg.Done()
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

	// Wait for interrupt
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	<-quit
	logrus.Info("Shutdown")
	done <- struct{}{}
	wg.Wait()
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
