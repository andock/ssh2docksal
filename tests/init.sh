#!/bin/bash

# Test init script for travis.
# Installs docksal and setup two docksal environments.

set -e

# ==================  Setup ssh ==============================
# Setup ssh keys
cp id_rsa.pub ~/.ssh/id_rsa.pub
cp id_rsa ~/.ssh/id_rsa
cp authorized_keys ~/.ssh/authorized_keys
chmod 600 ~/.ssh/id_rsa.pub
chmod 400 ~/.ssh/id_rsa
chmod 600 ~/.ssh/authorized_keys
eval `ssh-agent -s`
ssh-add ~/.ssh/id_rsa
ls -al ~/.ssh


# ================== Install docksal ==============================
curl -fsSL get.docksal.io | bash


# ================== Initialize docksal projects ==============================
# Clone test repositories
git clone https://github.com/docksal/drupal8.git ssh2docksal_source
cp docksal.yml ssh2docksal_source/.docksal/docksal.yml
cp -R ssh2docksal_source ssh2docksal_target
cd ssh2docksal_source
fin init
#fin ssh-add id_rsa

cd ..
cd ssh2docksal_target
fin init
cd ..

# ==================  Setup drush aliases ==============================
cp ssh2docksal.aliases.drushrc.php ssh2docksal_source/drush/



#  ==================  Start ssh2docksal docker image ==============================
./startup.sh
id -u travis
docker ps
docker exec andock-ssh2docksal ls -al /home/docker/.ssh
docker exec andock-ssh2docksal id -u docker
fin version