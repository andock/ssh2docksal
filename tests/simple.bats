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
  expect simple_tty.expect
  run ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no ssh2docksal_target@192.168.64.100 -p 2222 ls simple_tty.txt
  [[ "$output" =~ "simple_tty.txt" ]]
}

@test "Test scp download" {
  sleep 2
  run scp -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no  -P 2222 ssh2docksal_target@192.168.64.100:/var/www/README.md download_README.md
  run ls download_README.md
  [[ "$output" =~ "download_README.md" ]]
}

@test "Test scp upload" {
  run scp -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no  -P 2222  tty-upload.txt ssh2docksal_target@192.168.64.100:tty-upload.txt
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