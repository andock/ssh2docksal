package docker_client

import (
	"bytes"
	"fmt"
	"github.com/apex/log"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"golang.org/x/net/context"
	"strings"
)

func simpleExec(containerID string, command string) error {
	_, err := outpuExec(containerID, command)
	return err
}

func outpuExec(containerID string, command string) (string, error) {
	log.Debugf("Execute command: %s", command)
	cli, err := client.NewEnvClient()
	args := []string{"bash", "-c", command}
	if err != nil {
		return "", err
	}
	execConfig := types.ExecConfig{Tty: false, AttachStdout: true, AttachStderr: true, Cmd: args, User: "docker"}
	respIdExecCreate, err := cli.ContainerExecCreate(context.Background(), containerID, execConfig)
	if err != nil {
		return "", err
	}

	connection, err := cli.ContainerExecAttach(context.Background(), respIdExecCreate.ID, types.ExecConfig{})
	if err != nil {
		log.Errorf("Unable to execute %s", command)
		log.WithError(err)
	}
	connection.CloseWrite()
	stdoutput := new(bytes.Buffer)
	stderror := new(bytes                    .Buffer)
	stdcopy.StdCopy(stdoutput, stderror, connection.Reader)
	output := strings.TrimSpace(stdoutput.String())
	errorString := stderror.String()
	if errorString != "" {
		err := fmt.Errorf(errorString)
		log.Errorf("Unable to execute %s", command)
		log.WithError(err)
		return "", err
	}
	return output, nil
}
