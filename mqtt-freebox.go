package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"

	"./freebox"

	log "github.com/Sirupsen/logrus"
	"github.com/yosssi/gmq/mqtt"
	"github.com/yosssi/gmq/mqtt/client"
)

var fbx *freebox.Client
var cli *client.Client
var sigc chan os.Signal

var (
	smsLogin     string
	smsPass      string
	mqttURL      string
	mqttLogin    string
	mqttPassword string
)

func main() {
	// Set up channel on which to send signal notifications.
	sigc = make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, os.Kill)
	if os.Getenv("DEBUG") != "" {
		log.SetLevel(log.DebugLevel)
	}
	smsLogin = os.Getenv("GOFBX_SMS_LOGIN")
	smsPass = os.Getenv("GOFBX_SMS_PASS")
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
					lanhost, err := fbx.GetLanHost(deviceID)
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
		ClientID: []byte("mqtt-freebox"),
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

// check if there's incomming call
func checkCall() (*freebox.CallEntry, bool) {
	calls, err := fbx.GetCallEntries()
	if err != nil {
		return nil, false
	}
	for _, c := range calls {
		if (c.Type == "missed") && (c.New) {
			return &c, true
		}
	}
	return nil, false
}

// notify by sms
func notifySMS(msg string) {
	if (smsLogin == "") || (smsPass) == "" {
		return
	}

	data := url.Values{
		"user": {smsLogin},
		"pass": {smsPass},
		"msg":  {msg}}
	response, err := http.Get("https://smsapi.free-mobile.fr/sendmsg?" + data.Encode())
	fmt.Printf("DEBUG:data=%v", data.Encode())

	if err != nil {
		log.Warn(response)
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	fmt.Printf("DEBUG:body=%v", body)
	if err != nil {
		log.Warn(response)
	} else {
		log.WithFields(log.Fields{"http status": response.Status}).Info("Seding SMS ...")
	}
}
