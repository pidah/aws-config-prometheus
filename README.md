Simple Go microservice app for retrieving AWS Config metrics and pushing to prometheus push gateway
---------------------------------------------------------------------------------------------------
To build a docker image of the project just run the following:

docker build .

and then launch the container with the image built above:


docker run -e PROMETHEUS_ENDPOINT=http://host.docker.internal:9091 -e AWS_ACCESS_KEY_ID=XXXXXXXXX -e AWS_SECRET_ACCESS_KEY=XXXXXXXXXXXXXX -e AWS_REGION=eu-west-1 -e AWS_SESSION_TOKEN=XXXXXXXXXXXXXXXXX  -p8080:8080  $docker_image



Ideally AWS credentials should not be used. IAM instance profiles should be used instead.


## The microservice exposes an endpoint at /awsconfig
To trigger the microservice to retrieve metrics from AWS Config and push to the prometheus pushgateway, run the following:

curl localhost:8080/awsconfig


## For more details look at the docker log output



