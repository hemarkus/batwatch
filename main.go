package main

import (
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/distatus/battery"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/hemarkus/batwatch/sinks"
)

var sinksReg = []sinks.Sinker{}
var bats = make(chan *battery.Battery)

func init() {
	sinksReg = append(sinksReg, &sinks.DummySink{}, &sinks.LogSink{Severity: "info"})
}

func main() {
	viper.SetDefault("battery", 0)
	viper.SetDefault("interval", 5)
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/batwatch/")
	viper.AddConfigPath("$HOME/.batwatch")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		logrus.WithError(err).Fatal("Fatal error config file")
	}

	batNumber := viper.GetInt("battery")
	updateInterval := viper.GetInt("interval")

	ticker := time.NewTicker(time.Duration(updateInterval) * time.Second)
	done := make(chan struct{})

	var wg sync.WaitGroup

	// kick off source
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				logrus.Info("Source shutdown")
				return
			case <-ticker.C:
				err := updateBattery(batNumber)
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
				logrus.Info("Sinks shutdown")
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
	close(done)
	wg.Wait()
	close(bats)
}

func updateBattery(batteryno int) error {
	bat, err := battery.Get(batteryno)
	if err != nil {
		logrus.WithError(err).Error("Could not get battery info")
		return err
	}

	bats <- bat

	return nil
}
