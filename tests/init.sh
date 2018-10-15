#!/bin/bash
git clone git@github.com:docksal/drupal8.git ssh2docksal_source
cp -R ssh2docksal_source ssh2docksal_target

cd ssh2docksal_source 
fin init
cd ..
cd ssh2docksal_target
fin init 

cp ssh2docksal.aliases.drushrc.php ssh2docksal_source/drush/