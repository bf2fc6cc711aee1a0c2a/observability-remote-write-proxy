#!/usr/bin/env sh

PROMETHEUS_NAME="prometheus"
: "${PROMETHEUS_IMAGE:="docker.io/prom/prometheus:latest"}"

PROMETHEUS_NAME_PROXY="prometheus_proxy"

PROMETHEUS_PORT=9090

RUNTIME="docker"
OPTS=""

if command -v podman &> /dev/null
then
  RUNTIME=podman
  OPTS="--security-opt label=disable"
fi

if ! command -v $RUNTIME &> /dev/null
then
  echo "neither podman nor docker were found, aborting"
  exit 1
fi

echo "detected runtime is $RUNTIME, starting containers. this can take some time..."

# Create a basic prometheus config file
cat > prometheus.yml <<- EOF
global:
  scrape_interval: 10s
  evaluation_interval: 30s
remote_write:
  - name: proxy
    url: http://localhost:8080
    oauth2:
      client_id: <client_id>
      client_secret: <client_secret>
      token_url: <token_url>
EOF

# Create a basic prometheus config file for the proxy target
cat > prometheusProxy.yml <<- EOF
global:
  scrape_interval: 10s
  evaluation_interval: 30s
EOF

# start the prometheus container
echo "starting prometheus container from image $PROMETHEUS_IMAGE"
$RUNTIME run -d --rm --name $PROMETHEUS_NAME --network=host $OPTS -v $(pwd)/prometheus.yml:/etc/prometheus/prometheus.yml $PROMETHEUS_IMAGE --web.enable-remote-write-receiver --config.file=/etc/prometheus/prometheus.yml
PROMETHEUS_CONTAINER=`$RUNTIME ps -f name=$PROMETHEUS_NAME --format "{{.ID}}"`

cleanup() {
  echo "stopping prometheus container $PROMETHEUS_CONTAINER"
  $RUNTIME stop "$PROMETHEUS_CONTAINER" &> /dev/null
  rm -f prometheus.yml

  echo "stopping prometheus container $PROMETHEUS_CONTAINER_PROXY"
  $RUNTIME stop "$PROMETHEUS_CONTAINER_PROXY" &> /dev/null
  rm -f prometheusProxy.yml

  echo "done"
  exit 0
}

trap 'cleanup' SIGINT
echo "Prometheus is running on port $PROMETHEUS_PORT"
echo "press ctrl-c to quit"
while true; do
  sleep 1
done