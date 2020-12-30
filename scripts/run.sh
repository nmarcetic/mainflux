#!/bin/bash
# Copyright (c) Mainflux
# SPDX-License-Identifier: Apache-2.0

###
# Runs all Mainflux microservices (must be previously built and installed).
#
# Expects that PostgreSQL and needed messaging DB are alredy running.
# Additionally, MQTT microservice demands that Redis is up and running.
#
###

BUILD_DIR=../build

# Kill all mainflux-* stuff
function cleanup {
    pkill mainflux
    pkill nats
}

###
# NATS
###
gnatsd &
counter=1
until nc -zv localhost 4222 1>/dev/null 2>&1;
do
    sleep 0.5
    ((counter++))
    if [ ${counter} -gt 10 ]
    then
        echo -ne "gnatsd failed to start in 5 sec, exiting"
        exit 1
    fi
    echo -ne "Waiting for gnatsd"
done

###
# Users
###
MF_USERS_LOG_LEVEL=info MF_EMAIL_TEMPLATE=../docker/users/emailer/templates/email.tmpl $BUILD_DIR/mainflux-users &

###
# Things
###
MF_THINGS_LOG_LEVEL=info MF_THINGS_HTTP_PORT=8182 MF_THINGS_AUTH_GRPC_PORT=8183 MF_THINGS_AUTH_HTTP_PORT=8989 $BUILD_DIR/mainflux-things &

###
# HTTP
###
MF_HTTP_ADAPTER_LOG_LEVEL=info MF_HTTP_ADAPTER_PORT=8185 MF_THINGS_AUTH_GRPC_URL=localhost:8183 $BUILD_DIR/mainflux-http &

###
# MQTT
###
MF_MQTT_ADAPTER_LOG_LEVEL=info MF_THINGS_AUTH_GRPC_URL=localhost:8183 $BUILD_DIR/mainflux-mqtt &

###
# CoAP
###
MF_COAP_ADAPTER_LOG_LEVEL=info MF_COAP_ADAPTER_PORT=5683 MF_THINGS_AUTH_GRPC_URL=localhost:8183 $BUILD_DIR/mainflux-coap &

###
# AUTHN
###
MF_AUTH_LOG_LEVEL=debug MF_AUTH_HTTP_PORT=8189 MF_AUTH_GRPC_PORT=8181 MF_AUTH_DB_PORT=5432 MF_AUTH_DB_USER=mainflux MF_AUTH_DB_PASS=mainflux MF_AUTH_DB=auth MF_AUTH_SECRET=secret $BUILD_DIR/mainflux-auth &

trap cleanup EXIT

while : ; do sleep 1 ; done
