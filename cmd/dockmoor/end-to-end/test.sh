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

## list commandd

( # find any file with supported format (i.e. Dockerfile) in folder and subfolder
#tag::listAnyInFolder[]
find some-folder/ -type f -exec dockmoor list --any {} \; | sort | uniq
#end::listAnyInFolder[]
) >$RESULTS/listAnyInFolder.stdout 2>$RESULTS/listAnyInFolder.stderr
exitCode=$?
[ $exitCode -eq 0 ] || fail 9 "Unexpected exit code $exitCode"
stdout="$(cat $RESULTS/listAnyInFolder.stdout)"
stderr="$(cat $RESULTS/listAnyInFolder.stderr)"
[[ $stdout = *"nginx"* ]] || fail 9 "Unexpected stdout"
[[ $stdout = *"nginx:1.15.3"* ]] || fail 9 "Unexpected stdout"
[[ $stdout = *"nginx:1.15.3-alpine@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf"* ]] || fail 9 "Unexpected stdout"
[[ $stdout = *"nginx:latest"* ]] || fail 9 "Unexpected stdout"
[[ $stdout = *"nginx@sha256:db5acc22920799fe387a903437eb89387607e5b3f63cf0f4472ac182d7bad644"* ]] || fail 9 "Unexpected stdout"
[[ -z $stderr ]] || fail 9 "Expected empty stderr"
echo $exitCode >$RESULTS/listAnyInFolder.exitCode


( # find any file with latest/no tag and no digest
#tag::listLatestInFolder[]
find some-folder/ -type f -exec dockmoor list --latest {} \; | sort | uniq
#end::listLatestInFolder[]
) >$RESULTS/listLatestInFolder.stdout 2>$RESULTS/listLatestInFolder.stderr
exitCode=$?
[ $exitCode -eq 0 ] || fail 10 "Unexpected exit code $exitCode"
stdout="$(cat $RESULTS/listLatestInFolder.stdout)"
stderr="$(cat $RESULTS/listLatestInFolder.stderr)"
[[ $stdout = *"nginx"* ]] || fail 10 "Unexpected stdout"
[[ $stdout = *"nginx:latest"* ]] || fail 10 "Unexpected stdout"
[[ ! $stdout = *"nginx:1.15.3"* ]] || fail 10 "Unexpected stdout"
[[ ! $stdout = *"nginx:1.15.3-alpine@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf"* ]] || fail 10 "Unexpected stdout"
[[ ! $stdout = *"nginx@sha256:db5acc22920799fe387a903437eb89387607e5b3f63cf0f4472ac182d7bad644"* ]] || fail 10 "Unexpected stdout: $stdout"
[[ -z $stderr ]] || fail 10 "Expected empty stderr"
echo $exitCode >$RESULTS/listLatestInFolder.exitCode

( # find any file with latest/no tag
#tag::listUnpinnedInFolder[]
find some-folder/ -type f -exec dockmoor list --unpinned {} \; | sort | uniq
#end::listUnpinnedInFolder[]
) >$RESULTS/listUnpinnedInFolder.stdout 2>$RESULTS/listUnpinnedInFolder.stderr
exitCode=$?
[ $exitCode -eq 0 ] || fail 11 "Unexpected exit code $exitCode"
stdout="$(cat $RESULTS/listUnpinnedInFolder.stdout)"
stderr="$(cat $RESULTS/listUnpinnedInFolder.stderr)"
[[ $stdout = *"nginx"* ]] || fail 11 "Unexpected stdout"
[[ $stdout = *"nginx:latest"* ]] || fail 11 "Unexpected stdout"
[[ $stdout = *"nginx:1.15.3"* ]] || fail 11 "Unexpected stdout"
[[ ! $stdout = *"nginx:1.15.3-alpine@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf"* ]] || fail 11 "Unexpected stdout"
[[ ! $stdout = *"nginx@sha256:db5acc22920799fe387a903437eb89387607e5b3f63cf0f4472ac182d7bad644"* ]] || fail 11 "Unexpected stdout: $stdout"
[[ -z $stderr ]] || fail 11 "Expected empty stderr"
echo $exitCode >$RESULTS/listUnpinnedInFolder.exitCode


