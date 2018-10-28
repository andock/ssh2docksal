#!/bin/bash

# Test init script for travis.
# Installs docksal and setup two docksal environments.

set -e

# ================== Install docksal ==============================
curl -fsSL get.docksal.io | bash

# ==================  Setup drush aliases ==============================
cp ssh2docksal.aliases.drushrc.php ssh2docksal_source/drush/

# ==================  Setup ssh ==============================
# Setup ssh keys
cp id_rsa.pub ~/.ssh/id_rsa.pub
cp id_rsa ~/.ssh/id_rsa
cp authorized_keys ~/.ssh/authorized_keys
chmod 600 ~/.ssh/id_rsa.pub
chmod 600 ~/.ssh/id_rsa
chmod 600 ~/.ssh/authorized_keys
eval `ssh-agent -s`
ssh-add ~/.ssh/id_rsa

#  ==================  Start ssh2docksal docker image ==============================
./startup.sh

