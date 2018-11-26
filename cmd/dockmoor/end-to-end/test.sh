#!/usr/bin/env bash
set -v

function fail() {
    echo "$2 at CASE_ID $1"
    exit $1
}

function hasLine {
    local file="$1"
    local line="$2"
    [[ -z $(echo "$file" | grep -Fx "$line") ]] || return 0
    return 1
}

function hasNoLine {
    local file="$1"
    local line="$2"
    [[ ! -z $(echo "$file" | grep -Fx "$line") ]] || return 0
    return 1
}

(cd .. && go build -ldflags="-s -w" -o end-to-end/dockmoor) || fail 1 "Error building"

PATH=.:$PATH
RESULTS=results

mkdir -p $RESULTS || fail 2 "Cannot create $RESULTS folder"

## list commandd

CASE_ID=1
( # find any file with supported format (i.e. Dockerfile) in folder and subfolder
#tag::listAnyInFolder[]
find some-folder/ -type f -exec dockmoor list {} \; | sort | uniq
#end::listAnyInFolder[]
) >$RESULTS/listAnyInFolder.stdout 2>$RESULTS/listAnyInFolder.stderr
exitCode=$?
[ $exitCode -eq 0 ] || fail ${CASE_ID} "Unexpected exit code $exitCode"
stdout="$(cat $RESULTS/listAnyInFolder.stdout)"
stderr="$(cat $RESULTS/listAnyInFolder.stderr)"
echo "$stdout"
hasLine "$stdout" "nginx" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "nginx:1.15.3" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "nginx:1.15.3-alpine@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "nginx:latest" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "nginx@sha256:db5acc22920799fe387a903437eb89387607e5b3f63cf0f4472ac182d7bad644" || fail ${CASE_ID} "Unexpected stdout"
[[ -z $stderr ]] || fail ${CASE_ID} "Expected empty stderr"
echo $exitCode >$RESULTS/listAnyInFolder.exitCode

CASE_ID=2
( # find any file with latest/no tag and no digest
#tag::listLatestInFolder[]
find some-folder/ -type f -exec dockmoor list --latest {} \; | sort | uniq
#end::listLatestInFolder[]
) >$RESULTS/listLatestInFolder.stdout 2>$RESULTS/listLatestInFolder.stderr
exitCode=$?
[ $exitCode -eq 0 ] || fail ${CASE_ID} "Unexpected exit code $exitCode"
stdout="$(cat $RESULTS/listLatestInFolder.stdout)"
stderr="$(cat $RESULTS/listLatestInFolder.stderr)"
hasLine "$stdout" "nginx" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "nginx:latest" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "nginx:1.15.3" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "nginx:1.15.3-alpine@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "nginx@sha256:db5acc22920799fe387a903437eb89387607e5b3f63cf0f4472ac182d7bad644" || fail ${CASE_ID} "Unexpected stdout: $stdout"
[[ -z $stderr ]] || fail ${CASE_ID} "Expected empty stderr"
echo $exitCode >$RESULTS/listLatestInFolder.exitCode

CASE_ID=3
( # find any file with latest/no tag
#tag::listUnpinnedInFolder[]
find some-folder/ -type f -exec dockmoor list --unpinned {} \; | sort | uniq
#end::listUnpinnedInFolder[]
) >$RESULTS/listUnpinnedInFolder.stdout 2>$RESULTS/listUnpinnedInFolder.stderr
exitCode=$?
[ $exitCode -eq 0 ] || fail ${CASE_ID} "Unexpected exit code $exitCode"
stdout="$(cat $RESULTS/listUnpinnedInFolder.stdout)"
stderr="$(cat $RESULTS/listUnpinnedInFolder.stderr)"
hasLine "$stdout" "nginx" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "nginx:latest" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "nginx:1.15.3" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "nginx:1.15.3-alpine@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "nginx@sha256:db5acc22920799fe387a903437eb89387607e5b3f63cf0f4472ac182d7bad644" || fail ${CASE_ID} "Unexpected stdout: $stdout"
[[ -z $stderr ]] || fail ${CASE_ID} "Expected empty stderr"
echo $exitCode >$RESULTS/listUnpinnedInFolder.exitCode

