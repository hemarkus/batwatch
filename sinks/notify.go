package sinks

import (
	"fmt"

	"github.com/ctcpip/notifize"
	"github.com/distatus/battery"
	"github.com/sirupsen/logrus"

	"github.com/hemarkus/batwatch/source"
)

func init() {
	if err := registerSinkFactory("notify",
		func(config map[string]interface{}) (Sinker, error) {
			n := NotificationSink{}

			surg, ok := config["state_urgent"]
			if ok {
				n.stateUrgent, ok = surg.(bool)
				if !ok {
					return nil, fmt.Errorf("invalid notify sink config; invalid state_urgent")
				}
			}

			lt, ok := config["low_level"]
			if ok {
				n.lowThreshold, ok = lt.(float64)
				if !ok {
					return nil, fmt.Errorf("invalid notify sink config; invalid low_level")
				}
			}

			lurg, ok := config["low_urgent"]
			if ok {
				n.lowUrgent, ok = lurg.(bool)
				if !ok {
					return nil, fmt.Errorf("invalid notify sink config; invalid low_urgent")
				}
			}

			ct, ok := config["critical_level"]
			if ok {
				n.critThreshold, ok = ct.(float64)
				if !ok {
					return nil, fmt.Errorf("invalid notify sink config; invalid critical_level")
				}
			}

			curg, ok := config["critical_urgent"]
			if ok {
				n.critUrgent, ok = curg.(bool)
				if !ok {
					return nil, fmt.Errorf("invalid notify sink config; invalid critical_urgent")
				}
			}

			return &n, nil

		},
	); err != nil {
		logrus.WithError(err).Fatal("Failed registering notify sink factory")
	}
}

type NotificationSink struct {
	lastState     battery.State
	lastWarn      int
	stateUrgent   bool
	lowUrgent     bool
	critUrgent    bool
	lowThreshold  float64
	critThreshold float64
}

func (n *NotificationSink) Write(bat *source.Battery) error {
	percentage := bat.Percentage()

	if percentage <= n.critThreshold && n.lastWarn < 2 {
		notifize.Display("Battery level critical", fmt.Sprintf("Battery is at %.2f", percentage), n.critUrgent, "")
		n.lastWarn = 2
		return nil
	}

	if percentage <= n.lowThreshold && n.lastWarn < 1 {
		notifize.Display("Battery level low", fmt.Sprintf("Battery is at %.2f", percentage), n.lowUrgent, "")
		n.lastWarn = 1
		return nil
	}

	if n.lastState == bat.State {
		return nil
	}
	n.lastState = bat.State
	notifize.Display("New battery status", fmt.Sprintf("Battery is %s", bat.State), n.stateUrgent, "")

	if bat.State == battery.Full || bat.State == battery.Charging {
		n.lastWarn = 0
	}

	return nil
}

func (n *NotificationSink) Name() string {
	return "NotificationSink"
}
