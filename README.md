Amazon ECS Task Metadata :point_right: CloudWatch Custom Metrics
=

TODO: Write more about usage including required IAM role permissions

This is an example project which puts your container-level metrics into CloudWatch on Amazon ECS and AWS Fargate.

Pre-built docker image is available at a [Docker Hub repository](https://cloud.docker.com/repository/docker/toricls/ecs-taskmetadata-cloudwatch) to try this quickly.

To try, just run the pre-built container as a sidecar of your application container.

NOTE
- Fargate Platform Version v1.3.0 or later is required to work correctly with Fargate tasks
- For EC2 launch type, this sample requires v1.21.0 or later of the Amazon ECS container agent
- Windows container is not supported by this project for now

:camera: screenshots :point_down:

![Metrics](https://raw.githubusercontent.com/wiki/toricls/ecs-taskmetadata-cloudwatch/imgs/cw-metrics-1.png)

--
![Metrics](https://raw.githubusercontent.com/wiki/toricls/ecs-taskmetadata-cloudwatch/imgs/cw-metrics-2.png)

--
![Metrics](https://raw.githubusercontent.com/wiki/toricls/ecs-taskmetadata-cloudwatch/imgs/cw-metrics-3.png)

