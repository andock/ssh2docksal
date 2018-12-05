package ssh2docksal

import (
	"flag"
	"github.com/apex/log"
	"github.com/gliderlabs/ssh"
	"github.com/pkg/sftp"
	"testing"
)

var testIntegration = flag.Bool("integration", false, "perform integration tests against sftp server process")

func (a *testClient) Find(containerName string) (string, error) {
	return containerName, nil

}

func (a *testClient) Execute(containerID string, s ssh.Session, c Config) {

}

type testClient struct {
}

func (a *testClient) SftpHandler(containerID string, config Config) sftp.Handlers {
	var handler sftp.Handlers
	return handler
}

func TestGetContainerID(t *testing.T) {

	tests := []struct {
		name              string
		containerID       string
		shouldReturnError bool
	}{
		{name: "project", containerID: "project_cli_1", shouldReturnError: false},
		{name: "project---cli", containerID: "project_cli_1", shouldReturnError: false},
		{name: "project---db", containerID: "project_db_1", shouldReturnError: false},
	}

	for _, test := range tests {
		client := &testClient{}
		id, err := getContainerID(client, test.name)
		log.Infof("Container id: %s\n", id)

		if err != nil {
			t.Errorf("Execution: %err.", err)
		}

		if id != test.containerID {
			t.Errorf("Invalid id: %s", id)
		}
	}
}
