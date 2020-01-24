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

	"github.com/gin-gonic/gin"

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

var collectors map[string]*collector.DataCollector

type CollectorConfig struct {
	Topic   string `mapstructure:"topic"`
	Enabled bool   `mapstructure:"enabled"`
}

func main() {
	fmt.Println("initializing config ")

	err := initConfig()
	fmt.Println("initialized config ", viper.AllSettings())
	if err != nil {
		panic(err)
	}

	var collectorConfig []CollectorConfig

	viper.UnmarshalKey("collectors", &collectorConfig)

	client, err := connectMQTT()
	if err != nil {
		panic(err)
	}

	defer client.Disconnect(250)
	fmt.Println("Connected to mqtt")

	collectors = make(map[string]*collector.DataCollector)
	for _, config := range collectorConfig {
		if config.Enabled {
			collector := collector.StartCollector(client, config.Topic)
			collectors[config.Topic] = collector
		}
	}

	router := gin.Default()

	router.GET("/", home)
	router.Static("/dash", "./dash")
	router.GET("/collectors", collectorList)

	//	router.GET("/emon/:systemname/:topic", minuteAvgHandler)
	//	router.GET("/emon/:systemname/:topic/last", current)

	router.GET("/emon/:systemname/:topic/:subtopic", minuteAvgHandler)
	router.GET("/emon/:systemname/:topic/:subtopic/*action", current)

	router.Run(":9090")
}

func collectorList(c *gin.Context) {
	topics := make([]string, 0)

	for k, _ := range collectors {
		topics = append(topics, k)
	}

	c.JSON(http.StatusOK, topics)
}

func home(c *gin.Context) {
	var links string

	for k, _ := range collectors {
		link := fmt.Sprintf("<a href=%s>%s</a><br>", k, k)
		links += link
		link = fmt.Sprintf("<a href=%s/last>%s/last</a><br>", k, k)
		links += link
		st := time.Now().Add(-1 * time.Hour).Format("15:04")
		et := time.Now().Add(1 * time.Minute).Format("15:04")
		link = fmt.Sprintf("<a href=%s?startTime=%s&endTime=%s>%s between %s & %s</a><br>", k, st, et, k, st, et)
		links += link
	}

	c.Data(http.StatusOK, "text/html", []byte(links))

}

func minuteAvgHandler(c *gin.Context) {
	var vals []collector.TimeValue
	var err error

	topic := c.Request.URL.Path[1:]
	dc := collectors[topic]
	st := c.Query("startTime")
	et := c.Query("endTime")

	if st != "" && et != "" {
		vals, err = dc.GetData(st, et)
	} else {

		vals, err = dc.GetAllData()
	}

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	c.JSON(http.StatusOK, vals)
}

func current(c *gin.Context) {
	topic := fmt.Sprintf("emon/%s/%s", c.Param("systemname"), c.Param("topic"))
	if c.Param("subtopic") != "" {
		topic += fmt.Sprintf("/%s", c.Param("subtopic"))
	}
	fmt.Println("Topic : ", topic)

	dc := collectors[topic]
	val := dc.GetCurrent()

	c.JSON(http.StatusOK, val)
}
func connectMQTT() (mqtt.Client, error) {

	clientOpts := mqtt.NewClientOptions()
	clientOpts.AddBroker("tcp://" + mqttServer + ":" + strconv.Itoa(mqttPort))
	clientOpts.SetAutoReconnect(true)
	clientOpts.SetStore(mqtt.NewFileStore("/tmp/mqtt/" + mqttClientId))
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
