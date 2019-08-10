#!/usr/bin/env bash

gofmt -s -w .

report="$(golangci-lint run)"
if [[ ! -z "$report" ]]; then
  echo "Problems found by gometalinter"
  echo "$report"
  exit 2
fi
