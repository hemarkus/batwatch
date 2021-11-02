package sinks

import (
	"fmt"
	"strconv"

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

			top, ok := config["topic"]
			if !ok {
				return nil, fmt.Errorf("invalid mqtt sink config; missing topic")
			}
			topic, ok := top.(string)
			if !ok {
				return nil, fmt.Errorf("invalid mqtt sink config; invalid topic")
			}

			return &MQTTSink{topic: topic, broker: broker, port: brokerPort}, nil
		},
	); err != nil {
		logrus.WithError(err).Fatal("Failed registering mqtt sink factory")
	}
}

type MQTTSink struct {
	broker string
	port   int
	topic  string
}

func (m *MQTTSink) Write(bat *battery.Battery) error {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", m.broker, m.port))
	opts.SetClientID("batwatch")
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	stateTopic := fmt.Sprintf("%s/%s", m.topic, "state")
	token := client.Publish(stateTopic, 0, false, bat.State.String())
	token.Wait()
	if token.Error() != nil {
		return token.Error()
	}

	percentTopic := fmt.Sprintf("%s/%s", m.topic, "percentage")
	percentage := bat.Current * 100 / bat.Design
	token = client.Publish(percentTopic, 0, false, strconv.FormatFloat(percentage, 'f', 5, 64))
	token.Wait()
	return token.Error()
}

func (m *MQTTSink) Name() string {
	return "MQTTSink"
}
