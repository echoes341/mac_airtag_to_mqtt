package main

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

func loadYAML(r io.Reader) (Config, error) {
	cfg := Config{}
	if err := yaml.NewDecoder(r).Decode(&cfg); err != nil {
		return Config{}, err
	}
	cfg.Topic = fmt.Sprintf("mac_airtag_to_mqtt_%s", cfg.TopicName)
	if cfg.AirtagsDataFile == "" {
		cfg.AirtagsDataFile = fmt.Sprintf("/Users/%s/Library/Caches/com.apple.findmy.fmipcore/Items.data", cfg.MacUser)
	}
	return cfg, nil
}

type Config struct {
	Host      string `yaml:"MQTT_HOST,omitempty"`
	Port      int    `yaml:"MQTT_PORT,omitempty"`
	TopicName string `yaml:"MQTT_TOPIC_NAME,omitempty"`
	Topic     string `yaml:"-"`
	Username  string `yaml:"MQTT_USERNAME,omitempty"`
	Password  string `yaml:"MQTT_PASSWORD,omitempty"`

	MacUser         string `yaml:"MAC_USER,omitempty"`
	AirtagsDataFile string `yaml:"AIRTAGS_DATA_FILE,omitempty"`

	HomeStreetName    string `yaml:"HOME_STREET_NAME,omitempty"`
	HomeStreetAddress string `yaml:"HOME_STREET_ADDRESS,omitempty"`
	AirpodsName       string `yaml:"AIRPODS_NAME,omitempty"`
}
