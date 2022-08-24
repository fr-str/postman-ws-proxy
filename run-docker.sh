#!/bin/bash
mkdir -p ~/postman-proxy
docker build -t postman-ws-proxy .
# change volume path if you want
docker run -p 8008:8008 \
 -v ~/postman-proxy:/app/log-files/ \
  --name pp postman-ws-proxy