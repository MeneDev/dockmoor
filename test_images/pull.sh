#!/bin/sh

docker pull menedev/testimagea:1
docker pull menedev/testimagea:1.0
docker pull menedev/testimagea:1.0.0
docker pull menedev/testimagea:1.0.1
docker pull menedev/testimagea:1.1.0
docker pull menedev/testimagea:1.1
docker pull menedev/testimagea:1.1.1
docker pull menedev/testimagea:2
docker pull menedev/testimagea:2.0
docker pull menedev/testimagea:2.0.0
docker pull menedev/testimagea:edge
docker pull menedev/testimagea:latest
docker pull menedev/testimagea:mainline
docker pull menedev/testimagea:registry-only

docker rmi menedev/testimagea:registry-only