#!/usr/bin/expect -f

set timeout -1

#ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no ssh2docksal_target@192.168.64.100 -p 2222 ls
spawn ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no ssh2docksal_target@192.168.64.100 -p 2222
expect "docker@cli:/var/www" {send "touch 'tty.txt'\r"}


