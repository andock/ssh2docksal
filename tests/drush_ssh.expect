#!/usr/bin/expect -f

set timeout -1

#ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no ssh2docksal_target@192.168.64.100 -p 2222 ls
cd ssh2docksal_source/docroot
spawn fin drush @ssh2docksal_target ssh
expect "docker@cli:/var/www/docroot" {send "touch 'drush.txt'\r"}


