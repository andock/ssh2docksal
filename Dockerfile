FROM golang:1.10.1
COPY . /go/src/github.com/andock/ssh2docksal
WORKDIR /go/src/github.com/andock/ssh2docksal
RUN apt-get update && apt-get install -y libltdl7 && rm -rf /var/lib/apt/lists/*
RUN make
ENTRYPOINT ["/go/src/github.com/andock/ssh2docksal/ssh2docksal"]
