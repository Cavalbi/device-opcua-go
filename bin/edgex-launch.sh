#!/bin/bash
#
# Copyright (c) 2018
# Mainflux
#
# SPDX-License-Identifier: Apache-2.0
#

###
# Launches all EdgeX Go binaries (must be previously built).
#
# Expects that Consul and MongoDB are already installed and running.
#
###

DIR=$PWD
CMD=../cmd

function cleanup {
	pkill edgex-device-opcua
}

cd $CMD
export EDGEX_SECURITY_SECRET_STORE=false
exec -a edgex-device-opcua ./device-opcua/device-opcua &
cd $DIR


trap cleanup EXIT

while : ; do sleep 1 ; done