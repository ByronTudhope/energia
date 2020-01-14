package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/goburrow/serial"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/mindworks-software/energia/pkg/axpert"
	"github.com/mindworks-software/energia/pkg/connector"
	"github.com/mindworks-software/energia/pkg/pylontech"
	"github.com/mindworks-software/energia/pkg/schedule"
)

var timerRapid  time.Duration
var timerNormal  time.Duration
var timerPeriodic  time.Duration

var mqttServer string
var mqttPort int
var mqttClientId string
var mqttUsername string
var mqttPassword string

var inverterEnabled bool
var inverterPath string
var inverterCount int
var inverterTopic string

var batteryEnabled bool
var batteryPath string
var batteryBaud int
var batteryTopic string

var scheduleTickInterval int
var scheduleEnabled bool

type messageData struct {
	Timestamp   time.Time
	MessageType string
	Data        interface{}
}

type queryFunc func(chan connector.Connector, mqtt.Client, time.Time) error

type query struct {
	f        queryFunc
	cc       chan connector.Connector
	interval time.Duration
}

var ucc chan connector.Connector

func main() {
	fmt.Println("initializing config ")

	err := initConfig()
	fmt.Println("initialized config ", viper.AllSettings())
	if err != nil {
		panic(err)
	}

	fmt.Println("connecting to ", inverterPath)
	uc, err := connector.NewUSBConnector(inverterPath)
	if err != nil {
		panic(err)
	}
	err = uc.Open()
	if err != nil {
		panic(err)
	}
	defer uc.Close()

	ucc = make(chan connector.Connector, 1)
	ucc <- uc
	fmt.Println("connected to ", inverterPath)

	var sc connector.Connector
	var scc chan connector.Connector

	if batteryEnabled {

		serialConfig := serial.Config{
			Address:  batteryPath,
			BaudRate: batteryBaud,
			DataBits: 8,
			StopBits: 1,
			Parity:   "N",
			Timeout:  30 * time.Second,
		}

		sc = connector.NewSerialConnector(serialConfig)
		err = sc.Open()
		if err != nil {
			log.Panic(err)
		}
		defer sc.Close()

		scc = make(chan connector.Connector, 1)
		scc <- sc
	}

	clientOpts := mqtt.NewClientOptions()
	clientOpts.AddBroker("tcp://" + mqttServer + ":" + strconv.Itoa(mqttPort))
	clientOpts.SetAutoReconnect(true)
	clientOpts.SetStore(mqtt.NewFileStore("/tmp/mqtt"))
	clientOpts.SetCleanSession(false)
	clientOpts.SetClientID(mqttClientId)
	clientOpts.SetOnConnectHandler(logConnect)
	clientOpts.SetConnectionLostHandler(logConnectionLost)
	clientOpts.SetUsername(mqttUsername)
	clientOpts.SetPassword(mqttPassword)

	client := mqtt.NewClient(clientOpts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	defer client.Disconnect(250)
	fmt.Println("Connected to mqtt")

	queries := []query{
		{deviceMode, ucc, timerPeriodic * time.Second},
		{parallelDeviceInfo, ucc, timerRapid * time.Second},
		{deviceGeneralStatus, ucc, timerRapid * time.Second},
		{deviceFlagStatus, ucc, timerPeriodic * time.Second},
		{warningStatus, ucc, timerNormal * time.Second},
		{deviceRating, ucc, timerRapid * time.Second},
	}

	if batteryEnabled {
		queries = append(queries, query{batteryStatus, scc, timerNormal * time.Second})
	}

	ts := make([]*time.Ticker, len(queries))

	for i, q := range queries {
		ts[i] = scheduleQuery(q.f, q.interval, q.cc, client)
	}

	topics := make(map[string]byte, 3)
	topics[schedule.ScheduleTopic] = 1
	topics[schedule.OverrideTopic] = 1
	topics[schedule.EnableTopic] = 1
	client.SubscribeMultiple(topics, messageReceiver)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan
	close(sigChan)
	fmt.Println(sig, " stopping tickers")

	for _, t := range ts {
		t.Stop()
	}

	fmt.Println("exiting")
}

func scheduleQuery(f queryFunc, interval time.Duration, ucc chan connector.Connector, client mqtt.Client) *time.Ticker {
	ticker := time.NewTicker(interval)
	go func() {
		for t := range ticker.C {
			f(ucc, client, t)
		}
	}()
	return ticker
}

func deviceGeneralStatus(ucc chan connector.Connector, client mqtt.Client, t time.Time) error {

	uc := <-ucc

	defer func() { ucc <- uc }()

	status, err := axpert.DeviceGeneralStatus(uc)
	if err != nil {
		return err
	}
	msgData := messageData{Timestamp: t, MessageType: "Status", Data: status}
	err = sendInverterMessage(msgData, client)
	if err != nil {
		return err
	}

	return nil

}

func warningStatus(ucc chan connector.Connector, client mqtt.Client, t time.Time) error {

	uc := <-ucc

	defer func() { ucc <- uc }()

	warnings, err := axpert.WarningStatus(uc)
	if err != nil {
		return err
	}
	msgData := messageData{Timestamp: t, MessageType: "Warnings", Data: warnings}
	err = sendInverterMessage(msgData, client)
	if err != nil {
		return err
	}
	return nil

}

func deviceFlagStatus(ucc chan connector.Connector, client mqtt.Client, t time.Time) error {

	uc := <-ucc

	defer func() { ucc <- uc }()

	flags, err := axpert.DeviceFlagStatus(uc)
	msgData := messageData{Timestamp: t, MessageType: "Flags", Data: flags}
	err = sendInverterMessage(msgData, client)
	if err != nil {
		return err
	}
	return nil

}

func deviceRating(ucc chan connector.Connector, client mqtt.Client, t time.Time) error {

	uc := <-ucc

	defer func() { ucc <- uc }()

	ratingInfo, err := axpert.DeviceRatingInfo(uc)
	msgData := messageData{Timestamp: t, MessageType: "RatingInfo", Data: ratingInfo}
	err = sendInverterMessage(msgData, client)
	if err != nil {
		return err
	}

	return nil

}

func batteryStatus(ucc chan connector.Connector, client mqtt.Client, t time.Time) error {

	uc := <-ucc

	defer func() { ucc <- uc }()

	batteryStatus, err := pylontech.GetBatteryStatus(uc)
	msgData := messageData{Timestamp: t, MessageType: "BatteryStatus", Data: batteryStatus}
	err = sendBatteryMessage(msgData, client)
	if err != nil {
		return err
	}

	return nil

}

func parallelDeviceInfo(ucc chan connector.Connector, client mqtt.Client, t time.Time) error {

	uc := <-ucc

	defer func() { ucc <- uc }()

	for inv := 0; inv < inverterCount; inv++ {
		deviceInfo, err := axpert.ParallelDeviceInfo(uc, inv)
		if err != nil {
			return err
		}
		msgData := messageData{Timestamp: t, MessageType: "DeviceInfo", Data: deviceInfo}
		err = sendInverterMessage(msgData, client)
		if err != nil {
			return err
		}
	}
	return nil
}

func deviceMode(ucc chan connector.Connector, client mqtt.Client, t time.Time) error {

	uc := <-ucc

	defer func() { ucc <- uc }()

	mode, err := axpert.DeviceMode(uc)
	if err != nil {
		return err
	}
	m := map[string]string{"Mode": mode}
	msgData := messageData{Timestamp: t, MessageType: "Mode", Data: m}
	err = sendInverterMessage(msgData, client)
	if err != nil {
		return err
	}

	return nil
}

func sendInverterMessage(data messageData, client mqtt.Client) error {
	return sendMessage(data, inverterTopic+"/"+data.MessageType, client)
}

func sendBatteryMessage(data messageData, client mqtt.Client) error {
	return sendMessage(data, batteryTopic, client)
}

func sendMessage(data messageData, topic string, client mqtt.Client) error {
	msg, err := json.Marshal(data)
	if err != nil {
		return err
	}
	token := client.Publish(topic, 1, true, msg)
	token.Wait()
	return nil
}

func messageReceiver(client mqtt.Client, msg mqtt.Message) {

	go func() {
		payload := string(msg.Payload())
		switch msg.Topic() {
		case schedule.OverrideTopic:
			fmt.Printf("%s\n%s\n", msg.Topic(), msg.Payload())
			uc := <-ucc

			defer func() { ucc <- uc }()

			if payload[0] == '{' {
				// Handle json payload
				fmt.Println("JSON coming soon to a queue near you")
			} else {
				outputSource, suspendDuration, err := parseOverride(string(msg.Payload()))
				if err != nil {
					fmt.Println("Failed to override schedule", err)
				}
				overrideOutputSource(outputSource, uc, suspendDuration)

			}

		case schedule.ScheduleTopic:
			fmt.Printf("%s\n", msg.Topic())
			_, err := schedule.CreateSchedule(msg, ucc, scheduleTickInterval, scheduleEnabled)
			if err != nil {
				fmt.Println("Failed creating schedule", err)
				return
			}

		case schedule.EnableTopic:
			fmt.Printf("%s\n", msg.Topic())

			isEnabled, err := strconv.ParseBool(payload)
			if err != nil {
				fmt.Println("Failed to parse payload ", err)
			}

			if isEnabled {
				schedule.Enable()
			} else {
				schedule.Disable()
			}

		default:
			fmt.Printf("%s\n", msg.Topic())
		}
	}()
}

func overrideOutputSource(outputSource axpert.OutputSourcePriority, uc connector.Connector, suspendDuration time.Duration) {
	schedule.Disable()
	err := axpert.SetOutputSourcePriority(uc, outputSource)
	if err != nil {
		fmt.Println("Failed sending command ", err)
	}

	if suspendDuration > 0 {
		fmt.Printf("Schedule will be restored in %d minutes", suspendDuration)
		time.AfterFunc(suspendDuration, func() {
			fmt.Println("Resuming regularly scheduled programming")
			schedule.Enable()
		})
	}
}

func parseOverride(msg string) (axpert.OutputSourcePriority, time.Duration, error) {
	parts := strings.Split(msg, ",")
	params := make([]int, len(parts))
	var err error
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
		params[i], err = strconv.Atoi(parts[i])
		if err != nil {
			return 0, 0, err
		}
	}

	outputSource := axpert.OutputSourcePriority(params[0])

	switch len(parts) {
	case 0:
		return 0, 0, fmt.Errorf("invalid payload: blank")
	case 1:
		return outputSource, -1, nil
	case 2:
		fallthrough
	default:
		suspendDuration := time.Duration(params[1]) * time.Minute
		return outputSource, suspendDuration, nil

	}
}

