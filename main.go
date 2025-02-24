package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

var (
	debug      = false
	configPath = "config.yaml"
)

func loadCfg() Config {
	f := must(os.Open(configPath))
	defer f.Close()
	return must(loadYAML(f))
}

func main() {
	flag.StringVar(&configPath, "config", "config.yaml", "Path to the configuration file")
	flag.Parse()

	cfg := loadCfg()
	debug = cfg.Debug

	status := atomic.Bool{}

	var lastErr atomic.Pointer[string]
	lastErr.Store(&[]string{"not started yet"}[0])
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if !status.Load() {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(*lastErr.Load()))
		}
	})
	go func() {
		err := http.ListenAndServe(cfg.HealthcheckAddress, nil)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}()

	for {
		if err := runMQTTClient(cfg, &status); err != nil {
			errStr := err.Error()
			lastErr.Store(&errStr)
			fmt.Printf("Error: %v\n", err)
			fmt.Println("Waiting 15 seconds before retrying...")
			status.Store(false)
			time.Sleep(15 * time.Second)
		}
	}
}

func runMQTTClient(cfg Config, status *atomic.Bool) error {
	fmt.Printf("Connecting to MQTT broker at %s:%d...\n", cfg.Host, cfg.Port)

	opts := MQTT.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", cfg.Host, cfg.Port))
	opts.SetUsername(cfg.Username)
	opts.SetPassword(cfg.Password)
	opts.SetWill(cfg.Topic+"/status", "offline", 1, true)

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	fmt.Println("Connected!")

	publishHomeAssistantConfig(client, cfg.Topic)

	tick := time.Tick(time.Minute / 2)
	changes, err := watcher(cfg)
	if err != nil {
		return fmt.Errorf("watchfile: %w", err)
	}

	for {
		select {
		case <-tick:
			if debug {
				fmt.Println("loop: tick")
			}
		case <-changes:
			if debug {
				fmt.Println("loop: watchfile")
			}
		}
		if debug {
			fmt.Printf("Reading airtags data from %s...\n", cfg.AirtagsDataFile)
		}

		airtags, err := readAirtagsData(cfg.AirtagsDataFile)
		if err != nil {
			return err
		}

		if debug {
			fmt.Printf("Publishing MQTT messages for %d airtags...\n", len(airtags))
		}

		for _, airtag := range airtags {
			forwardLocation(client, cfg, airtag)
		}
		status.Store(true)
	}
}

func publishHomeAssistantConfig(client MQTT.Client, mqttTopic string) {
	configPayload := map[string]interface{}{
		"name":    "Mac Airtag To MQTT",
		"uniq_id": mqttTopic + "_connectivity",
		"stat_t":  mqttTopic + "/status",
		"dev_cla": "connectivity",
		"pl_on":   "online",
		"pl_off":  "offline",
	}

	jsonPayload, _ := json.Marshal(configPayload)
	client.Publish(
		fmt.Sprintf("homeassistant/binary_sensor/%s/connectivity/config", mqttTopic),
		0, false, jsonPayload,
	)
	client.Publish(
		fmt.Sprintf("%s/status", mqttTopic),
		0, false, "online",
	)
}

func forwardLocation(client MQTT.Client, cfg Config, airtag Airtag) {
	stateTopic := fmt.Sprintf("%s/%s/state", cfg.Topic, airtag.Identifier)
	jsonAttributesTopic := fmt.Sprintf("%s/%s/attributes", cfg.Topic, airtag.Identifier)
	haConfigTopic := fmt.Sprintf("homeassistant/device_tracker/%s_%s/config", cfg.Topic, airtag.Identifier)

	name := airtag.Name
	if strings.HasSuffix(name, "Bud") {
		name = fmt.Sprintf("%s - %s", cfg.AirpodsName, name)
	} else {
		name = fmt.Sprintf("AirTag - %s", name)
	}

	location := airtag.Location
	address := airtag.Address

	isHome := address.StreetName == cfg.HomeStreetName &&
		strings.HasPrefix(address.StreetAddress, cfg.HomeStreetAddress)

	if debug {
		fmt.Printf("=> %s: %s\n %#v", haConfigTopic, name, airtag)
	}

	configPayload := map[string]interface{}{
		"state_topic":           stateTopic,
		"name":                  name,
		"unique_id":             fmt.Sprintf("%s_%s", cfg.Topic, airtag.Identifier),
		"payload_home":          "home",
		"payload_not_home":      "not_home",
		"json_attributes_topic": jsonAttributesTopic,
	}
	jsonConfig, _ := json.Marshal(configPayload)
	client.Publish(haConfigTopic, 0, false, jsonConfig)

	state := "not_home"
	if isHome {
		state = "home"
	}
	client.Publish(stateTopic, 0, false, state)

	attributes := map[string]interface{}{
		"latitude":     location.Latitude,
		"longitude":    location.Longitude,
		"gps_accuracy": location.HorizontalAccuracy,
		"address":      address.MapItemFullAddress,
		"device_type":  "Apple AirTag",
	}
	jsonAttributes, err := json.Marshal(attributes)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	client.Publish(jsonAttributesTopic, 0, false, jsonAttributes)
}
