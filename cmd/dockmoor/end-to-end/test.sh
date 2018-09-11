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
#tag::containsAnyInFolder[]
find some-folder/ -type f -exec dockmoor contains --any {} \; -print
#end::containsAnyInFolder[]
) >$RESULTS/containsAnyInFolder.stdout 2>$RESULTS/containsAnyInFolder.stderr
exitCode=$?
[ $exitCode -eq 0 ] || fail 3 "Unexpected exit code $exitCode"
stdout="$(cat $RESULTS/containsAnyInFolder.stdout)"
stderr="$(cat $RESULTS/containsAnyInFolder.stderr)"
[[ $stdout = *"some-folder/Dockerfile-nginx-latest"* ]] || fail 3 "Unexpected stdout"
[[ $stdout = *"some-folder/Dockerfile-nginx-untagged"* ]] || fail 3 "Unexpected stdout"
[[ $stdout = *"some-folder/Dockerfile-nginx-1.15.3"* ]] || fail 3 "Unexpected stdout"
[[ $stdout = *"some-folder/subfolder/Dockerfile-nginx-latest"* ]] || fail 3 "Unexpected stdout"
[[ -z $stderr ]] || fail 3 "Expected empty stderr"
echo $exitCode >$RESULTS/containsAnyInFolder.exitCode


( # find any file with latest/no tag and no digest
#tag::containsLatestInFolder[]
find some-folder/ -type f -exec dockmoor contains --latest {} \; -print
#end::containsLatestInFolder[]
) >$RESULTS/containsLatestInFolder.stdout 2>$RESULTS/containsLatestInFolder.stderr
exitCode=$?
[ $exitCode -eq 0 ] || fail 7 "Unexpected exit code $exitCode"
stdout="$(cat $RESULTS/containsLatestInFolder.stdout)"
stderr="$(cat $RESULTS/containsLatestInFolder.stderr)"
[[ $stdout = *"some-folder/Dockerfile-nginx-latest"* ]] || fail 7 "Unexpected stdout"
[[ $stdout = *"some-folder/Dockerfile-nginx-untagged"* ]] || fail 7 "Unexpected stdout"
[[ $stdout = *"some-folder/subfolder/Dockerfile-nginx-latest"* ]] || fail 7 "Unexpected stdout"
[[ -z $stderr ]] || fail 7 "Expected empty stderr"
echo $exitCode >$RESULTS/containsLatestInFolder.exitCode

( # test if file is of supported format
#tag::containsAnyTestFormatInvalid[]
dockmoor contains --any some-folder/NotADockerfile
#end::containsAnyTestFormatInvalid[]
) >$RESULTS/containsAnyTestFormatInvalid.stdout 2>$RESULTS/containsAnyTestFormatInvalid.stderr
exitCode=$?
echo $exitCode >$RESULTS/containsAnyTestFormatInvalid.exitCode
[ $exitCode -eq 4 ] || fail 4 "Unexpected exit code $exitCode"

( # test if file is of supported format
#tag::containsAnyTestFormatValid[]
dockmoor contains --any some-folder/Dockerfile-nginx-latest
#end::containsAnyTestFormatValid[]
) >$RESULTS/containsAnyTestFormatValid.stdout 2>$RESULTS/containsAnyTestFormatValid.stderr
exitCode=$?
echo $exitCode >$RESULTS/containsAnyTestFormatValid.exitCode
[ $exitCode -eq 0 ] || fail 5 "Unexpected exit code $exitCode"

( # calling contains with non existing file exits with 5
# tag::containsAnyNonExisting[]
dockmoor contains --any nonExisting
# end::containsAnyNonExisting[]
) >$RESULTS/containsAnyNonExisting.stdout 2>$RESULTS/containsAnyNonExisting.stderr
exitCode=$?
echo $exitCode >$RESULTS/containsAnyNonExisting.exitCode
[ $exitCode -eq 5 ] || fail 6 "Unexpected exit code $exitCode"

echo "All tests passed!"
