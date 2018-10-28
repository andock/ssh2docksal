#!/usr/bin/env bats



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
  run fin drush sql-sync -y @ssh2docksal.target @self
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