CASE_ID=4
( # list all image references from file
#tag::listAnyInFile[]
dockmoor list Dockerfile
#end::listAnyInFile[]
) >$RESULTS/listAnyInFile.stdout 2>$RESULTS/listAnyInFile.stderr
exitCode=$?
[ $exitCode -eq 0 ] || fail ${CASE_ID} "Unexpected exit code $exitCode"
stdout="$(cat $RESULTS/listAnyInFile.stdout)"
stderr="$(cat $RESULTS/listAnyInFile.stderr)"
hasLine "$stdout" "image-name" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "image-name:latest" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "image-name:1.12" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "image-name@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "docker.io/library/image-name:1.12@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "docker.io/library/image-name" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "docker.io/library/image-name:latest" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "example.com/image-name:1.12" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "example.com/image-name:latest@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "example.com/image-name@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "example.com/other-image" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "example.com/other-image:latest" || fail ${CASE_ID} "Unexpected stdout"
[[ -z $stderr ]] || fail ${CASE_ID} "Expected empty stderr"
echo $exitCode >$RESULTS/listAnyInFile.exitCode

CASE_ID=5
( # list unpinned image references with latest/no tag from file
#tag::listLatestInFile[]
dockmoor list --latest Dockerfile
#end::listLatestInFile[]
) >$RESULTS/listLatestInFile.stdout 2>$RESULTS/listLatestInFile.stderr
exitCode=$?
[ $exitCode -eq 0 ] || fail ${CASE_ID} "Unexpected exit code $exitCode"
stdout="$(cat $RESULTS/listLatestInFile.stdout)"
stderr="$(cat $RESULTS/listLatestInFile.stderr)"
hasLine "$stdout" "image-name" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "image-name:latest" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "image-name:1.12" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "image-name@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "docker.io/library/image-name:1.12@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "docker.io/library/image-name" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "docker.io/library/image-name:latest" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "example.com/image-name:1.12" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "example.com/image-name:latest@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "example.com/image-name@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "example.com/other-image" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "example.com/other-image:latest" || fail ${CASE_ID} "Unexpected stdout"
[[ -z $stderr ]] || fail ${CASE_ID} "Expected empty stderr"
echo $exitCode >$RESULTS/listLatestInFile.exitCode

CASE_ID=6
( # list unpinned image references
#tag::listUnpinnedInFile[]
dockmoor list --unpinned Dockerfile
#end::listUnpinnedInFile[]
) >$RESULTS/listUnpinnedInFile.stdout 2>$RESULTS/listUnpinnedInFile.stderr
exitCode=$?
[ $exitCode -eq 0 ] || fail ${CASE_ID} "Unexpected exit code $exitCode"
stdout="$(cat $RESULTS/listUnpinnedInFile.stdout)"
stderr="$(cat $RESULTS/listUnpinnedInFile.stderr)"
hasLine "$stdout" "image-name" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "image-name:latest" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "image-name:1.12" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "image-name@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "docker.io/library/image-name:1.12@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "docker.io/library/image-name" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "docker.io/library/image-name:latest" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "example.com/image-name:1.12" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "example.com/image-name:latest@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "example.com/image-name@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "example.com/other-image" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "example.com/other-image:latest" || fail ${CASE_ID} "Unexpected stdout"
echo $exitCode >$RESULTS/listUnpinnedInFile.exitCode


