package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/imroc/req"
)

type InputDetail struct {
	Name       string `json:"name"`
	Identifier string `json:"id"`
	State      string `json:"state"`
	Message    string `json:"message"`
}

type InputOverview []struct {
	ID string `json:"id"`
}

type CloudService struct {
	Name  string `json:"service"`
	State string `json:"state"`
}

type MQTTRestResponse struct {
	Default    string `json:"defaultBroker"`
	Kubernetes string `json:"kubernetes"`
}

type MQTTServer struct {
	Name  string
	State string
}

type SystemResponse struct {
	FreeMemory float64 `json:"freeMemory"`
	Disks      []struct {
		DriveLetter string  `json:"disk"`
		FreeSpace   float64 `json:"freeSpace"`
	} `json:"disks"`
}

type StatusResponse struct {
	State             string  `json:"state"`
	Message           string  `json:"message"`
	NbrOfLoggedErrors float64 `json:"nbrOfLoggedErrors"`
	RedundancyState   struct {
		State      string `json:"state"`
		Current    string `json:"current"`
		Configured string `json:"configured"`
	} `json:"redundancyState"`
}

// CheckState compares two string values and returns either 0 oder 1
func CheckState(value string, state string) float64 {
	if value == state {
		return 1
	} else {
		return 0
	}
}

// GetValue searches in the message field for a possible numerical value
func (i InputDetail) GetValue() (float64, error) {
	r, _ := regexp.Compile("[0-9]*,[0-9[0-9]]*")
	if v := r.FindString(i.Message); v != "" {

		// Convert float from german to american notation
		v = strings.ReplaceAll(v, ",", ".")

		f, _ := strconv.ParseFloat(v, 64)
		return f, nil
	} else {
		return 0, errors.New("no value in message")
	}
}

// QueryInputs returns a list with information about all alarm inputs
func QueryInputs(hostname string, AccessKey string) (*[]InputDetail, error) {
	authHeader := req.Header{
		"Accept":        "application/json",
		"Authorization": AccessKey,
	}
	var inputDetails []InputDetail
	var inputOverview InputOverview

	r, err := req.Get(fmt.Sprintf("%s/rest/monitoring/input", hostname), authHeader)
	if err != nil {
		return nil, err
	}

	err = r.ToJSON(&inputOverview)
	if err != nil {
		return nil, err
	}

	for _, input := range inputOverview {
		r, err := req.Get(fmt.Sprintf("%s/rest/monitoring/input/%s", hostname, input.ID), authHeader)
		if err != nil {
			return nil, err
		}

		var detail InputDetail
		err = r.ToJSON(&detail)
		if err != nil {
			return nil, err
		}

		detail.Identifier = input.ID
		inputDetails = append(inputDetails, detail)
	}

	return &inputDetails, nil
}

// QueryCloudServices returns a list of all cloudservices
func QueryCloudServices(hostname string, AccessKey string) (*[]CloudService, error) {
	authHeader := req.Header{
		"Accept":        "application/json",
		"Authorization": AccessKey,
	}
	var services []CloudService

	r, err := req.Get(fmt.Sprintf("%s/rest/monitoring/cloud", hostname), authHeader)
	if err != nil {
		return nil, err
	}

	err = r.ToJSON(&services)
	if err != nil {
		return nil, err
	}

	return &services, nil
}

// QueryMQTTServer returns a list of all mqtt brokers
func QueryMQTTServer(hostname string, AccessKey string) (*[]MQTTServer, error) {
	authHeader := req.Header{
		"Accept":        "application/json",
		"Authorization": AccessKey,
	}
	var resp MQTTRestResponse

	r, err := req.Get(fmt.Sprintf("%s/rest/monitoring/mqtt", hostname), authHeader)
	if err != nil {
		return nil, err
	}

	err = r.ToJSON(&resp)
	if err != nil {
		return nil, err
	}

	// Convert data structure from single object to list of objects
	// to fit the structure of all other endpoints
	mqttServer := []MQTTServer{
		MQTTServer{
			Name:  "defaultBroker",
			State: resp.Default,
		},
		MQTTServer{
			Name:  "kubernetes",
			State: resp.Kubernetes,
		},
	}

	return &mqttServer, nil
}

// QuerySystem returns information about memory and disc space
func QuerySystem(hostname string, AccessKey string) (*SystemResponse, error) {
	authHeader := req.Header{
		"Accept":        "application/json",
		"Authorization": AccessKey,
	}
	var resp SystemResponse

	r, err := req.Get(fmt.Sprintf("%s/rest/monitoring/system", hostname), authHeader)
	if err != nil {
		return nil, err
	}

	err = r.ToJSON(&resp)
	if err != nil {
		return nil, err
	}

	// Convert MB back to base unit
	// MB -> KB -> Byte
	resp.FreeMemory *= (1024 * 1024)

	for idx := range resp.Disks {
		// Convert GB back to base unit
		// GB -> MB -> KB -> Byte
		resp.Disks[idx].FreeSpace *= (1024 * 1024 * 1024)
		resp.Disks[idx].DriveLetter = strings.Split(resp.Disks[idx].DriveLetter, ":")[0]
	}

	return &resp, nil
}

// QueryStatus returns status information about the FE2 software
func QueryStatus(hostname string, AccessKey string) (*StatusResponse, error) {
	authHeader := req.Header{
		"Accept":        "application/json",
		"Authorization": AccessKey,
	}
	var resp StatusResponse

	r, err := req.Get(fmt.Sprintf("%s/rest/monitoring/status", hostname), authHeader)
	if err != nil {
		return nil, err
	}

	err = r.ToJSON(&resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
