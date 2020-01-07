package schedule

import (
	"encoding/json"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/mindworks-software/energia/pkg/axpert"
	"github.com/mindworks-software/energia/pkg/connector"
)

type Entry struct {
	Timestamp    time.Time
	OutputSource axpert.OutputSourcePriority
}

type Schedule struct {
	DefaultOutputSource axpert.OutputSourcePriority
	Entries             []Entry
}

const (
	ScheduleTopic = "inverter/cmd/schedule"
	OverrideTopic = "inverter/cmd/override"
	EnableTopic   = "inverter/cmd/enableSchedule"
)

func CreateSchedule(msg mqtt.Message, ucc chan connector.Connector) (*Schedule, error) {
	var s Schedule
	err := json.Unmarshal(msg.Payload(), &s)
	if err != nil {
		return nil, err
	}

	// Set output to current from schedule
	// create  timer for now + (time to next clock 10 min interval)
	// when timer ticks, start ticker for 10 min intervals

	return &s, nil
}