## contains command
CASE_ID=7
( # contains all image references in supported file (i.e. Dockerfile)
#tag::containsAnyInFolder[]
find some-folder/ -type f -exec dockmoor contains {} \; -print
#end::containsAnyInFolder[]
) >$RESULTS/containsAnyInFolder.stdout 2>$RESULTS/containsAnyInFolder.stderr
exitCode=$?
[ $exitCode -eq 0 ] || fail ${CASE_ID} "Unexpected exit code $exitCode"
stdout="$(cat $RESULTS/containsAnyInFolder.stdout)"
stderr="$(cat $RESULTS/containsAnyInFolder.stderr)"
hasLine "$stdout" "some-folder/Dockerfile-nginx-latest" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "some-folder/Dockerfile-nginx-untagged" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "some-folder/Dockerfile-nginx-1.15.3" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "some-folder/subfolder/Dockerfile-nginx-latest" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "some-folder/Dockerfile-nginx-digest" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "some-folder/Dockerfile-nginx-tagged-digest" || fail ${CASE_ID} "Unexpected stdout"
[[ -z $stderr ]] || fail ${CASE_ID} "Expected empty stderr"
echo $exitCode >$RESULTS/containsAnyInFolder.exitCode

CASE_ID=8
CASE_NAME=containsLatestInFolder
( # find any file with latest/no tag and no digest
#tag::containsLatestInFolder[]
find some-folder/ -type f -exec dockmoor contains --latest {} \; -print
#end::containsLatestInFolder[]
) >$RESULTS/$CASE_NAME.stdout 2>$RESULTS/$CASE_NAME.stderr
exitCode=$?
[ $exitCode -eq 0 ] || fail ${CASE_ID} "Unexpected exit code $exitCode"
stdout="$(cat $RESULTS/$CASE_NAME.stdout)"
stderr="$(cat $RESULTS/$CASE_NAME.stderr)"
hasLine "$stdout" "some-folder/Dockerfile-nginx-latest" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "some-folder/Dockerfile-nginx-untagged" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "some-folder/subfolder/Dockerfile-nginx-latest" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "some-folder/subfolder/Dockerfile-nginx-digest" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "some-folder/Dockerfile-nginx-tagged-digest" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "some-folder/Dockerfile-nginx-1.15.3" || fail ${CASE_ID} "Unexpected stdout: $stdout"
[[ -z $stderr ]] || fail ${CASE_ID} "Expected empty stderr"
echo $exitCode >$RESULTS/CASE_NAME.exitCode

CASE_ID=9
( # find any file with latest/no tag
#tag::containsUnpinnedInFolder[]
find some-folder/ -type f -exec dockmoor contains --unpinned {} \; -print
#end::containsUnpinnedInFolder[]
) >$RESULTS/containsUnpinnedInFolder.stdout 2>$RESULTS/containsUnpinnedInFolder.stderr
exitCode=$?
[ $exitCode -eq 0 ] || fail ${CASE_ID} "Unexpected exit code $exitCode"
stdout="$(cat $RESULTS/containsUnpinnedInFolder.stdout)"
stderr="$(cat $RESULTS/containsUnpinnedInFolder.stderr)"
hasLine "$stdout" "some-folder/Dockerfile-nginx-latest" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "some-folder/Dockerfile-nginx-untagged" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "some-folder/subfolder/Dockerfile-nginx-latest" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "some-folder/subfolder/Dockerfile-nginx-digest" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "some-folder/Dockerfile-nginx-tagged-digest" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "some-folder/Dockerfile-nginx-1.15.3" || fail ${CASE_ID} "Unexpected stdout"
[[ -z $stderr ]] || fail ${CASE_ID} "Expected empty stderr"
echo $exitCode >$RESULTS/containsUnpinnedInFolder.exitCode


CASE_ID=10
( # test if file is of supported format
#tag::containsAnyTestFormatInvalid[]
dockmoor contains some-folder/NotADockerfile
#end::containsAnyTestFormatInvalid[]
) >$RESULTS/containsAnyTestFormatInvalid.stdout 2>$RESULTS/containsAnyTestFormatInvalid.stderr
exitCode=$?
echo $exitCode >$RESULTS/containsAnyTestFormatInvalid.exitCode
[ $exitCode -eq 4 ] || fail ${CASE_ID} "Unexpected exit code $exitCode"

