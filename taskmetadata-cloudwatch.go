package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"

	"github.com/toricls/ecs-taskmetadata-cloudwatch/pkg/cw"
	"github.com/toricls/ecs-taskmetadata-cloudwatch/pkg/ecs"
)

const (
	interval = 10 * time.Second
)

var taskMetadata ecs.TaskResponse

func main() {
	// Wait for the Health information to be ready
	time.Sleep(5 * time.Second)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	containerIDToNameMap := make(map[string]string)

	fmt.Print("waiting for the task to be ready\n")
	for {
		if t, err := ecs.GetTaskMetadata(client); err != nil {
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
		if ecs.IsPauseContainer(con) {
			pauseContainerId = con.ID
			fmt.Print("detected the awsvpc networking mode is enabled\n")
		}
		containerIDToNameMap[con.ID] = con.DockerName
	}

	sigs := make(chan os.Signal, 1)
	quit := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				if taskStats, err := ecs.GetTaskStats(client); err != nil {
					fmt.Fprintf(os.Stderr, "unable to get task stats: %v\n", err)
				} else {
					var d []*cloudwatch.MetricDatum
					clusterName := taskMetadata.Cluster
					var containerName string
					for key, conStats := range taskStats {
						// We ignore the CNI pause container's metrics or a no-stats container
						if key == pauseContainerId || conStats == nil {
							continue
						}
						containerName = containerIDToNameMap[key]
						if data, _ := cw.GetMemoryUtilization(conStats, clusterName, containerName); data != nil {
							d = append(d, data)
						}
						if data, _ := cw.GetCpuUtilization(conStats, clusterName, containerName); data != nil {
							d = append(d, data)
						}
					}
					if len(d) == 0 {
						fmt.Print("nothing to report for now\n")
						continue
					}
					// TODO: Validate the data size is under `40 KB for HTTP POST requests`.
					//  see https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/cloudwatch_limits.html
					if err := cw.PutMetrics(svc, d...); err != nil {
						fmt.Fprintf(os.Stderr, "unable to put metrics: err [%v]\n", err)
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
