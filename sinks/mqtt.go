package sinks

import (
	"encoding/json"
	"fmt"

	"github.com/distatus/battery"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
)

func init() {
	if err := registerSinkFactory("mqtt",
		func(config map[string]interface{}) (Sinker, error) {
			brok, ok := config["broker"]
			if !ok {
				return nil, fmt.Errorf("invalid mqtt sink config; missing broker")
			}
			broker, ok := brok.(string)
			if !ok {
				return nil, fmt.Errorf("invalid mqtt sink config; invalid broker")
			}

			port, ok := config["port"]
			if !ok {
				return nil, fmt.Errorf("invalid mqtt sink config; missing port")
			}
			brokerPort, ok := port.(int)
			if !ok {
				return nil, fmt.Errorf("invalid mqtt sink config; invalid port")
			}

			top, ok := config["base_topic"]
			if !ok {
				return nil, fmt.Errorf("invalid mqtt sink config; missing base_topic")
			}
			topic, ok := top.(string)
			if !ok {
				return nil, fmt.Errorf("invalid mqtt sink config; invalid base_topic")
			}

			var configPrefix string
			if configPref, ok := config["config_prefix"]; ok {
				configPrefix, ok = configPref.(string)
				if !ok {
					return nil, fmt.Errorf("invalid mqtt sink config; invalid config_prefix")
				}
			}

			mqttSnk := MQTTSink{baseTopic: topic, broker: broker, port: brokerPort, configPrefix: configPrefix}
			opts := mqtt.NewClientOptions()
			opts.AddBroker(fmt.Sprintf("tcp://%s:%d", mqttSnk.broker, mqttSnk.port))
			opts.SetClientID("batwatch")
			mqttSnk.client = mqtt.NewClient(opts)
			return &mqttSnk, nil
		},
	); err != nil {
		logrus.WithError(err).Fatal("Failed registering mqtt sink factory")
	}
}

type MQTTSink struct {
	broker       string
	port         int
	baseTopic    string
	configPrefix string
	client       mqtt.Client
}

func (m *MQTTSink) Connect() error {
	if token := m.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	if m.configPrefix != "" {
		if err := m.sendConfig(); err != nil {
			logrus.WithError(err).Error("Failed sending config")
		}
	}
	return nil
}

type Config struct {
	Name        string `json:"name"`
	StateTopic  string `json:"state_topic"`
	UniqueID    string `json:"uniq_id"`
	DeviceClass string `json:"device_class,omitempty"`
}

func (m *MQTTSink) sendConfig() error {
	stateTopic := fmt.Sprintf("%s/%s", m.baseTopic, "state")
	batTopic := fmt.Sprintf("%s/%s", m.baseTopic, "percentage")
	confs := []Config{
		Config{
			Name:       "batwatch state",
			StateTopic: stateTopic,
			UniqueID:   "batwatch_state",
		},
		Config{
			Name:        "batwatch percent",
			StateTopic:  batTopic,
			UniqueID:    "batwatch_percent",
			DeviceClass: "battery",
		}}

	for _, c := range confs {
		payload, err := json.Marshal(c)
		if err != nil {
			logrus.WithError(err).Error("Failed marshalling config")
			return err
		}
		token := m.client.Publish(fmt.Sprintf("%s/sensor/%s/config", m.configPrefix, c.UniqueID), 0, true, payload)
		token.Wait()
		if token.Error() != nil {
			logrus.WithError(err).Error("Failed sending config")
			return token.Error()
		}
	}
	return nil
}

func (m *MQTTSink) Write(bat *battery.Battery) error {
	if !m.client.IsConnected() {
		if err := m.Connect(); err != nil {
			return fmt.Errorf("connection failed: %w", err)
		}
	}
	stateTopic := fmt.Sprintf("%s/%s", m.baseTopic, "state")
	token := m.client.Publish(stateTopic, 0, false, bat.State.String())
	token.Wait()
	if token.Error() != nil {
		return token.Error()
	}

	percentTopic := fmt.Sprintf("%s/%s", m.baseTopic, "percentage")
	percentage := bat.Current * 100 / bat.Design
	token = m.client.Publish(percentTopic, 0, false, fmt.Sprintf("%.2f", percentage))
	token.Wait()
	return token.Error()
}

func (m *MQTTSink) Name() string {
	return "MQTTSink"
}
