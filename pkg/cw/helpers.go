package cw

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/docker/docker/api/types"
	"github.com/toricls/ecs-taskmetadata-cloudwatch/pkg/docker"
)

const (
	nameSpace = "ECS/Containers"

	metricNameMemoryUtilization = "MemoryUtilization"
	metricNameCPUUtilization    = "CPUUtilization"
)

func GetMemoryUtilization(stats *types.Stats, clusterName, containerName string) (*cloudwatch.MetricDatum, error) {
	value := docker.CalculateMemUtilization(stats)
	d := &cloudwatch.MetricDatum{
		MetricName: aws.String(metricNameMemoryUtilization),
		Unit:       aws.String(cloudwatch.StandardUnitPercent),
		Value:      aws.Float64(value),
		Dimensions: []*cloudwatch.Dimension{
			&cloudwatch.Dimension{
				Name:  aws.String("ClusterName"),
				Value: aws.String(clusterName),
			},
			&cloudwatch.Dimension{
				Name:  aws.String("ContainerName"),
				Value: aws.String(containerName),
			},
		},
	}
	return d, nil
}

func GetCpuUtilization(stats *types.Stats, clusterName, containerName string) (*cloudwatch.MetricDatum, error) {
	value := docker.CalculateCpuUtilization(stats)
	d := &cloudwatch.MetricDatum{
		MetricName: aws.String(metricNameCPUUtilization),
		Unit:       aws.String(cloudwatch.StandardUnitPercent),
		Value:      aws.Float64(value),
		Dimensions: []*cloudwatch.Dimension{
			&cloudwatch.Dimension{
				Name:  aws.String("ClusterName"),
				Value: aws.String(clusterName),
			},
			&cloudwatch.Dimension{
				Name:  aws.String("ContainerName"),
				Value: aws.String(containerName),
			},
		},
	}
	return d, nil
}

func PutMetrics(client *cloudwatch.CloudWatch, input ...*cloudwatch.MetricDatum) error {
	_, err := client.PutMetricData(&cloudwatch.PutMetricDataInput{
		Namespace:  aws.String(nameSpace),
		MetricData: input,
	})
	return err
}
