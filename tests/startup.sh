#!/bin/bash
docker run \
-d \
-u docker \
-v /var/run/docker.sock:/var/run/docker.sock \
-v /usr/bin/docker:/usr/bin/docker \
--name ssh2docksal \
-v ${HOME}/.ssh/authorized_keys:/home/docker/.ssh/authorized_keys \
-p 192.168.64.100:2222:2222 andockio/ssh2docksal

