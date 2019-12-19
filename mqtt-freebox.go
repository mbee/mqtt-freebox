package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/juju2013/go-freebox"
	"github.com/yosssi/gmq/mqtt"
	"github.com/yosssi/gmq/mqtt/client"
)

var fbx *freebox.Client
var cli *client.Client
var sigc chan os.Signal

var (
	mqttURL      string
	mqttLogin    string
	mqttPassword string
)

type lanHost struct {
	L2Ident struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	} `json:"l2ident"`
	Active            bool          `json:"active"`
	ID                string        `json:"id"`
	LastTimeReachable freebox.Epoch `json:"last_time_reachable"`
	Persistent        bool          `json:"persistent"`
	VendorName        string        `json:"vendor_name"`
	HostType          string        `json:"host_type"`
	PrimaryName       string        `json:"primary_name"`
	L3Connectivities  []struct {
		Addr              string        `json:"addr"`
		Active            bool          `json:"active"`
		Reachable         bool          `json:"reachable"`
		LastActivity      freebox.Epoch `json:"last_activity"`
		AF                string        `json:"af"`
		LastTimeReachable freebox.Epoch `json:"last_time_reachable"`
	} `json:"l3connectivities"`
	Reachable         bool          `json:"reachable"`
	LastActivity      freebox.Epoch `json:"last_activity"`
	PrimaryNameManual bool          `json:"primary_name_manual"`
	Interface         string        `json:"interface"`
}

// Get all lan hosts
func getLanHosts(c *freebox.Client) ([]lanHost, error) {
	payload := []lanHost{}
	err := c.GetResult("lan/browser/pub/", &payload)
	return payload, err
}

// Get a contact
func getLanHost(c *freebox.Client, name string) (lanHost, error) {
	payload := lanHost{}
	err := c.GetResult(fmt.Sprintf("lan/browser/pub/%s", name), &payload)
	return payload, err
}

func initMqtt() {
	// Create an MQTT Client.
	cli = client.New(&client.Options{
		// Define the processing of the error handler.
		ErrorHandler: func(err error) {
			log.Fatal(err)
		},
	})
	// Connect to the MQTT Server.
	err := cli.Connect(&client.ConnectOptions{
		Network:  "tcp",
		Address:  mqttURL,
		UserName: []byte(mqttLogin),
		Password: []byte(mqttPassword),
		ClientID: []byte("mqtt-freebox2"),
	})
	if err != nil {
		log.Fatal(err)
	}

}

// send a message
func publish(topic, message string) error {
	// Publish a message.
	err := cli.Publish(&client.PublishOptions{
		QoS:       mqtt.QoS0,
		TopicName: []byte(topic),
		Message:   []byte(message),
	})
	if err != nil {
		log.Warn(err)
	}
	return err
}

func initFreebox() {
	fbx = freebox.New()

	err := fbx.Connect()
	if err != nil {
		log.Fatalf("fbx.Connect(): %v", err)
	}

	err = fbx.Authorize()
	if err != nil {
		log.Fatalf("fbx.Authorize(): %v", err)
	}

	err = fbx.Login()
	if err != nil {
		log.Fatalf("fbx.Login(): %v", err)
	}
}

func main() {
	// Set up channel on which to send signal notifications.
	sigc = make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, os.Kill)
	if os.Getenv("DEBUG") != "" {
		log.SetLevel(log.DebugLevel)
	}
	mqttURL = os.Getenv("GOFBX_MQTT_URL")
	mqttLogin = os.Getenv("GOFBX_MQTT_LOGIN")
	mqttPassword = os.Getenv("GOFBX_MQTT_PASSWORD")

	initMqtt()
	log.Info("Mqtt ... OK")
	defer cli.Terminate()

	initFreebox()

	// Subscribe to topics.
	err := cli.Subscribe(&client.SubscribeOptions{
		SubReqs: []*client.SubReq{
			&client.SubReq{
				TopicFilter: []byte("mqtt-freebox/get/host/#"),
				QoS:         mqtt.QoS0,
				// Define the processing of the message handler.
				Handler: func(topicName, message []byte) {
					deviceID := strings.Split(string(topicName), "/")[3]

					log.Info("get info for device: " + deviceID)
					lanhost, err := getLanHost(fbx, deviceID)
					if err != nil {
						log.Error(err)
						return
					}

					log.Info("publishing mqtt-freebox/status/host/" + lanhost.ID)
					payload := new(bytes.Buffer)
					encoder := json.NewEncoder(payload)
					if err := encoder.Encode(lanhost); err != nil {
						log.Error(err)
						return
					}
					payloadString := strings.TrimSpace(fmt.Sprintf("%s", payload))
					publish("mqtt-freebox/status/host/"+lanhost.ID, payloadString)

				},
			},
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	<-sigc

	// Disconnect the Network Connection.
	if err := cli.Disconnect(); err != nil {
		panic(err)
	}
}