func logConnect(_ mqtt.Client) {
	fmt.Println("Connected to broker")
}

func logConnectionLost(_ mqtt.Client, err error) {
	fmt.Println("Connection lost:", err)
}

func initConfig() error {
	var configPath string
	pflag.StringVarP(&configPath, "config-path", "c", ".", "Path to config file (datalogd-conf.yaml)")
	pflag.Parse()

	viper.SetDefault("mqtt.server", "localhost")
	viper.SetDefault("mqtt.port", 1883)
	viper.SetDefault("mqtt.clientid", "datalogd")
	viper.SetDefault("timer.interval.rapid", 5)
	viper.SetDefault("timer.interval.normal", 10)
	viper.SetDefault("timer.interval.periodic", 30)
	viper.SetDefault("inverter.count", 1)
	viper.SetDefault("inverter.topic", "datalogd/inverter")
	viper.SetDefault("battery.baud", 1200)
	viper.SetDefault("battery.topic", "datalogd/battery")
	viper.SetDefault("schedule.tickInterval", 10)

	viper.SetEnvPrefix("dlog")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetConfigName("datalogd-conf")
	if configPath != "" {
		viper.AddConfigPath(configPath)
	}
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("Config file not found, relying on defaults/ENV")
		} else {
			return err
		}
	}

	fmt.Println("config: ", viper.AllSettings())
	timerRapid = viper.GetDuration("timer.interval.rapid")
	timerNormal = viper.GetDuration("timer.interval.normal")
	timerPeriodic = viper.GetDuration("timer.interval.periodic")
	mqttServer = viper.GetString("mqtt.server")
	mqttPort = viper.GetInt("mqtt.port")
	mqttUsername = viper.GetString("mqtt.username")
	mqttPassword = viper.GetString("mqtt.password")
	mqttClientId = viper.GetString("mqtt.clientId")
	inverterEnabled = viper.GetBool("inverter.enabled")
	inverterPath = viper.GetString("inverter.path")
	inverterCount = viper.GetInt("inverter.count")
	inverterTopic = viper.GetString("inverter.topic")
	batteryEnabled = viper.GetBool("battery.enabled")
	batteryPath = viper.GetString("battery.path")
	batteryBaud = viper.GetInt("battery.baud")
	batteryTopic = viper.GetString("battery.topic")
	scheduleEnabled = viper.GetBool("schedule.enabled")
	scheduleTickInterval = viper.GetInt("schedule.tickInterval")

	return nil
}
