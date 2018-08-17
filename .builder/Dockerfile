FROM circleci/golang:1.10.3

RUN go get -u github.com/tcnksm/ghr \
    && go get -u github.com/stevenmatthewt/semantics

RUN curl -L -s https://github.com/golang/dep/releases/download/v0.5.0/dep-linux-amd64 -o /go/bin/dep \
    && chmod +x /go/bin/dep

RUN mkdir -p /go/src/github.com/mitchellh/gox
RUN git clone --branch master --depth 1 https://github.com/mitchellh/gox.git /go/src/github.com/mitchellh/gox
RUN cd /go/src/github.com/mitchellh/gox && git reset --hard e05df8dccde397491cd731693a93d4506b5850d3
RUN cd /go/src/github.com/mitchellh/gox && dep ensure && go install -v ./...

RUN wget https://raw.githubusercontent.com/mle86/man-to-md/v0.9.0/man-to-md.pl -o bin/man-to-md.pl
RUN chmod +x bin/man-to-md.pl