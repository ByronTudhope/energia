package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/mindworks-software/energia/pkg/collector"
)

var timerRapid time.Duration
var timerNormal time.Duration
var timerPeriodic time.Duration

var mqttServer string
var mqttPort int
var mqttClientId string
var mqttUsername string
var mqttPassword string

func main() {
	fmt.Println("initializing config ")

	err := initConfig()
	fmt.Println("initialized config ", viper.AllSettings())
	if err != nil {
		panic(err)
	}

	client, err := connectMQTT()
	if err != nil {
		panic(err)
	}

	defer client.Disconnect(250)
	fmt.Println("Connected to mqtt")

	handler := collector.StartCollector(client, "emon/datalogd-ng/Status_ACOutputVoltage")

	http.HandleFunc("/emon/datalogd-ng/Status_ACOutputVoltage", handler)

	http.ListenAndServe(":9090", nil)

}

func connectMQTT() (mqtt.Client, error) {

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
		// panic(token.Error())
		return nil, token.Error()
	}

	return client, nil

}

func logConnect(_ mqtt.Client) {
	fmt.Println("Connected to broker")
}

func logConnectionLost(_ mqtt.Client, err error) {
	fmt.Println("Connection lost:", err)
}

func initConfig() error {
	var configPath string
	pflag.StringVarP(&configPath, "config-path", "c", ".", "Path to config file (server-conf.yaml)")
	pflag.Parse()

	viper.SetDefault("mqtt.server", "localhost")
	viper.SetDefault("mqtt.port", 1883)
	viper.SetDefault("mqtt.clientid", "server")
	viper.SetDefault("timer.interval.rapid", 5)
	viper.SetDefault("timer.interval.normal", 10)
	viper.SetDefault("timer.interval.periodic", 30)

	viper.SetEnvPrefix("dlog")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetConfigName("server-conf")
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

	return nil
}
