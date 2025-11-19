#!/bin/bash

nohup ./intro-rpc -f etc/intro-rpc.yaml &

./intro-api -f etc/intro-api.yaml

