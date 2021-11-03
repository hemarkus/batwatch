package sinks

import (
	"fmt"

	"github.com/ctcpip/notifize"
	"github.com/distatus/battery"
	"github.com/sirupsen/logrus"
)

func init() {
	if err := registerSinkFactory("notify",
		func(config map[string]interface{}) (Sinker, error) {
			urg, ok := config["urgent"]
			if !ok {
				return nil, fmt.Errorf("invalid notify sink config; missing urgent")
			}
			urgent, ok := urg.(bool)
			if !ok {
				return nil, fmt.Errorf("invalid notify sink config; invalid urgent")
			}

			return &NotificationSink{urgent: urgent}, nil
		},
	); err != nil {
		logrus.WithError(err).Fatal("Failed registering notify sink factory")
	}
}

type NotificationSink struct {
	lastState battery.State
	urgent    bool
}

func (n *NotificationSink) Write(bat *battery.Battery) error {
	if n.lastState == bat.State {
		return nil
	}
	n.lastState = bat.State
	notifize.Display("New battery status", fmt.Sprintf("Battery is %s", bat.State), n.urgent, "")
	return nil
}

func (n *NotificationSink) Name() string {
	return "NotificationSink"
}
