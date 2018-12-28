// Most lines were copied from the 'github.com/aws/amazon-ecs-agent' repo

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"

	"github.com/fsouza/go-dockerclient"
)

const (
	containerMetadataEnvVar = "ECS_CONTAINER_METADATA_URI"
	interval                = 5 * time.Second
	maxRetries              = 4
	durationBetweenRetries  = time.Second
)

var taskMetadata TaskResponse

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

// LimitsResponse defines the schema for task/cpu limits response
// JSON object
type LimitsResponse struct {
	CPU    *float64 `json:"CPU,omitempty"`
	Memory *int64   `json:"Memory,omitempty"`
}

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

func getTaskMetadata(client *http.Client, taskMetadataEndpoint string) (*TaskResponse, error) {
	var err error
	body, err := metadataResponse(client, taskMetadataEndpoint)
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

func getTaskStats(client *http.Client, taskStatsEndpoint string) (map[string]*docker.Stats, error) {
	body, err := metadataResponse(client, taskStatsEndpoint)
	if err != nil {
		return nil, err
	}

	//fmt.Printf("received task stats: %s \n", string(body))

	var taskStats map[string]*docker.Stats
	err = json.Unmarshal(body, &taskStats)
	if err != nil {
		return nil, fmt.Errorf("task stats: unable to parse response body: %v", err)
	}

	return taskStats, nil
}

func isPauseContainer(containerMetadata ContainerResponse) bool {
	return containerMetadata.Type == "CNI_PAUSE"
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

func putMetrics(client *cloudwatch.CloudWatch, value float64, clusterName, dimension string) error {
	_, err := client.PutMetricData(&cloudwatch.PutMetricDataInput{
		Namespace: aws.String("ECS/Containers"),
		MetricData: []*cloudwatch.MetricDatum{
			&cloudwatch.MetricDatum{
				MetricName: aws.String("MemoryUtilization"),
				Unit:       aws.String("Percent"),
				Value:      aws.Float64(value),
				Dimensions: []*cloudwatch.Dimension{
					&cloudwatch.Dimension{
						Name:  aws.String("ContainerName"),
						Value: aws.String(dimension),
					},
					&cloudwatch.Dimension{
						Name:  aws.String("ClusterName"),
						Value: aws.String(clusterName),
					},
				},
			},
		},
	})
	return err
}

func main() {
	// Wait for the Health information to be ready
	time.Sleep(5 * time.Second)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	v3BaseEndpoint := os.Getenv(containerMetadataEnvVar)
	taskMetadataPath := v3BaseEndpoint + "/task"
	taskStatsPath := v3BaseEndpoint + "/task/stats"
	containerIDToNameMap := make(map[string]string)

	fmt.Print("waiting for the task to be ready\n")
	for {
		if t, err := getTaskMetadata(client, taskMetadataPath); err != nil {
			fmt.Fprintf(os.Stderr, "unable to get task metadata: %v\n", err)
			os.Exit(1)
		} else {
			taskMetadata = *t
		}
		// Wait for the task is ready
		if taskMetadata.KnownStatus == "RUNNING" {
			break
		}
		time.Sleep(time.Second)
	}

	// init CloudWatch client
	awsRegion := strings.Split(taskMetadata.TaskARN, ":")[3]
	fmt.Printf("detected aws region: %v\n", awsRegion)
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(awsRegion),
	}))
	svc := cloudwatch.New(sess)

	// store the pause container's ID if the task is running with awsvpc networking mode
	pauseContainerId := ""
	for _, con := range taskMetadata.Containers {
		if isPauseContainer(con) {
			pauseContainerId = con.ID
			fmt.Print("detected the awsvpc networking mode is enabled\n")
		}
		containerIDToNameMap[con.ID] = con.DockerName
	}

	sigs := make(chan os.Signal, 1)
	quit := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	var memPercent = 0.0
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				if stat, err := getTaskStats(client, taskStatsPath); err != nil {
					fmt.Fprintf(os.Stderr, "unable to get task stats: %v\n", err)
				} else {
					for key, con := range stat {
						// We won't put the CNI pause container's metrics into CW
						if key == pauseContainerId || con == nil {
							continue
						}
						memPercent = 0.0
						if con.MemoryStats.Limit != 0 {
							memPercent = float64(con.MemoryStats.Usage) / float64(con.MemoryStats.Limit) * 100.0
						}
						if err := putMetrics(svc, memPercent, taskMetadata.Cluster, containerIDToNameMap[key]); err != nil {
							fmt.Fprintf(
								os.Stderr,
								"unable to put metrics: dimension [%v], val [%v], err [%v]\n",
								containerIDToNameMap[key], memPercent, err)
						}
					}
				}
			case sig := <-sigs:
				fmt.Printf("signal recieved: %d\n", sig)
				ticker.Stop()
				quit <- true
				return
			}
		}
	}()
	fmt.Printf("taskmetadata-cloudwatch is up and running, awaiting termination signal\n")
	<-quit
	fmt.Printf("exiting")
}
