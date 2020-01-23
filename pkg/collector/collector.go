package collector

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Collector interface {
	GetData(startTime string, endTime string) ([]TimeValue, error)
	GetAllData() ([]TimeValue, error)
	GetCurrent() float64
}

type DataCollector struct {
	name string

	buffer    []float64
	minuteAvg []float64
	tck       *time.Ticker
}

type TimeValue struct {
	Time  string
	Value float64
}

type DailyArchive struct {
	Date   string
	Values []float64
}

/*
  Collector should use given mqtt client and subscripe to given topic.
  On start up it should create an internal ticker that will ticke every full minute.
  Should it be started in between a full minute, create timer that will start ticker when full minuite strikes
  On each tick buffered values are read, mean calculated, and stored in DataCollector.minuteAvg
  minuteAvg is initiated with 24x60 zero values.
    Each buffer avg is inserted in specific minutes index.
    Index is calculated using current Hour and Minute, H*60+M
  Output will use minuteAvg index to workout the minute for each store value
    H = I/60
    M = I%60



*/

var dcMap = make(map[string]*DataCollector)
var isEnabled = true
var tickInterval = 1
var tempChan chan []float64

const dailyMinutes = 24 * 60
const ISODate = "2006-01-02"

func StartCollector(mqc mqtt.Client, topic string) *DataCollector {
	tempChan = make(chan []float64, 1)
	mqc.Subscribe(topic+"/temp", 1, handleTempMessage)

	var dc *DataCollector
	select {
	case minAvg := <-tempChan:
		fmt.Println("Initialising from stored temp msg. Topic is", topic)
		dc = &DataCollector{name: topic, minuteAvg: minAvg}

	case <-time.After(2000 * time.Millisecond):
		fmt.Println("Starting with empty dataset. Topic is", topic)
		dc = &DataCollector{name: topic, minuteAvg: make([]float64, dailyMinutes)}
	}

	mqc.Unsubscribe(topic + "/temp")

	dcMap[topic] = dc
	// if we don't have a ticker yet
	// create  timer for now + (time to next clock 60 sec interval)
	// when timer ticks, start ticker for 60 sec intervals
	if dc.tck == nil {
		ch := time.After(calculateOffset())
		go func() {
			<-ch
			dc.tck = startTicker(mqc, topic)
		}()

	}

	return dc

}

func handleTempMessage(client mqtt.Client, message mqtt.Message) {
	fmt.Println()
	temp := &DailyArchive{}
	err := json.Unmarshal(message.Payload(), &temp)
	if err != nil {
		fmt.Println("error unmarshalling temp values", message.Payload())
	}
	if temp.Date == time.Now().Format(ISODate) {
		tempChan <- temp.Values
		fmt.Println("Posted temp to channel. Topic is", message.Topic())
	}
}

func (dc *DataCollector) GetData(startTime string, endTime string) ([]TimeValue, error) {
	var err error
	startIndex := 0
	if startTime != "" {
		startIndex, err = parseIndex(startTime)
		if err != nil {
			return nil, err
		}
	}

	endIndex := len(dc.minuteAvg) - 1
	if endTime != "" {
		endIndex, err = parseIndex(endTime)
		if err != nil {
			return nil, err
		}
	}

	if startIndex < 0 || endIndex < 0 || startIndex >= dailyMinutes || endIndex >= dailyMinutes {
		err = fmt.Errorf("invalid time, must be 00:00-23:59")
		return nil, err
	}

	if startIndex >= len(dc.minuteAvg) {
		startIndex = len(dc.minuteAvg) - 1
	}

	if endIndex >= len(dc.minuteAvg) {
		endIndex = len(dc.minuteAvg) - 1
	}

	vals := dc.getTimeValues(startIndex, endIndex)

	return vals, nil
}

func (dc *DataCollector) getTimeValues(startIndex int, endIndex int) []TimeValue {
	vals := make([]TimeValue, 0, 2000)
	for i, v := range dc.minuteAvg[startIndex : endIndex+1] {
		ts := fmt.Sprintf("%02d:%02d", (i+startIndex)/60, (i+startIndex)%60)
		vals = append(vals, TimeValue{
			Time:  ts,
			Value: v,
		})
	}
	return vals
}

func (dc *DataCollector) GetAllData() ([]TimeValue, error) {
	return dc.GetData("", "")
}

func (dc *DataCollector) GetCurrent() float64 {
	return dc.buffer[len(dc.buffer)-1]
}

func parseIndex(time string) (int, error) {
	split := strings.Split(time, ":")
	if len(split) != 2 {
		err := fmt.Errorf("invalid time string: %s, should be hh:mm", time)
		return -1, err
	}

	hour, err := strconv.Atoi(split[0])
	if err != nil {
		return -1, err
	}

	min, err := strconv.Atoi(split[1])
	if err != nil {
		return -1, err
	}

	return hour*60 + min, nil
}

func startTicker(mqc mqtt.Client, topic string) *time.Ticker {
	fmt.Println("Starting ticker")
	tck := time.NewTicker(time.Duration(tickInterval) * time.Minute)
	mqc.Subscribe(topic, 1, handleDataMessage)
	go func() {
		for t := range tck.C {

			if isEnabled {
				fmt.Println("Tick for : ", topic)
				dc := dcMap[topic]
				if t.Hour() == 0 && t.Minute() == 0 {
					archive := DailyArchive{
						Date:   t.Add(-1 * time.Hour).Format(ISODate),
						Values: dc.minuteAvg,
					}
					archiveJson, err := json.Marshal(archive)
					if err != nil {
						fmt.Println("Error formatting temp", err)
					}
					mqc.Publish(topic+"/archive", 1, true, archiveJson)
					dc.minuteAvg = make([]float64, dailyMinutes)
				}

				avgBuff(dc, t)

				temp := DailyArchive{
					Date:   t.Format(ISODate),
					Values: dc.minuteAvg,
				}
				avgJson, err := json.Marshal(temp)
				if err != nil {
					fmt.Println("Error formatting temp", err)
				}
				mqc.Publish(topic+"/temp", 1, true, avgJson)

				fmt.Println("Tock for : ", topic)
			} else {
				fmt.Println("Collector is disabled: ", topic)
			}
		}

	}()
	return tck
}

func handleDataMessage(_ mqtt.Client, msg mqtt.Message) {

	dc := dcMap[msg.Topic()]

	if dc != nil {
		v, err := strconv.ParseFloat(string(msg.Payload()), 64)
		if err == nil {
			dc.buffer = append(dc.buffer, v)

		}
	}

}

func avgBuff(dc *DataCollector, t time.Time) {
	l := float64(len(dc.buffer))

	var avg float64
	var sum float64

	if l > 0 {
		for _, v := range dc.buffer {
			sum += v
		}

		avg = sum / l
	}

	h := t.Hour()
	m := t.Minute()

	i := h*60 + m
	dc.minuteAvg[i] = avg

	dc.buffer = nil
}

func calculateOffset() time.Duration {

	t := time.Now()
	i := tickInterval - (t.Second() % tickInterval)
	return time.Duration(i) * time.Second
}
