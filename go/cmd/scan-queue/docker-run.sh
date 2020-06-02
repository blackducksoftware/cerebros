#!/bin/bash

config_path=$1

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
# echo $DIR

docker run -ti -v "${DIR}"/"$config_path":/conf.json -p 8900:4100 gcr.io/eng-dev/blackducksoftware/cerebros/scan-queue:master ./scan-queue conf.json
# -v: copy in the config file
# -p: map the port
# call the executable with the config file

# then: curl http://localhost:8900/model
