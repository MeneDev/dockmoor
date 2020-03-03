FROM circleci/golang:1.14.0 AS builder-base
ARG PROJECT_USERNAME
ARG PROJECT_REPONAME

RUN GO111MODULE=on go get -u github.com/tcnksm/ghr

RUN GO111MODULE=on go get github.com/mitchellh/gox@v1.0.1

RUN GO111MODULE=on go get github.com/mattn/goveralls@v0.0.5

RUN sudo apt-get install ruby \
    && sudo gem install asciidoctor

RUN wget https://github.com/jgm/pandoc/releases/download/2.4/pandoc-2.4-1-amd64.deb \
    && sudo dpkg -i pandoc-2.4-1-amd64.deb

RUN curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(go env GOPATH)/bin v1.17.1

RUN echo "/go/src/github.com/${PROJECT_USERNAME}/${PROJECT_REPONAME}"
RUN mkdir -p "/go/src/github.com/${PROJECT_USERNAME}/${PROJECT_REPONAME}"
WORKDIR "/go/src/github.com/${PROJECT_USERNAME}/${PROJECT_REPONAME}"

ENV GO111MODULE on

FROM builder-base

COPY --chown=circleci go.mod .
COPY --chown=circleci go.sum .

RUN go mod download
