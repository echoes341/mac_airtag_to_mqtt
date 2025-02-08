package main

import (
	"encoding/json"
	"os"
)

func readAirtagsData(path string) ([]Airtag, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var airtags []Airtag
	if err := json.Unmarshal(data, &airtags); err != nil {
		return nil, err
	}

	return airtags, nil
}

type DeviceLocation struct {
	Latitude           float64 `json:"latitude,omitempty"`
	Longitude          float64 `json:"longitude,omitempty"`
	HorizontalAccuracy float64 `json:"horizontalAccuracy,omitempty"`
}

type DeviceAddress struct {
	StreetName         string `json:"streetName,omitempty"`
	StreetAddress      string `json:"streetAddress,omitempty"`
	MapItemFullAddress string `json:"mapItemFullAddress,omitempty"`
}

type Airtag struct {
	Identifier string         `json:"identifier"`
	Name       string         `json:"name"`
	Location   DeviceLocation `json:"location"`
	Address    DeviceAddress  `json:"address"`
}
