#!/usr/bin/env bash

set -v
ROOT=$(git rev-parse --show-toplevel)

(cd ../end-to-end && ./test.sh)
../end-to-end/dockmoor --asciidoc-usage > dockmoor.adoc

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

perl -ni -e 'print "[subs=+macros]\n" if ($_ eq "....\n" && $p eq "\n"); $p=$_; print $_' "${ROOT}/README.adoc"

(
cd ../end-to-end
while read path
do
    path=${path:2}
    if [ -d "../end-to-end/$path" ]; then
        path="$path/"
    fi
    prefix="$(perl -e 'print quotemeta($ARGV[0])' "https://github.com/MeneDev/dockmoor/blob/master/cmd/dockmoor/end-to-end/")"
    path="$(perl -e 'print quotemeta($ARGV[0])' "$path")"
    echo "path: $path"
    perl -pi -e "s/(^| )(${path})( |\n)/\$1$prefix\$2\[\$2\]\$3/g" "${ROOT}/README.adoc"
done <<< "$(find . -mindepth 1 ! -name '*.sh' ! -name 'dockmoor')"
)
