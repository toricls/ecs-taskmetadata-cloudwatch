package cw

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

const (
	nameSpace = "ECS/Containers"

	metricNameMemoryUtilization = "MemoryUtilization"
)

type metricData struct {
	MetricName string
	Unit       string
	Value      float64
	Dimensions []*cloudwatch.Dimension
}

func PutMemoryUtilization(client *cloudwatch.CloudWatch, value float64, clusterName, containerName string) error {

	input := &metricData{
		MetricName: metricNameMemoryUtilization,
		Unit:       cloudwatch.StandardUnitPercent,
		Value:      value,
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
	return putMetrics(client, input)
}

func putMetrics(client *cloudwatch.CloudWatch, input *metricData) error {
	_, err := client.PutMetricData(&cloudwatch.PutMetricDataInput{
		Namespace: aws.String(nameSpace),
		MetricData: []*cloudwatch.MetricDatum{
			&cloudwatch.MetricDatum{
				MetricName: aws.String(input.MetricName),
				Unit:       aws.String(input.Unit),
				Value:      aws.Float64(input.Value),
				Dimensions: input.Dimensions,
			},
		},
	})
	return err
}
