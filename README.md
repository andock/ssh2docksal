[![Latest Release](https://img.shields.io/github/release/andock/ssh2docksal.svg?style=flat-square)](https://github.com/andock/ssh2docksal/releases/latest) [![Build Status](https://img.shields.io/travis/andock/ssh2docksal.svg?style=flat-square)](https://travis-ci.org/andock/ssh2docksal)

![alt text](images/logo_circle.svg "andock")
# ssh2docksal
Andock ssh2docksal is a simple ssh server which connects you directly to your docksal container via ssh.

## Samples:

Connect to the `cli` container of `projectname`
```
    ssh projectname@localhost -p 2222
```

Connect to the `db` container of `projectname`
```
    ssh projectname--db@localhost -p 2222
```

## Usage without authorization:
```
docker run \
-d \
-v /var/run/docker.sock:/var/run/docker.sock \
-v /usr/bin/docker:/usr/bin/docker \
-p 192.168.64.100:2222:2222 andockio/ssh2docksal --auth-type=noauth
```

## Usage with authorization:
```
docker run \
-d \
-v /var/run/docker.sock:/var/run/docker.sock \
-v /usr/bin/docker:/usr/bin/docker \
-v ${HOME}/.ssh/authorized_keys:/root/.ssh/authorized_keys \
-p 192.168.64.100:2222:2222 andockio/ssh2docksal
```
