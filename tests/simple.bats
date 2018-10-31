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
  expect tty.sh
  run ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no ssh2docksal_target@192.168.64.100 -p 2222 ls tty.txt
  [[ "$output" =~ "tty.txt" ]]
}

@test "Test scp download" {
  scp -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no  -P 2222 ssh2docksal_target@192.168.64.100:tty.txt .
  run ls tty.txt
  [[ "$output" =~ "tty.txt" ]]
}

@test "Test scp upload" {
  run scp -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no  -P 2222  tty.txt ssh2docksal_target@192.168.64.100:tty-upload.txt
  [ $status = 0 ]
  run ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no ssh2docksal_target@192.168.64.100 -p 2222 ls tty-upload.txt
  [[ "$output" =~ "tty-upload.txt" ]]
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