( # list all image references from file
#tag::listAnyInFile[]
dockmoor list --any Dockerfile
#end::listAnyInFile[]
) >$RESULTS/listAnyInFile.stdout 2>$RESULTS/listAnyInFile.stderr
exitCode=$?
[ $exitCode -eq 0 ] || fail 12 "Unexpected exit code $exitCode"
stdout="$(cat $RESULTS/listAnyInFile.stdout)"
stderr="$(cat $RESULTS/listAnyInFile.stderr)"
[[ $stdout = *"image-name"* ]] || fail 12 "Unexpected stdout"
[[ $stdout = *"image-name:latest"* ]] || fail 12 "Unexpected stdout"
[[ $stdout = *"image-name:latest@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf"* ]] || fail 12 "Unexpected stdout"
[[ $stdout = *"image-name:1.12"* ]] || fail 12 "Unexpected stdout"
[[ $stdout = *"image-name:1.12@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf"* ]] || fail 12 "Unexpected stdout"
[[ $stdout = *"image-name@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf"* ]] || fail 12 "Unexpected stdout"
[[ $stdout = *"example.com/image-name"* ]] || fail 12 "Unexpected stdout"
[[ $stdout = *"example.com/image-name:latest"* ]] || fail 12 "Unexpected stdout"
[[ $stdout = *"example.com/image-name:latest@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf"* ]] || fail 12 "Unexpected stdout"
[[ $stdout = *"example.com/image-name:1.12"* ]] || fail 12 "Unexpected stdout"
[[ $stdout = *"example.com/image-name:1.12@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf"* ]] || fail 12 "Unexpected stdout"
[[ $stdout = *"example.com/image-name@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf"* ]] || fail 12 "Unexpected stdout"
[[ -z $stderr ]] || fail 12 "Expected empty stderr"
echo $exitCode >$RESULTS/listAnyInFile.exitCode

( # list unpinned image references with latest/no tag from file
#tag::listLatestInFile[]
dockmoor list --latest Dockerfile
#end::listLatestInFile[]
) >$RESULTS/listLatestInFile.stdout 2>$RESULTS/listLatestInFile.stderr
exitCode=$?
[ $exitCode -eq 0 ] || fail 12 "Unexpected exit code $exitCode"
stdout="$(cat $RESULTS/listLatestInFile.stdout)"
stderr="$(cat $RESULTS/listLatestInFile.stderr)"
[[ $stdout = *"image-name"* ]] || fail 12 "Unexpected stdout"
[[ $stdout = *"image-name:latest"* ]] || fail 12 "Unexpected stdout"
[[ $stdout = *"image-name:latest@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf"* ]] || fail 12 "Unexpected stdout"
[[ ! $stdout = *"image-name:1.12"* ]] || fail 12 "Unexpected stdout"
[[ ! $stdout = *"image-name:1.12@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf"* ]] || fail 12 "Unexpected stdout"
[[ ! $stdout = *"image-name@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf"* ]] || fail 12 "Unexpected stdout"
[[ $stdout = *"example.com/image-name"* ]] || fail 12 "Unexpected stdout"
[[ $stdout = *"example.com/image-name:latest"* ]] || fail 12 "Unexpected stdout"
[[ $stdout = *"example.com/image-name:latest@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf"* ]] || fail 12 "Unexpected stdout"
[[ ! $stdout = *"example.com/image-name:1.12"* ]] || fail 12 "Unexpected stdout"
[[ ! $stdout = *"example.com/image-name:1.12@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf"* ]] || fail 12 "Unexpected stdout"
[[ ! $stdout = *"example.com/image-name@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf"* ]] || fail 12 "Unexpected stdout"
[[ -z $stderr ]] || fail 12 "Expected empty stderr"
echo $exitCode >$RESULTS/listLatestInFile.exitCode

