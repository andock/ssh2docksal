package ssh2docksal

import (

	"github.com/apex/log"
	"testing"
	"github.com/gliderlabs/ssh"
)

func (a *testClient) Find(containerName string) (string, error){
	return containerName, nil

}

func (a *testClient) Execute (containerID string, s ssh.Session, c Config) {

}

type testClient struct {

}

func TestGetContainerID(t *testing.T) {

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
		client := &testClient{};
		id, err := getContainerID(client, test.name)
		log.Infof("Container id: %s\n", id )

		if err != nil {
			t.Errorf("Execution: %err.", err)
		}

		if (id != test.containerID) {
			t.Errorf("Invalid id: %s", id)
		}
	}
}