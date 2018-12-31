// The original version of this file was
// https://github.com/aws/amazon-ecs-agent/blob/44c6b7612777e865e0c2bf8e2eb7871317ce228e/misc/v3-task-endpoint-validator/v3-task-endpoint-validator.go

package ecs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/docker/docker/api/types"
)

const (
	containerMetadataEnvVar = "ECS_CONTAINER_METADATA_URI"
	maxRetries              = 4
	durationBetweenRetries  = time.Second
)

var taskMetadataPath string
var taskStatsPath string

func init() {
	v3BaseEndpoint := os.Getenv(containerMetadataEnvVar)
	taskMetadataPath = v3BaseEndpoint + "/task"
	taskStatsPath = v3BaseEndpoint + "/task/stats"
}

// TaskResponse defines the schema for the task response JSON object
type TaskResponse struct {
	Cluster            string              `json:"Cluster"`
	TaskARN            string              `json:"TaskARN"`
	Family             string              `json:"Family"`
	Revision           string              `json:"Revision"`
	DesiredStatus      string              `json:"DesiredStatus,omitempty"`
	KnownStatus        string              `json:"KnownStatus"`
	AvailabilityZone   string              `json:"AvailabilityZone"`
	Containers         []ContainerResponse `json:"Containers,omitempty"`
	Limits             *LimitsResponse     `json:"Limits,omitempty"`
	PullStartedAt      *time.Time          `json:"PullStartedAt,omitempty"`
	PullStoppedAt      *time.Time          `json:"PullStoppedAt,omitempty"`
	ExecutionStoppedAt *time.Time          `json:"ExecutionStoppedAt,omitempty"`
}

// ContainerResponse defines the schema for the container response
// JSON object
type ContainerResponse struct {
	ID            string            `json:"DockerId"`
	Name          string            `json:"Name"`
	DockerName    string            `json:"DockerName"`
	Image         string            `json:"Image"`
	ImageID       string            `json:"ImageID"`
	Ports         []PortResponse    `json:"Ports,omitempty"`
	Labels        map[string]string `json:"Labels,omitempty"`
	DesiredStatus string            `json:"DesiredStatus"`
	KnownStatus   string            `json:"KnownStatus"`
	ExitCode      *int              `json:"ExitCode,omitempty"`
	Limits        LimitsResponse    `json:"Limits"`
	CreatedAt     *time.Time        `json:"CreatedAt,omitempty"`
	StartedAt     *time.Time        `json:"StartedAt,omitempty"`
	FinishedAt    *time.Time        `json:"FinishedAt,omitempty"`
	Type          string            `json:"Type"`
	Networks      []Network         `json:"Networks,omitempty"`
	Health        HealthStatus      `json:"Health,omitempty"`
}

// LimitsResponse defines the schema for task/cpu limits response
// JSON object
type LimitsResponse struct {
	CPU    *float64 `json:"CPU,omitempty"`
	Memory *int64   `json:"Memory,omitempty"`
}

// HealthStatus defines the schema for health status response
// JSON object
type HealthStatus struct {
	Status   string     `json:"status,omitempty"`
	Since    *time.Time `json:"statusSince,omitempty"`
	ExitCode int        `json:"exitCode,omitempty"`
	Output   string     `json:"output,omitempty"`
}

// PortResponse defines the schema for portmapping response JSON
// object.
type PortResponse struct {
	ContainerPort uint16 `json:"ContainerPort,omitempty"`
	Protocol      string `json:"Protocol,omitempty"`
	HostPort      uint16 `json:"HostPort,omitempty"`
}

// Network is a struct that keeps track of metadata of a network interface
type Network struct {
	NetworkMode   string   `json:"NetworkMode,omitempty"`
	IPv4Addresses []string `json:"IPv4Addresses,omitempty"`
	IPv6Addresses []string `json:"IPv6Addresses,omitempty"`
}

// IsPauseContainer returns true if this container is a "pause container"
func IsPauseContainer(containerMetadata ContainerResponse) bool {
	return containerMetadata.Type == "CNI_PAUSE"
}

// GetTaskMetadata returns the ECS task's metadata by making the api call to
// the Task Metadata endpoint v3
func GetTaskMetadata(client *http.Client) (*TaskResponse, error) {
	var err error
	body, err := metadataResponse(client, taskMetadataPath)
	if err != nil {
		return nil, err
	}

	//fmt.Printf("received task metadata: %s \n", string(body))

	var taskMetadata TaskResponse
	if err = json.Unmarshal(body, &taskMetadata); err != nil {
		return nil, fmt.Errorf("unable to parse response body: %v", err)
	}

	return &taskMetadata, nil
}

// GetTaskMetadata returns stats of the ECS task's containers by making the api
// call to the Task Metadata endpoint v3
func GetTaskStats(client *http.Client) (map[string]*types.Stats, error) {
	body, err := metadataResponse(client, taskStatsPath)
	if err != nil {
		return nil, err
	}

	//fmt.Printf("received task stats: %s \n", string(body))

	var taskStats map[string]*types.Stats
	err = json.Unmarshal(body, &taskStats)
	if err != nil {
		return nil, fmt.Errorf("task stats: unable to parse response body: %v", err)
	}

	return taskStats, nil
}

func metadataResponse(client *http.Client, endpoint string) ([]byte, error) {
	var resp []byte
	var err error
	for i := 0; i < maxRetries; i++ {
		resp, err = metadataResponseOnce(client, endpoint)
		if err == nil {
			return resp, nil
		}
		fmt.Fprintf(os.Stderr, "Attempt [%d/%d]: unable to get metadata response from '%s': %v",
			i, maxRetries, endpoint, err)
		time.Sleep(durationBetweenRetries)
	}

	return nil, err
}

func metadataResponseOnce(client *http.Client, endpoint string) ([]byte, error) {
	resp, err := client.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("unable to get response: %v", err)
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("incorrect status code  %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body: %v", err)
	}

	return body, nil
}
