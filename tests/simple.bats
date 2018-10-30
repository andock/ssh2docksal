#!/usr/bin/env bats

@test "Access denied" {
  run ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no ssh2docksal_target@192.168.64.100 -p 2222 ls
  [ $status != 0 ]
}

@test "Access successful" {
  cat id_rsa.pub >> ~/.ssh/authorized_keys
  run ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no ssh2docksal_target@192.168.64.100 -p 2222 ls
  [ $status = 0 ]
}

@test "Test tty" {
  ./tty.sh
  run ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no ssh2docksal_target@192.168.64.100 -p 2222 ls tty.txt
  [[ "$output" =~ "tty.txt" ]]
}

@test "Test scp" {
  run scp ssh2docksal_target@192.168.64.100:tty.txt tty.txt
  [ $status = 0 ]
  run ls tty.txt
  [[ "$output" =~ "tty.txt" ]]
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