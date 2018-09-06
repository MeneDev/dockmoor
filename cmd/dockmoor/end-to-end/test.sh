#!/usr/bin/env bash
set -v

function fail() {
    echo "$2"
    exit $1
}

(cd .. && go build -ldflags="-s -w" -o end-to-end/dockmoor || exit 1) || fail 1 "Error building"

PATH=$PATH:.
RESULTS=results

mkdir -p $RESULTS || fail 2 "Cannot create $RESULTS folder"

( # find any file with supported format (i.e. Dockerfile) in folder and subfolder
#tag::findAnyInFolder[]
find some-folder/ -type f -exec dockmoor find --any {} \; -print
#end::findAnyInFolder[]
) >$RESULTS/findAnyInFolder.stdout 2>$RESULTS/findAnyInFolder.stderr
exitCode=$?
echo $exitCode >$RESULTS/findAnyInFolder.exitCode
[ $exitCode -eq 0 ] || fail 3 "Unexpected exit code $exitCode"

( # test if file is of supported format
#tag::findAnyTestFormatInvalid[]
dockmoor find --any some-folder/NotADockerfile
#end::findAnyTestFormatInvalid[]
) >$RESULTS/findAnyTestFormatInvalid.stdout 2>$RESULTS/findAnyTestFormatInvalid.stderr
exitCode=$?
echo $exitCode >$RESULTS/findAnyTestFormatInvalid.exitCode
[ $exitCode -eq 4 ] || fail 4 "Unexpected exit code $exitCode"

( # test if file is of supported format
#tag::findAnyTestFormatValid[]
dockmoor find --any some-folder/Dockerfile-nginx-latest
#end::findAnyTestFormatValid[]
) >$RESULTS/findAnyTestFormatValid.stdout 2>$RESULTS/findAnyTestFormatValid.stderr
exitCode=$?
echo $exitCode >$RESULTS/findAnyTestFormatValid.exitCode
[ $exitCode -eq 0 ] || fail 5 "Unexpected exit code $exitCode"

( # calling find with non existing file exits with 5
# tag::findAnyNonExisting[]
dockmoor find --any nonExisting
# end::findAnyNonExisting[]
) >$RESULTS/findAnyNonExisting.stdout 2>$RESULTS/findAnyNonExisting.stderr
exitCode=$?
echo $exitCode >$RESULTS/findAnyNonExisting.exitCode
[ $exitCode -eq 5 ] || fail 6 "Unexpected exit code $exitCode"

