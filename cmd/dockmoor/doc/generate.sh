#!/usr/bin/env bash

set -v
ROOT=$(git rev-parse --show-toplevel)
asciidoctor -b docbook5 _readme.adoc
pandoc -f docbook -t asciidoc --wrap=none --columns=120 _readme.xml -o ${ROOT}/README.adoc

VARS="$(cat _vars.adoc)"
BODY="$(cat ${ROOT}/README.adoc)"
HEADER="$(cat _header.adoc)"

echo "$VARS
:branch: $(git symbolic-ref --short HEAD)

$HEADER

$BODY
" > ${ROOT}/README.adoc
