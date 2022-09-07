#!/bin/bash
mkdir -p ~/postman-proxy
docker build -t postman-ws-proxy .
# change volume path if you want
docker run \
  --restart=always \
  -v ~/.proxylog:/app/log-files \
  --network=host \
  -d --name pp postman-ws-proxy
