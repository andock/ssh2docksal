#!/bin/bash
docker run \
-d \
-v /var/run/docker.sock:/var/run/docker.sock \
-v /usr/bin/docker:/usr/bin/docker \
-v /home/cw/.ssh/authorized_keys:/root/.ssh/authorized_keys \
-p 192.168.64.100:2222:2222 andockio/ssh2docksal

