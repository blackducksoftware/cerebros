#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
# echo $DIR

docker run -ti -v ${DIR}/conf.json:/conf.json -p 8901:4101 gcr.io/eng-dev/blackducksoftware/cerebros/polaris-cli:latest ./polaris-cli conf.json
# -v: copy in the config file
# -p: map the port
# call the executable with the config file

# then: curl http://localhost:8901
