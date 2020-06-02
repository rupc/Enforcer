#!/bin/bash

# Build Audit System Chain Code(ASCC) image as a shanred library

PATH_TO_PLUGIN=$(pwd) 
PATH_TO_FABRIC="/home/jyr/go/src/github.com/hyperledger/fabric"
PLUGIN_NAME="ascc"
PLUGIN_NAME_FILE=$PLUGIN_NAME.so

if [[ -f $PLUGIN_NAME_FILE ]]; then
    sudo rm $PLUGIN_NAME_FILE
fi

echo "Build ascc.so plugin..."

docker run -i --rm  -v $PATH_TO_PLUGIN:/opt/gopath/src/github.com/$PLUGIN_NAME -w /opt/gopath/src/github.com/$PLUGIN_NAME -v $PATH_TO_FABRIC:/opt/gopath/src/github.com/hyperledger/fabric hyperledger/fabric-baseimage:latest  go build -buildmode=plugin 

# go build -buildmode=plugin

# Need only once
(cd $GOPATH/src/github.com/hyperledger/fabric; DOCKER_DYNAMIC_LINK=true GO_TAGS+=" pluginsenabled" make peerd -Bj;)