( # list unpinned image references
#tag::listUnpinnedInFile[]
dockmoor list --unpinned Dockerfile
#end::listUnpinnedInFile[]
) >$RESULTS/listUnpinnedInFile.stdout 2>$RESULTS/listUnpinnedInFile.stderr
exitCode=$?
[ $exitCode -eq 0 ] || fail 13 "Unexpected exit code $exitCode"
stdout="$(cat $RESULTS/listUnpinnedInFile.stdout)"
stderr="$(cat $RESULTS/listUnpinnedInFile.stderr)"
[[ $stdout = *"image-name"* ]] || fail 13 "Unexpected stdout"
[[ $stdout = *"image-name:latest"* ]] || fail 13 "Unexpected stdout"
[[ ! $stdout = *"image-name:latest@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf"* ]] || fail 13 "Unexpected stdout"
[[ $stdout = *"image-name:1.12"* ]] || fail 13 "Unexpected stdout"
[[ ! $stdout = *"image-name:1.12@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf"* ]] || fail 13 "Unexpected stdout"
[[ ! $stdout = *"image-name@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf"* ]] || fail 13 "Unexpected stdout"
[[ $stdout = *"example.com/image-name"* ]] || fail 13 "Unexpected stdout"
[[ $stdout = *"example.com/image-name:latest"* ]] || fail 13 "Unexpected stdout"
[[ ! $stdout = *"example.com/image-name:latest@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf"* ]] || fail 13 "Unexpected stdout"
[[ $stdout = *"example.com/image-name:1.12"* ]] || fail 13 "Unexpected stdout"
[[ ! $stdout = *"example.com/image-name:1.12@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf"* ]] || fail 13 "Unexpected stdout"
[[ ! $stdout = *"example.com/image-name@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf"* ]] || fail 13 "Unexpected stdout"
[[ -z $stderr ]] || fail 13 "Expected empty stderr"
echo $exitCode >$RESULTS/listUnpinnedInFile.exitCode


## contains command

( # contains all image references in supported file (i.e. Dockerfile)
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
[[ $stdout = *"some-folder/Dockerfile-nginx-digest"* ]] || fail 3 "Unexpected stdout"
[[ $stdout = *"some-folder/Dockerfile-nginx-tagged-digest"* ]] || fail 3 "Unexpected stdout"
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
[[ ! $stdout = *"some-folder/subfolder/Dockerfile-nginx-digest"* ]] || fail 7 "Unexpected stdout"
[[ ! $stdout = *"some-folder/Dockerfile-nginx-tagged-digest"* ]] || fail 7 "Unexpected stdout"
[[ ! $stdout = *"some-folder/Dockerfile-nginx-1.15.3"* ]] || fail 7 "Unexpected stdout: $stdout"
[[ -z $stderr ]] || fail 7 "Expected empty stderr"
echo $exitCode >$RESULTS/containsLatestInFolder.exitCode

( # find any file with latest/no tag
#tag::containsUnpinnedInFolder[]
find some-folder/ -type f -exec dockmoor contains --unpinned {} \; -print
#end::containsUnpinnedInFolder[]
) >$RESULTS/containsUnpinnedInFolder.stdout 2>$RESULTS/containsUnpinnedInFolder.stderr
exitCode=$?
[ $exitCode -eq 0 ] || fail 8 "Unexpected exit code $exitCode"
stdout="$(cat $RESULTS/containsUnpinnedInFolder.stdout)"
stderr="$(cat $RESULTS/containsUnpinnedInFolder.stderr)"
[[ $stdout = *"some-folder/Dockerfile-nginx-latest"* ]] || fail 8 "Unexpected stdout"
[[ $stdout = *"some-folder/Dockerfile-nginx-untagged"* ]] || fail 8 "Unexpected stdout"
[[ $stdout = *"some-folder/subfolder/Dockerfile-nginx-latest"* ]] || fail 8 "Unexpected stdout"
[[ ! $stdout = *"some-folder/subfolder/Dockerfile-nginx-digest"* ]] || fail 8 "Unexpected stdout"
[[ ! $stdout = *"some-folder/Dockerfile-nginx-tagged-digest"* ]] || fail 8 "Unexpected stdout"
[[ $stdout = *"some-folder/Dockerfile-nginx-1.15.3"* ]] || fail 8 "Unexpected stdout"
[[ -z $stderr ]] || fail 8 "Expected empty stderr"
echo $exitCode >$RESULTS/containsUnpinnedInFolder.exitCode


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
