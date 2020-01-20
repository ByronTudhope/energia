package collector

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type dataCollector struct {
	Name string

	Buffer    []float64
	MinuteAvg []float64
	tck       *time.Ticker
}

type outputVal struct {
	Time  string
	Value float64
}

/*
  Collector should use given mqtt client and subscripe to given topic.
  On start up it should create an internal ticker that will ticke every full minute.
  Should it be started in between a full minute, create timer that will start ticker when full minuite strikes
  On each tick buffered values are read, mean calculated, and stored in dataCollector.MinuteAvg
  MinuteAvg is initiated with 24x60 zero values.
    Each buffer avg is inserted in specific minutes index.
    Index is calculated using current Hour and Minute, H*60+M
  Output will use MinuteAvg index to workout the minute for each store value
    H = I/60
    M = I%60



*/

var dcMap = make(map[string]*dataCollector)
var isEnabled = true
var tickInterval = 1

func StartCollector(mqc mqtt.Client, topic string) func(http.ResponseWriter, *http.Request) {

	dc := &dataCollector{Name: topic, MinuteAvg: make([]float64, 24*60)}

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

	return httpHandler

}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	topic := r.URL.Path[1:]
	dc := dcMap[topic]
	vals := make([]outputVal, 0, 2000)
	for i, v := range dc.MinuteAvg {
		ts := fmt.Sprintf("%02d:%02d", i/60, i%60)
		vals = append(vals, outputVal{
			Time:  ts,
			Value: v,
		})
	}

	jsonBytes, err := json.MarshalIndent(vals, "", "  ")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
	}
	header := w.Header()
	header["Content-Type"] = []string{"application/json"}
	w.WriteHeader(200)
	w.Write(jsonBytes)
}

func startTicker(mqc mqtt.Client, topic string) *time.Ticker {
	fmt.Println("Starting ticker")
	tck := time.NewTicker(time.Duration(tickInterval) * time.Minute)
	mqc.Subscribe(topic, 1, handleMessage)
	go func() {
		for t := range tck.C {

			if isEnabled {
				fmt.Println("Tick for : ", topic)
				avgBuff(dcMap[topic], t)
				//TODO something clever to reset the thing at midnight

				fmt.Println("Tock for : ", topic)
			} else {
				fmt.Println("Collector is disabled: ", topic)
			}
		}

	}()
	return tck
}

func handleMessage(client mqtt.Client, msg mqtt.Message) {

	dc := dcMap[msg.Topic()]

	if dc != nil {
		v, err := strconv.ParseFloat(string(msg.Payload()), 64)
		if err == nil {
			dc.Buffer = append(dc.Buffer, v)

		}
	}

}

func avgBuff(dc *dataCollector, t time.Time) {
	l := float64(len(dc.Buffer))
	var sum float64
	for _, v := range dc.Buffer {
		sum += v
	}

	avg := sum / l

	h := t.Hour()
	m := t.Minute()

	i := h*60 + m
	dc.MinuteAvg[i] = avg

	dc.Buffer = nil
}

func calculateOffset() time.Duration {

	t := time.Now()
	i := tickInterval - (t.Second() % tickInterval)
	return time.Duration(i) * time.Second
}
