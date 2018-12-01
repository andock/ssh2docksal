[![Latest Release](https://img.shields.io/github/release/andock/ssh2docksal.svg?style=flat-square)](https://github.com/andock/ssh2docksal/releases/latest) [![Build Status](https://img.shields.io/travis/andock/ssh2docksal.svg?style=flat-square)](https://travis-ci.org/andock/ssh2docksal)

![alt text](images/logo_circle.svg "andock")
# ssh2docksal
ssh2docksal is a ssh server which connects you directly to your docksal container via ssh. 

### Currently implemented: 
* Supports TTY
* Supports drush / rsync
* Supports sftp
### TODO: 
* Symlinks

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
-d \
-e "HOST_UID=$(id -u)" \
-e "HOST_GID=$(cut -d: -f3 < <(getent group docker))" \
-v /var/run/docker.sock:/var/run/docker.sock \
-v /usr/bin/docker:/usr/bin/docker \
--name andock-ssh2docksal \
-v ${HOME}/.ssh/authorized_keys:/home/docker/.ssh/authorized_keys \
-p 192.168.64.100:2222:2222 andockio/ssh2docksal
```

## Usage with authorization:
E.g. for your sandbox server.
```
docker run \
-d \
-e "HOST_UID=$(id -u)" \
-e "HOST_GID=$(cut -d: -f3 < <(getent group docker))" \
-v /var/run/docker.sock:/var/run/docker.sock \
-v /usr/bin/docker:/usr/bin/docker \
--name andock-ssh2docksal \
-v ${HOME}/.ssh/authorized_keys:/home/docker/.ssh/authorized_keys \
-p 0.0.0.0:2222:2222 andockio/ssh2docksal 
```
