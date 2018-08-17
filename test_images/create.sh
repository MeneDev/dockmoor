#!/bin/sh

echo "this is version 1.0.0 for image a" > file1
docker build -f DockerfileA_1 -t menedev/testimagea:1.0.0 -t menedev/testimagea:latest .
docker push menedev/testimagea:1.0.0
docker push menedev/testimagea:latest

echo "this is version 1.0.1 for image a" > file1
docker build -f DockerfileA_1 -t menedev/testimagea:1.0.1 -t menedev/testimagea:latest .
docker push menedev/testimagea:1.0.1
docker push menedev/testimagea:latest

echo "this is version 1.1.0 for image a" > file1
docker build -f DockerfileA_1 -t menedev/testimagea:1.1.0 -t menedev/testimagea:latest .
docker push menedev/testimagea:1.1.0
docker push menedev/testimagea:latest

echo "this is version 1.1.1 for image a" > file1
docker build -f DockerfileA_1 -t menedev/testimagea:1.1.1 -t menedev/testimagea:latest .
docker push menedev/testimagea:1.1.1
docker push menedev/testimagea:latest

echo "this is version 2.0.0 for image a" > file1
docker build -f DockerfileA_1 -t menedev/testimagea:2.0.0 -t menedev/testimagea:latest .
docker push menedev/testimagea:2.0.0
docker push menedev/testimagea:latest
