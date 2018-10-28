#!/usr/bin/env bats

setup() {

    if [ ! -d "ssh2docksal_source" ]; then
        git clone https://github.com/docksal/drupal8.git ssh2docksal_source
        cd ssh2docksal_source
        fin init
        cd ..
    fi

    if [ ! -d "ssh2docksal_target" ]; then
        cp -R ssh2docksal_source ssh2docksal_target
         cd ssh2docksal_target
         fin init
         cd ..
     fi
}

@test "drush sa" {
  cd ssh2docksal_source/docroot
  run fin drush sa
  [ $status = 0 ]

  run fin drush @ssh2docksal.target sa
  [ $status = 0 ]
}

@test "drush sql-sync" {
  cd ssh2docksal_source/docroot
  run fin drush sql-drop -y
  [ $status = 0 ]
  run fin drush sql-sync @ssh2docksal.target @self
  [ $status = 0 ]
  run 'curl -sL -I  http://ssh2docksal-source.docksal | grep "HTTP/1.1 200 OK"'
  [[ "$output" =~ "HTTP/1.1 200 OK" ]]
}

teardown() {
    echo "Status: $status"
    echo "Output:"
    echo "================================================================"
    for line in "${lines[@]}"; do
        echo $line
    done
    echo "================================================================"
}