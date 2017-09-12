package main

import (
    "fmt"
    "os"
    "log"
    "strconv"
    "bytes"
    "encoding/json"
    "flag"
    "time"
    "os/signal"

    MQTT "github.com/eclipse/paho.mqtt.golang"
)

// Structures to decode the JSON schema

type Gateway struct {
    Gtw_id         string
    Timestanp      int64
    Time           string
    Channel        int
    Rssi           int
    Snr            int
    Latitude       float32
    Longitude      float32
    Altitude       float32
}

type Gateway_list []Gateway

type Meta  struct {
    Time           string
    Frequency      float32
    Modulation     string
    Data_rate      string
    Coding_rate    string
    Gateways       Gateway_list
}

type Payload_decoded struct {
    Temperature    string
}

// Main Http post structure
type Payload struct{
    App_id          string
    Dev_id          string
    Hardware_serial string
    Port            int
    Counter         int
    Is_retry        bool
    Payload_raw     string
    Payload_fields  Payload_decoded
    Metadata        Meta
    Download_url    string
}

// Broker configuration
type MqttConfig struct {
    Host           string
    Port           int
    ClientID       string
    TopicPrefix    string
    TopicSuffix    string
    Qos            byte
    Cleansess      bool
    ReconnectDelay time.Duration
}

type TtnConfig struct {
    Host            string
    Port            int
    AppID           string
    AppKey          string
    RegionalHandler string
    Topic           string
    ClientID        string
}

// Http server configuration
type HttpConfig struct {
    Host        string
    Port        int
}

type Configuration struct {
    Broker MqttConfig
    Ttn    TtnConfig
}

// Configuration file
var (
    configFilename string
    config         Configuration
)

func onConnectionLost(client MQTT.Client, err error) {
    log.Println("Connection lost : ", err)
}

func onConnect(client MQTT.Client) {
    log.Println("Connected to Broker : " + "tcp://" + config.Broker.Host + ":" + strconv.Itoa(config.Broker.Port))
}

func checkErr(err error) {
    if err != nil {
        panic(err)
    }
}
var f MQTT.MessageHandler = func(mqttClient MQTT.Client, msg MQTT.Message) {
    var payload Payload
    reader := bytes.NewReader(msg.Payload())
    decoder := json.NewDecoder(reader)
    err := decoder.Decode(&payload)
    checkErr(err)
    log.Println(payload)

    broker := "tcp://" + config.Broker.Host + ":" + strconv.Itoa(config.Broker.Port)
    topic := config.Broker.TopicPrefix + "/" + payload.App_id + "_" + payload.Dev_id + "/" + config.Broker.TopicSuffix

    fmt.Printf("\nSample Info:\n")
    fmt.Printf("\tbroker:    %s\n", broker)
    fmt.Printf("\tclientId:  %s\n", config.Broker.ClientID)
    fmt.Printf("\ttopic:     %s\n", topic)
    fmt.Printf("\ttemp:     %s\n", payload.Payload_fields.Temperature)

    // MQTT client options
    opts := MQTT.NewClientOptions()
    opts.AddBroker(broker)
    opts.SetClientID(config.Broker.ClientID)
    opts.SetCleanSession(config.Broker.Cleansess)
    opts.SetAutoReconnect(true)
    opts.SetMaxReconnectInterval(config.Broker.ReconnectDelay)
    opts.SetConnectionLostHandler(onConnectionLost)
    opts.SetOnConnectHandler(onConnect)

    client := MQTT.NewClient(opts)
    if token := client.Connect(); token.Wait() && token.Error() != nil {
        panic(token.Error())
    }

    // Publish temperature
    publish(client, "Temperature", payload.Payload_fields.Temperature, topic)

    client.Disconnect(250)
    fmt.Println("\nPublished and disconnected")
}

func publish(client MQTT.Client, dataName string, data string, topic string) {
    mqtt_payload := `{ "` + dataName + `" : ` + data + `}`
    token := client.Publish(topic, config.Broker.Qos, config.Broker.Cleansess, mqtt_payload)
    token.Wait()
}

func main() {

    // Processing optional config file
    flag.StringVar(&configFilename, "config", "./config.json", "default value ./config.json")
    flag.Parse()

    // Set up channel on which to send signal notifications.
    sigc := make(chan os.Signal, 1)
    signal.Notify(sigc, os.Interrupt, os.Kill)

    log.Print("Using Configuration filename : " + configFilename)

    // Parsing configuration
    file, err := os.Open(configFilename)
    if err != nil {
        log.Println("error:", err)
        os.Exit(1)
    }
    decoder := json.NewDecoder(file)

    err = decoder.Decode(&config)
    checkErr(err)

    log.Println("Configuration : ")
    log.Println(config.Broker)
    log.Println(config.Ttn)

    br := "tcp://" + config.Ttn.RegionalHandler + "." + config.Ttn.Host + ":" + strconv.Itoa(config.Ttn.Port)
    log.Println("Opening :", br)
    opts := MQTT.NewClientOptions().AddBroker(br)
    opts.SetClientID(config.Broker.ClientID)
    opts.SetDefaultPublishHandler(f)
    opts.SetUsername(config.Ttn.AppID)
    opts.SetPassword(config.Ttn.AppKey)

    //create and start a client using the above ClientOptions
    c := MQTT.NewClient(opts)
    if token := c.Connect(); token.Wait() && token.Error() != nil {
        panic(token.Error())
    }
    log.Println("Connected to Broker")

    if token := c.Subscribe(config.Ttn.Topic, 0, nil); token.Wait() && token.Error() != nil {
        log.Println(token.Error())
        os.Exit(1)
    }

    for {
        select {
        case <-sigc:
            log.Println("interrupt")
            return
        }
    }
}
