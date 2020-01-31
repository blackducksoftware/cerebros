#!/bin/bash

function fileExists() {
    if [ "$( find . -name "$1" | wc -l | sed 's/^ *//' )" == "0" ];
    then return 1;
    else return 0;
    fi
}

function installPython() {
    apt-get update -y && \
    apt-get install python3-pip idle3 -y && \
    pip3 install --no-cache-dir --upgrade pip && \
    \
    # delete cache and tmp files
    apt-get clean && \
    apt-get autoclean && \
    rm -rf /var/cache/* && \
    rm -rf /tmp/* && \
    rm -rf /var/tmp/* && \
    rm -rf /var/lib/apt/lists/* && \
    \
    # make some useful symlinks that are expected to exist
    cd /usr/bin && \
    ln -s idle3 idle && \
    ln -s pydoc3 pydoc && \
    ln -s python3 python && \
    ln -s python3-config python-config && \
    cd /

    wget https://repo.continuum.io/archive/Anaconda3-5.0.1-Linux-x86_64.sh && \
    bash Anaconda3-5.0.1-Linux-x86_64.sh -b && \
    rm Anaconda3-5.0.1-Linux-x86_64.sh

    # Set path to conda
    echo PATH=/root/anaconda3/bin:$PATH
}

if fileExists "Pipfile" || fileExists "Pipfile.lock" || fileExists "setup.py" || fileExists "requirements.txt" || fileExists "environment.yml";
then installPython;
fi
