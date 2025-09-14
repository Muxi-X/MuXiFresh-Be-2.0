#!/bin/bash

source /etc/profile

cd /MuxiFresh-Be-2.0/userauth

chmod +x ./accountCenter

nohup ./accountCenter &

chmod +x ./user-auth

./user-auth
