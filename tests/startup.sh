#!/bin/bash
docker rm andock-ssh2docksal -f

docker run \
-d \
-e "HOST_UID=$(id -u)" \
-e "HOST_GID=$(cut -d: -f3 < <(getent group docker))" \
-v /var/run/docker.sock:/var/run/docker.sock \
-v /usr/bin/docker:/usr/bin/docker \
--name andock-ssh2docksal \
--mount type=bind,src=${HOME}/.ssh/authorized_keys,dst=/home/docker/.ssh/authorized_keys \
-p 192.168.64.100:2222:2222 andockio/ssh2docksal --verbose