CASE_ID=11
( # test if file is of supported format
#tag::containsAnyTestFormatValid[]
dockmoor contains Dockerfile
#end::containsAnyTestFormatValid[]
) >$RESULTS/containsAnyTestFormatValid.stdout 2>$RESULTS/containsAnyTestFormatValid.stderr
exitCode=$?
echo $exitCode >$RESULTS/containsAnyTestFormatValid.exitCode
[ $exitCode -eq 0 ] || fail ${CASE_ID} "Unexpected exit code $exitCode"

CASE_ID=12
( # calling contains with non existing file exits with 5
# tag::containsAnyNonExisting[]
dockmoor contains nonExisting
# end::containsAnyNonExisting[]
) >$RESULTS/containsAnyNonExisting.stdout 2>$RESULTS/containsAnyNonExisting.stderr
exitCode=$?
echo $exitCode >$RESULTS/containsAnyNonExisting.exitCode
[ $exitCode -eq 5 ] || fail ${CASE_ID} "Unexpected exit code $exitCode"

    CASE_ID=13
    CASE_NAME=listDomainInFile
    ( # list image references from example.com
    #tag::listDomainInFile[]
    dockmoor list --domain=example.com Dockerfile
    #end::listDomainInFile[]
    ) >$RESULTS/${CASE_NAME}.stdout 2>$RESULTS/${CASE_NAME}.stderr
    exitCode=$?
    [ $exitCode -eq 0 ] || fail ${CASE_ID} "Unexpected exit code $exitCode"
    stdout="$(cat $RESULTS/${CASE_NAME}.stdout)"
    stderr="$(cat $RESULTS/${CASE_NAME}.stderr)"
    hasNoLine "$stdout" "image-name" || fail ${CASE_ID} "Unexpected stdout"
    hasNoLine "$stdout" "image-name:latest" || fail ${CASE_ID} "Unexpected stdout"
    hasNoLine "$stdout" "image-name:1.12" || fail ${CASE_ID} "Unexpected stdout"
    hasLine "$stdout" "example.com/image-name:1.12" || fail ${CASE_ID} "Unexpected stdout"
    hasLine "$stdout" "example.com/other-image" || fail ${CASE_ID} "Unexpected stdout"
    hasLine "$stdout" "example.com/other-image:latest" || fail ${CASE_ID} "Unexpected stdout"
    [[ -z $stderr ]] || fail ${CASE_ID} "Expected empty stderr"
    echo $exitCode >$RESULTS/${CASE_NAME}.exitCode

CASE_ID=14
CASE_NAME=listDomainWithLatestInFile
( # list image references from example.com with latest or no tag
#tag::listDomainWithLatestInFile[]
dockmoor list --domain=example.com --latest Dockerfile
#end::listDomainWithLatestInFile[]
) >$RESULTS/${CASE_NAME}.stdout 2>$RESULTS/${CASE_NAME}.stderr
exitCode=$?
[ $exitCode -eq 0 ] || fail ${CASE_ID} "Unexpected exit code $exitCode"
stdout="$(cat $RESULTS/${CASE_NAME}.stdout)"
stderr="$(cat $RESULTS/${CASE_NAME}.stderr)"

hasNoLine "$stdout" "image-name" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "image-name:latest" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "image-name:1.12" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "image-name@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "docker.io/library/image-name:1.12@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "docker.io/library/image-name" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "docker.io/library/image-name:latest" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "example.com/image-name:1.12" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "example.com/image-name:latest@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "example.com/image-name@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "example.com/other-image" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "example.com/other-image:latest" || fail ${CASE_ID} "Unexpected stdout"
[[ -z $stderr ]] || fail ${CASE_ID} "Expected empty stderr"
echo $exitCode >$RESULTS/${CASE_NAME}.exitCode

