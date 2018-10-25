#!/usr/bin/env bats

@test "connect" {
  run ssh ssh2docksal_target@192.168.64.100 -p 2222 ls
  [ $status = 0 ]
}

@test "server:install" {
  ../../bin/andock.sh @${ANDOCK_CONNECTION} server install "andock" "${ANDOCK_ROOT_USER}"
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