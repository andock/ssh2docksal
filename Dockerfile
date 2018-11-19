FROM golang:1.10.1
COPY . /go/src/github.com/andock/ssh2docksal
WORKDIR /go/src/github.com/andock/ssh2docksal
RUN apt-get update && apt-get install -y libltdl7 && rm -rf /var/lib/apt/lists/*
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
RUN go get -u golang.org/x/lint/golint

# Add packages.
RUN set -xe; \
	apt-get update >/dev/null; \
	apt-get -y --no-install-recommends install >/dev/null \
        apt-transport-https \
        gnupg \
        locales \
        wget \
        apt-utils \
        sudo \
        ; \
    apt-get clean;

RUN set -xe; \
	# Create a regular user/group "docker" (uid = 1000, gid = 999 ) with access to sudo
	groupadd docker -g 999; \
	useradd -m -s /bin/bash -u 1000 -g 999 -G sudo -p docker docker; \
echo 'docker ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers


ENV GOSU_VERSION=1.10 \
	GOMPLATE_VERSION=2.4.0
RUN set -xe; \
	# Install gosu and give access to the docker user primary group to use it.
	# gosu is used instead of sudo to start the main container process (pid 1) in a docker friendly way.
	# https://github.com/tianon/gosu
	curl -fsSL "https://github.com/tianon/gosu/releases/download/${GOSU_VERSION}/gosu-$(dpkg --print-architecture)" -o /usr/local/bin/gosu; \
	chown root:"$(id -gn docker)" /usr/local/bin/gosu; \
	chmod +sx /usr/local/bin/gosu; \
	# gomplate (to process configuration templates in startup.sh)
	curl -fsSL https://github.com/hairyhenderson/gomplate/releases/download/v${GOMPLATE_VERSION}/gomplate_linux-amd64-slim -o /usr/local/bin/gomplate; \
	chmod +x /usr/local/bin/gomplate


ENV \
    # ssh-agent proxy socket (requires docksal/ssh-agent)
    SSH_AUTH_SOCK=/.ssh-agent/proxy-socket \
	TERM=xterm \
	PROJECT_ROOT=/var/www \
	# Default values for HOST_UID and HOST_GUI to match the default Ubuntu user. These are used in startup.sh
	HOST_UID=1000 \
	HOST_GID=1000

RUN make
ENTRYPOINT ["/go/src/github.com/andock/ssh2docksal/ssh2docksal"]
