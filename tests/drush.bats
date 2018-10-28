#!/usr/bin/env bats



@test "drush sa" {
  cd ssh2docksal_source/docroot
  fin drush sa
  fin drush @ssh2docksal.target sa

}

@test "drush sql-sync" {
  cd ssh2docksal_source/docroot
  fin drush sql-drop -y
  fin drush sql-sync @ssh2docksal.target @self -y

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