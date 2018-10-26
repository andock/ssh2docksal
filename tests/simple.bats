#!/usr/bin/env bats

@test "remote ls" {
  run ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no ssh2docksal_target@192.168.64.100 -p 2222 ls
  [ $status = 0 ]
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