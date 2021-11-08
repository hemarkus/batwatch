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
	"github.com/hemarkus/batwatch/source"
)

var snks = []sinks.Sinker{}
var bats = make(chan *source.Battery)

func init() {
	viper.SetDefault("battery", 0)
	viper.SetDefault("interval", 5)
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.batwatch")
	viper.AddConfigPath("/etc/batwatch/")
	err := viper.ReadInConfig()
	if err != nil {
		logrus.WithError(err).Fatal("Fatal error config file")
	}
	confSinks := viper.Get("sinks")
	confSinkMap := confSinks.([]interface{})

	for _, snkConf := range confSinkMap {
		name, ok := snkConf.(map[interface{}]interface{})["name"]
		if !ok {
			logrus.Fatal("Invalid sink config; missing name")
		}

		// convert config
		conf := make(map[string]interface{})
		for k, v := range snkConf.(map[interface{}]interface{}) {
			key, ok := k.(string)
			if !ok {
				logrus.Fatal("Invalid sink config; non string key")
			}
			conf[key] = v
		}

		snk, err := sinks.Get(name.(string), conf)
		if err != nil {
			logrus.WithError(err).Fatal("Failed initializing sink")
		}
		logrus.WithField("name", snk.Name()).WithField("sink", snk).Info("Sink initialized")
		snks = append(snks, snk)
	}
}

func main() {
	batNumber := viper.GetInt("battery")
	updateInterval := viper.GetDuration("interval")

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
				for _, sink := range snks {
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

	bats <- &source.Battery{Battery: bat}

	return nil
}
