package schedule

import (
	"encoding/json"
	"fmt"
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

	tickInterval = 10
)

var sch *Schedule
var ucc chan connector.Connector

func CreateSchedule(msg mqtt.Message, ucchan chan connector.Connector) (*Schedule, error) {
	s, err := umarshalSchedule(msg.Payload())
	if err != nil {
		return nil, err
	}

	sch = s
	ucc = ucchan
	// Set output to current from schedule
	err = setToCurrent(sch, ucc)
	if err != nil {
		return nil, err
	}

	// create  timer for now + (time to next clock 10 min interval)
	// when timer ticks, start ticker for 10 min intervals
	time.AfterFunc(calculateOffset(), startTicker)

	return sch, nil
}

func umarshalSchedule(msg []byte) (*Schedule, error) {
	var s Schedule
	err := json.Unmarshal(msg, &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func startTicker() {
	fmt.Println("Starting ticker")
	tck := time.NewTicker(tickInterval * time.Minute)
	go func(sch *Schedule, uchan chan connector.Connector) {
		for range tck.C {
			fmt.Println("Tick")
			err := setToCurrent(sch, uchan)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("Tock")
		}

	}(sch, ucc)
}

func calculateOffset() time.Duration {

	t := time.Now()
	i := t.Minute() % tickInterval

	return time.Duration(i) * time.Minute
}

func setToCurrent(s *Schedule, ucc chan connector.Connector) error {
	os := s.DefaultOutputSource

	now := time.Now()
	if len(s.Entries) > 0 {
		for _, e := range s.Entries {
			if now.After(e.Timestamp) {
				os = e.OutputSource
			} else if e.Timestamp.After(now) {
				break
			}
		}
	}
	uc := <-ucc
	defer func() { ucc <- uc }()

	err := axpert.SetOutputSourcePriority(uc, os)

	return err
}
