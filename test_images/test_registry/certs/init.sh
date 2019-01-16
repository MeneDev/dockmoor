#!/usr/bin/env bash
./easyrsa init-pki
dd if=/dev/urandom of=pki/.rnd bs=256 count=1
./easyrsa --batch build-ca nopass
#./easyrsa gen-dh

./easyrsa build-server-full registry.localhost nopass
