package ssh2docksal

import (
	"fmt"
	"github.com/apex/log"
	"os"
	"os/exec"
	"strings"
	"testing"
)

var execMode = ""

func fakeExecCommand(command string, args...string) *exec.Cmd {
	cmd := exec.Command(os.Args[0], command)
	cmd.Env = []string{"GO_TEST_MODE=" + execMode, "GO_TEST_ARG_0="+args[0], "GO_TEST_ARG_1="+args[1]}
	return cmd
}

func TestMain(m *testing.M) {
	switch os.Getenv("GO_TEST_MODE") {
	case "":
		// Normal test mode
		os.Exit(m.Run())
	case "getContainerId":
		fmt.Println(strings.Replace(os.Getenv("GO_TEST_ARG_1") ,"--filter=name=","",1))
	}
}

func TestGetContainerId(t *testing.T) {
	findExecCommand = fakeExecCommand
	execMode = "getContainerId"

	defer func(){ findExecCommand = exec.Command }()

	tests := []struct{
		name      string
		containerID      string
		shouldReturnError      bool

	}{
		{ name: "project", containerID: "project_cli_1", shouldReturnError: false},
		{ name: "project---cli", containerID: "project_cli_1", shouldReturnError: false},
		{ name: "project---db", containerID: "project_db_1", shouldReturnError: false},
	}

	for _, test := range tests {
		id, err := getContainerID(test.name)
		log.Infof("Container id: %s\n", id )

		if err != nil {
			t.Errorf("Execution: %err.", err)
		}

		if (id != test.containerID) {
			t.Errorf("Invalid id: %s", id)
		}
	}

}