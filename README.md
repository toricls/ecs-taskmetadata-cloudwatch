Amazon ECS Task Metadata :point_right: CloudWatch Custom Metrics
=

TODO: Write more about usage including required IAM role permissions

This is an example project which puts your container-level metrics into CloudWatch on Amazon ECS and AWS Fargate.

Pre-built docker image is available at a [Docker Hub repository](https://cloud.docker.com/repository/docker/toricls/ecs-taskmetadata-cloudwatch) to try this quickly.

To try, just run the pre-built container as a sidecar of your application container, then you'll see `CPUUtilization` and `MemoryUtilization` metrics on your CloudWatch console under `ECS/Containers` namespace.

NOTE
- This project uses [Task Metadata Endpoint v3](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-metadata-endpoint-v3.html) and not works with v2
- For Fargate launch type, Fargate Platform Version v1.3.0 or later is required
- For EC2 launch type, v1.21.0 or later of the Amazon ECS container agent is required
- Windows container is not supported for now

:camera: screenshots :point_down:

![Metrics](https://raw.githubusercontent.com/wiki/toricls/ecs-taskmetadata-cloudwatch/imgs/cw-metrics-1.png)

--
![Metrics](https://raw.githubusercontent.com/wiki/toricls/ecs-taskmetadata-cloudwatch/imgs/cw-metrics-2.png)

--
![Metrics](https://raw.githubusercontent.com/wiki/toricls/ecs-taskmetadata-cloudwatch/imgs/cw-metrics-3.png)

## Contribution

Any contributions are welcome :raised_hands:

1. Fork ([https://github.com/toricls/ecs-taskmetadata-cloudwatch/fork](https://github.com/toricls/ecs-taskmetadata-cloudwatch/fork))
1. Create a feature branch
1. Commit your changes
1. Rebase your local changes against the master branch
1. Create a new Pull Request

## Author

[toricls](https://github.com/toricls)
