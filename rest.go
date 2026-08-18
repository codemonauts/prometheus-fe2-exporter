package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

var valueRe = regexp.MustCompile(`[0-9]+,[0-9]+`)

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

// get request against the given URL
func get(ctx context.Context, url string, accessKey string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", accessKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d from %s", resp.StatusCode, url)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("decoding response from %s: %w", url, err)
	}

	return nil
}

// CheckState compares two string values and returns either 0 or 1
func CheckState(value string, state string) float64 {
	if value == state {
		return 1
	}
	return 0
}

// GetValue searches in the message field for a possible numerical value
func (i InputDetail) GetValue() (float64, error) {
	v := valueRe.FindString(i.Message)
	if v == "" {
		return 0, errors.New("no value in message")
	}

	// Convert float from german to american notation
	v = strings.ReplaceAll(v, ",", ".")

	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing value %q: %w", v, err)
	}
	return f, nil
}

// QueryInputs returns a list with information about all alarm inputs
func QueryInputs(ctx context.Context, hostname string, accessKey string) (*[]InputDetail, error) {
	var inputDetails []InputDetail
	var inputOverview InputOverview

	if err := get(ctx, fmt.Sprintf("%s/rest/monitoring/input", hostname), accessKey, &inputOverview); err != nil {
		return nil, err
	}

	for _, input := range inputOverview {
		var detail InputDetail
		if err := get(ctx, fmt.Sprintf("%s/rest/monitoring/input/%s", hostname, input.ID), accessKey, &detail); err != nil {
			return nil, err
		}

		detail.Identifier = input.ID
		inputDetails = append(inputDetails, detail)
	}

	return &inputDetails, nil
}

// QueryCloudServices returns a list of all cloudservices
func QueryCloudServices(ctx context.Context, hostname string, accessKey string) (*[]CloudService, error) {
	var services []CloudService

	if err := get(ctx, fmt.Sprintf("%s/rest/monitoring/cloud", hostname), accessKey, &services); err != nil {
		return nil, err
	}

	return &services, nil
}

// QueryMQTTServer returns a list of all mqtt brokers
func QueryMQTTServer(ctx context.Context, hostname string, accessKey string) (*[]MQTTServer, error) {
	var resp MQTTRestResponse

	if err := get(ctx, fmt.Sprintf("%s/rest/monitoring/mqtt", hostname), accessKey, &resp); err != nil {
		return nil, err
	}

	// Convert data structure from single object to list of objects
	// to fit the structure of all other endpoints
	mqttServer := []MQTTServer{
		{
			Name:  "defaultBroker",
			State: resp.Default,
		},
		{
			Name:  "kubernetes",
			State: resp.Kubernetes,
		},
	}

	return &mqttServer, nil
}

// QuerySystem returns information about memory and disc space
func QuerySystem(ctx context.Context, hostname string, accessKey string) (*SystemResponse, error) {
	var resp SystemResponse

	if err := get(ctx, fmt.Sprintf("%s/rest/monitoring/system", hostname), accessKey, &resp); err != nil {
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
func QueryStatus(ctx context.Context, hostname string, accessKey string) (*StatusResponse, error) {
	var resp StatusResponse

	if err := get(ctx, fmt.Sprintf("%s/rest/monitoring/status", hostname), accessKey, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
