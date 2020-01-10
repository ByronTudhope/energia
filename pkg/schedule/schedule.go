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
	Timestamp    string
	OutputSource axpert.OutputSourcePriority
	timeOnly     time.Time
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

var sch *Schedule
var ucc chan connector.Connector
var tck *time.Ticker

var tickInterval int
var isEnabled bool

func CreateSchedule(msg mqtt.Message, ucchan chan connector.Connector, interval int, enabled bool) (*Schedule, error) {
	s, err := unmarshalSchedule(msg.Payload())
	if err != nil {
		return nil, err
	}

	sch = s
	ucc = ucchan
	tickInterval = interval
	isEnabled = enabled
	// Set output to current from schedule
	err = setToCurrent(sch, ucc)
	if err != nil {
		return nil, err
	}

	// if we don't have a ticker yet
	// create  timer for now + (time to next clock 10 min interval)
	// when timer ticks, start ticker for 10 min intervals
	if tck == nil {
		time.AfterFunc(calculateOffset(), startTicker)
	}

	return sch, nil
}

func unmarshalSchedule(msg []byte) (*Schedule, error) {
	var s Schedule
	err := json.Unmarshal(msg, &s)
	if err != nil {
		return nil, err
	}
	for i := range s.Entries {
		s.Entries[i].timeOnly, err = time.Parse("15:04:05", s.Entries[i].Timestamp)
		if err != nil {
			return nil, err
		}
	}

	return &s, nil
}

func startTicker() {
	fmt.Println("Starting ticker")
	tck = time.NewTicker(time.Duration(tickInterval) * time.Minute)
	go func() {
		for range tck.C {
			if isEnabled {
				fmt.Println("Tick")
				err := setToCurrent(sch, ucc)
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println("Tock")
			} else {
				fmt.Println("Schedule is disabled")
			}
		}

	}()
	err := setToCurrent(sch, ucc)
	if err != nil {
		fmt.Println("Failed setrting current", err)
	}
}

func calculateOffset() time.Duration {

	t := time.Now()
	i := tickInterval - (t.Minute() % tickInterval)
	return time.Duration(i) * time.Minute
}

func setToCurrent(s *Schedule, ucc chan connector.Connector) error {
	os := s.DefaultOutputSource

	now, err := timeOnly(time.Now())

	if err != nil {
		fmt.Println("Failed to parse time only ", err)
		return err
	}

	if len(s.Entries) > 0 {
		for _, e := range s.Entries {
			if now.After(e.timeOnly) {
				os = e.OutputSource
			} else if e.timeOnly.After(now) {
				break
			}
		}
	}
	uc := <-ucc
	defer func() { ucc <- uc }()

	err = axpert.SetOutputSourcePriority(uc, os)

	return err
}

func timeOnly(t time.Time) (time.Time, error) {

	return time.Parse("15:04:05", t.Format("15:04:05"))

}
