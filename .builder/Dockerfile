FROM circleci/golang:1.12.7

RUN go get -u github.com/tcnksm/ghr

RUN mkdir -p /go/src/github.com/mitchellh/gox
RUN git clone --branch master https://github.com/mitchellh/gox.git /go/src/github.com/mitchellh/gox
RUN cd /go/src/github.com/mitchellh/gox && git reset --hard d8caaff5a9dc98f4cfa1fcce6e7265a04689f641
RUN cd /go/src/github.com/mitchellh/gox && go install -v ./... && go clean
RUN rm -rf /go/src/github.com/mitchellh/gox

RUN go get github.com/mattn/goveralls

RUN sudo apt-get install ruby \
    && sudo gem install asciidoctor

RUN wget https://github.com/jgm/pandoc/releases/download/2.4/pandoc-2.4-1-amd64.deb \
    && sudo dpkg -i pandoc-2.4-1-amd64.deb

RUN curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(go env GOPATH)/bin v1.17.1
