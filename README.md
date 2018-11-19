[![Latest Release](https://img.shields.io/github/release/andock/ssh2docksal.svg?style=flat-square)](https://github.com/andock/ssh2docksal/releases/latest) [![Build Status](https://img.shields.io/travis/andock/ssh2docksal.svg?style=flat-square)](https://travis-ci.org/andock/ssh2docksal)

![alt text](images/logo_circle.svg "andock")
# ssh2docksal
Andock ssh2docksal is a simple ssh server which connects you directly to your docksal container via ssh.

## Samples:

Connect to the `cli` container of `projectname`
```
    ssh projectname@192.168.64.100 -p 2222
```

Connect to the `db` container of `projectname`
```
    ssh projectname--db@192.168.64.100 -p 2222
```

## Usage without authorization:
E.g. To connect phpStorm via ssh.
```
docker run \
-u docker \
-d \
-v /var/run/docker.sock:/var/run/docker.sock \
-v /usr/bin/docker:/usr/bin/docker \
--name ssh2docksal \
-p 192.168.64.100:2222:2222 andockio/ssh2docksal --auth-type=noauth
```

## Usage with authorization:
E.g. for your sandbox server.
```
docker run \
-u docker \
-d \
-v /var/run/docker.sock:/var/run/docker.sock \
-v /usr/bin/docker:/usr/bin/docker \
-v ${HOME}/.ssh/authorized_keys:/root/.ssh/authorized_keys \
--name ssh2docksal \
-p 192.168.64.100:2222:2222 andockio/ssh2docksal
```
