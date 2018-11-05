#!/usr/bin/expect -f

set timeout 10
cd ssh2docksal_source/docroot

spawn fin drush @ssh2docksal.target ssh
expect "docker@cli:/var/www" {send "touch 'drush.txt'\r"}


