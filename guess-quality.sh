#!/usr/bin/env bash

gofmt -s -w .

report="$(ineffassign .)"
if [[ ! -z "$report" ]]; then
  echo "Found inefficient assignments with ineffassign"
  echo "$report"
  exit 2
fi

report="$(go vet ./...)"
if [[ ! -z "$report" ]]; then
  echo "Problems found by go vet"
  echo "$report"
  exit 2
fi

report="$(~/go/bin/gometalinter --vendor --aggregate --exclude=markdown --exclude=asciidoc --exclude="should have comment" ./...)"
if [[ ! -z "$report" ]]; then
  echo "Problems found by gometalinter"
  echo "$report"
  exit 2
fi
