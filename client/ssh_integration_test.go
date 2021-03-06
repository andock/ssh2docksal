package client

import (
	"testing"
)

func TestDockerClient_Find(t *testing.T) {
	if !*testIntegration {
		t.Skip("skipping integration test")
	}
	handler := DockerClient{}
	containerID, err := handler.Find("ssh2docksal_source_cli_1")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if containerID == "" {
		t.Errorf("unexpected empty container id")
	}
}
