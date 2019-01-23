#!/bin/sh

echo "this is version 1.0.0 for image a" > file1
docker build -f DockerfileA_1 -t testimagea:1.0.0 -t testimagea:1.0 -t testimagea:1 -t testimagea:mainline -t testimagea:latest .
#docker push testimagea:1.0.0
#docker push testimagea:1.0
#docker push testimagea:1
#docker push testimagea:latest
#docker push testimagea:mainline

echo "this is version 1.0.1 for image a" > file1
docker build -f DockerfileA_1 -t testimagea:1.0.1 -t testimagea:1.0 -t testimagea:1 -t testimagea:mainline -t testimagea:latest .
#docker push testimagea:1.0.1
#docker push testimagea:1.0
#docker push testimagea:1
#docker push testimagea:latest
#docker push testimagea:mainline

echo "this is version 1.1.0 for image a" > file1
docker build -f DockerfileA_1 -t testimagea:1.1.0 -t testimagea:1.1 -t testimagea:1 -t testimagea:mainline -t testimagea:latest .
#docker push testimagea:1.1.0
#docker push testimagea:1.1
#docker push testimagea:1
#docker push testimagea:latest
#docker push testimagea:mainline

echo "this is version 1.1.1 for image a" > file1
docker build -f DockerfileA_1 -t testimagea:1.1.1 -t testimagea:1.1 -t testimagea:1 -t testimagea:mainline -t testimagea:latest .
#docker push testimagea:1.1.1
#docker push testimagea:1.1
#docker push testimagea:1
#docker push testimagea:latest
#docker push testimagea:mainline

echo "this is version 2.0.0 for image a" > file1
docker build -f DockerfileA_1 -t testimagea:2.0.0 -t testimagea:2.0 -t testimagea:2 -t testimagea:edge -t testimagea:registry-only -t testimagea:latest .
#docker push testimagea:2.0.0
#docker push testimagea:2.0
#docker push testimagea:2
#docker push testimagea:latest
#docker push testimagea:edge
#docker push testimagea:registry-only
#
#docker rmi testimagea:registry-only