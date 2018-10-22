FROM golang:1.10.1
COPY . /go/src/github.com/andock/ssh2docksal
WORKDIR /go/src/github.com/andock/ssh2docksal
RUN make
ENTRYPOINT ["/go/src/github.com/andock/ssh2docksal/ssh2docksal"]
