Amazon ECS Task Metadata :point_right: CloudWatch Custom Metrics
=

This is an example project which puts your container-level metrics into CloudWatch on Amazon ECS and AWS Fargate.

Pre-build docker image is available at a [Docker Hub repository](https://cloud.docker.com/repository/docker/toricls/ecs-taskmetadata-cloudwatch) to try quickly.

NOTE
- Fargate Platform Version 1.3 or later is required to work correctly with Fargate tasks
- Windows container is not supported by this project for now

:camera: screenshots :point_down:

![Metrics](https://raw.githubusercontent.com/wiki/toricls/ecs-taskmetadata-cloudwatch/imgs/cw-metrics-1.png)

--
![Metrics](https://raw.githubusercontent.com/wiki/toricls/ecs-taskmetadata-cloudwatch/imgs/cw-metrics-2.png)

--
![Metrics](https://raw.githubusercontent.com/wiki/toricls/ecs-taskmetadata-cloudwatch/imgs/cw-metrics-3.png)

TODO: Write more about usage including required IAM role permissions