CASE_ID=15
CASE_NAME=listUnpinnedWithLatest
( # list unpinned image references with latest/no tag from file
#tag::listUnpinnedWithLatest[]
dockmoor list --latest --unpinned Dockerfile
#end::listUnpinnedWithLatest[]
) >$RESULTS/${CASE_NAME}.stdout 2>$RESULTS/${CASE_NAME}.stderr
exitCode=$?
[ $exitCode -eq 0 ] || fail ${CASE_ID} "Unexpected exit code $exitCode"
stdout="$(cat $RESULTS/${CASE_NAME}.stdout)"
stderr="$(cat $RESULTS/${CASE_NAME}.stderr)"
hasLine "$stdout" "image-name" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "image-name:latest" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "image-name:1.12" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "image-name@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "docker.io/library/image-name:1.12@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "docker.io/library/image-name" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "docker.io/library/image-name:latest" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "example.com/image-name:1.12" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "example.com/image-name:latest@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout""example.com/image-name@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "example.com/other-image" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "example.com/other-image:latest" || fail ${CASE_ID} "Unexpected stdout"
[[ -z $stderr ]] || fail ${CASE_ID} "Expected empty stderr"
echo $exitCode >$RESULTS/${CASE_NAME}.exitCode

CASE_ID=16
CASE_NAME=listTagsWithRegexMatch
( # list image references where tag ends with -test
#tag::listTagsWithRegexMatch[]
dockmoor list --tag=/-test$/ Dockerfile
#end::listTagsWithRegexMatch[]
) >$RESULTS/${CASE_NAME}.stdout 2>$RESULTS/${CASE_NAME}.stderr
exitCode=$?
[ $exitCode -eq 0 ] || fail ${CASE_ID} "Unexpected exit code $exitCode"
stdout="$(cat $RESULTS/${CASE_NAME}.stdout)"
stderr="$(cat $RESULTS/${CASE_NAME}.stderr)"
hasNoLine "$stdout" "image-name" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "image-name:latest" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "image-name:1.12" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "image-name:1.12-test" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "image-name:1.11-test" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "image-name@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "docker.io/library/image-name:1.12@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "docker.io/library/image-name" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "docker.io/library/image-name:latest" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "docker.io/library/image-name:latest-test" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "example.com/image-name:1.12" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "example.com/image-name:1.12-test" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "example.com/image-name:1.12-testing" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "example.com/image-name:latest@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf" || fail ${CASE_ID} "Unexpected stdout"
hasLine "$stdout" "example.com/image-name:latest-test@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "example.com/image-name@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "example.com/other-image" || fail ${CASE_ID} "Unexpected stdout"
hasNoLine "$stdout" "example.com/other-image:latest" || fail ${CASE_ID} "Unexpected stdout"
[[ -z $stderr ]] || fail ${CASE_ID} "Expected empty stderr"
echo $exitCode >$RESULTS/${CASE_NAME}.exitCode

CASE_ID=17
CASE_NAME=pinWithDockerd
( # pin all image references to same file
rm pin-tests/Dockerfile-testimagea
cp pin-tests/Dockerfile-testimagea.org pin-tests/Dockerfile-testimagea

#tag::pinWithDockerd[]
dockmoor pin pin-tests/Dockerfile-testimagea
#end::pinWithDockerd[]
) >$RESULTS/${CASE_NAME}.stdout 2>$RESULTS/${CASE_NAME}.stderr
[ $exitCode -eq 0 ] || fail ${CASE_ID} "Unexpected exit code $exitCode"
stdout="$(cat $RESULTS/${CASE_NAME}.stdout)"
stderr="$(cat $RESULTS/${CASE_NAME}.stderr)"
[[ -z $stdout ]] || fail ${CASE_ID} "Expected empty stdout"
[[ -z $stderr ]] || fail ${CASE_ID} "Expected empty stderr"
cmp --silent pin-tests/Dockerfile-testimagea-any.expected pin-tests/Dockerfile-testimagea || fail ${CASE_ID} "unexpected result"
# cleanup
rm pin-tests/Dockerfile-testimagea
cp pin-tests/Dockerfile-testimagea.org pin-tests/Dockerfile-testimagea


# When we reach this, everything is fine!
echo "All tests passed!"
