#!/bin/sh

echo "this is version 1.0.0 for image a" > file1
docker build -f DockerfileA_1 -t localhost:5000/menedev/testimagea:1.0.0 -t localhost:5000/menedev/testimagea:1.0 -t localhost:5000/menedev/testimagea:1 -t localhost:5000/menedev/testimagea:mainline -t localhost:5000/menedev/testimagea:latest .
docker push localhost:5000/menedev/testimagea:1.0.0
docker push localhost:5000/menedev/testimagea:1.0
docker push localhost:5000/menedev/testimagea:1
docker push localhost:5000/menedev/testimagea:latest
docker push localhost:5000/menedev/testimagea:mainline

echo "this is version 1.0.1 for image a" > file1
docker build -f DockerfileA_1 -t localhost:5000/menedev/testimagea:1.0.1 -t localhost:5000/menedev/testimagea:1.0 -t localhost:5000/menedev/testimagea:1 -t localhost:5000/menedev/testimagea:mainline -t localhost:5000/menedev/testimagea:latest .
docker push localhost:5000/menedev/testimagea:1.0.1
docker push localhost:5000/menedev/testimagea:1.0
docker push localhost:5000/menedev/testimagea:1
docker push localhost:5000/menedev/testimagea:latest
docker push localhost:5000/menedev/testimagea:mainline

echo "this is version 1.1.0 for image a" > file1
docker build -f DockerfileA_1 -t localhost:5000/menedev/testimagea:1.1.0 -t localhost:5000/menedev/testimagea:1.1 -t localhost:5000/menedev/testimagea:1 -t localhost:5000/menedev/testimagea:mainline -t localhost:5000/menedev/testimagea:latest .
docker push localhost:5000/menedev/testimagea:1.1.0
docker push localhost:5000/menedev/testimagea:1.1
docker push localhost:5000/menedev/testimagea:1
docker push localhost:5000/menedev/testimagea:latest
docker push localhost:5000/menedev/testimagea:mainline

echo "this is version 1.1.1 for image a" > file1
docker build -f DockerfileA_1 -t localhost:5000/menedev/testimagea:1.1.1 -t localhost:5000/menedev/testimagea:1.1 -t localhost:5000/menedev/testimagea:1 -t localhost:5000/menedev/testimagea:mainline -t localhost:5000/menedev/testimagea:latest .
docker push localhost:5000/menedev/testimagea:1.1.1
docker push localhost:5000/menedev/testimagea:1.1
docker push localhost:5000/menedev/testimagea:1
docker push localhost:5000/menedev/testimagea:latest
docker push localhost:5000/menedev/testimagea:mainline

echo "this is version 2.0.0 for image a" > file1
docker build -f DockerfileA_1 -t localhost:5000/menedev/testimagea:2.0.0 -t localhost:5000/menedev/testimagea:2.0 -t localhost:5000/menedev/testimagea:2 -t localhost:5000/menedev/testimagea:edge -t localhost:5000/menedev/testimagea:registry-only -t localhost:5000/menedev/testimagea:latest .
docker push localhost:5000/menedev/testimagea:2.0.0
docker push localhost:5000/menedev/testimagea:2.0
docker push localhost:5000/menedev/testimagea:2
docker push localhost:5000/menedev/testimagea:latest
docker push localhost:5000/menedev/testimagea:edge
docker push localhost:5000/menedev/testimagea:registry-only

docker rmi localhost:5000/menedev/testimagea:registry-only