#!/bin/bash

source ./logging.sh || { echo "Error: logging.sh not found"; exit 1; }

# Set default Grafana version if not provided externally
: "${GRAFANA_VERSION:=latest}"

export GF_AUTH_ANONYMOUS_ENABLED=false

export GF_PATHS_HOME=/data/grafana
export GF_PATHS_DATA=/data/grafana/data
export GF_PATHS_PLUGINS=/data/grafana/plugins

cd ./grafana || exit
run_with_logging "Grafana ${GRAFANA_VERSION}" "${ENABLE_LOGS_GRAFANA:-false}" ./bin/grafana server
