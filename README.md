# ssh2docksal
:wrench: SSH server to user based docksal container access

> SSH chroot with steroids

```ruby
┌────────────┐
│bobby@laptop│
└────────────┘
       │
       └──ssh project@mycorp.biz──┐
                                     ▼
                               ┌──────────┐
┌──────────────────────────────┤ssh2docksal├──┐
│                              └──────────┘  │
│              docker exec -it       │       │
│                 container1         │       │
│          ┌──────/bin/bash──────────┘       │
│ ┌────────┼───────────────────────────────┐ │
│ │docker  │                               │ │
│ │┌───────▼──┐ ┌──────────┐ ┌──────────┐  │ │
│ ││container1│ │container2│ │container3│  │ │
│ │└──────────┘ └──────────┘ └──────────┘  │ │
│ └────────────────────────────────────────┘ │
└────────────────────────────────────────────┘
```

## Usage

```
NAME:
   ssh2docksal - SSH portal to Docker containers

USAGE:
   ssh2docksal [global options] command [command options] [arguments...]

COMMANDS:
   help, h	Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --verbose, -V                 Enable verbose mode
   --syslog-server               Configure a syslog server, i.e: udp://localhost:514
   --bind, -b ":2222"            Listen to address
   --host-key, -k "built-in"     Path or complete SSH host key to use, use 'system' for keys in /etc/ssh
   --shell "/bin/sh"             DEFAULT shell
   --docker-run-args "-it --rm"  'docker run' arguments
   --no-join                     Do not join existing containers, always create new ones
   --clean-on-startup            Cleanup Docker containers created by ssh2docksal on start
   --banner 			         Display a banner on connection
   --help, -h			         show help
   --version, -v		         print the version
```

## Example

Server

```console
$ ssh2docksal
INFO[0000] Listening on port 2222
INFO[0001] NewClient (0): User="alpine", ClientVersion="5353482d322e302d4f70656e5353485f362e362e317031205562756e74752d327562756e747532"
INFO[0748] NewClient (1): User="ubuntu", ClientVersion="5353482d322e302d4f70656e5353485f362e362e317031205562756e74752d327562756e747532"
```

Client

```console
$ ssh localhost -p 2222 -l alpine
Host key fingerprint is 59:46:d7:cf:ca:33:be:1f:58:fd:46:c8:ca:5d:56:03
+--[ RSA 2048]----+
|          . .E   |
|         . .  o  |
|          o    +.|
|         +   . .*|
|        S    .oo=|
|           . oB+.|
|            oo.+o|
|              ...|
|              .o.|
+-----------------+

alpine@localhost's password:
/ # cat /etc/alpine-release
3.2.0
/ # ^D
```

```console
$ ssh localhost -p 2222 -l ubuntu
Host key fingerprint is 59:46:d7:cf:ca:33:be:1f:58:fd:46:c8:ca:5d:56:03
+--[ RSA 2048]----+
|          . .E   |
|         . .  o  |
|          o    +.|
|         +   . .*|
|        S    .oo=|
|           . oB+.|
|            oo.+o|
|              ...|
|              .o.|
+-----------------+

ubuntu@localhost's password:
# lsb_release -a
No LSB modules are available.
Distributor ID:	Ubuntu
Description:	Ubuntu 14.04.3 LTS
Release:	14.04
Codename:	trusty
# ^D
```

## Install

Install latest version using Golang (recommended)

```console
$ go get github.com/andock/ssh2docksal/cmd/ssh2docksal
```


## Test with Docker


Here is an example about how to use ssh2docksal inside Docker

```console
$ docker run --privileged -v /var/lib/docker:/var/lib/docker -it --rm -p 2222:2222 andock/ssh2docksal
```

## Initialized fored and massiv inspired by  https://github.com/moul/ssh2docker
