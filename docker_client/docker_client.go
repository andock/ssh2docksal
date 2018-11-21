package docker_client

import (
	"bytes"
	"fmt"
	"github.com/apex/log"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"strings"
	"golang.org/x/net/context"
)

func simpleExec(containerID string, command string ) ( error) {
	_, err := outpuExec(containerID, command)
	return err
}
func outpuExec(containerID string, command string) (string, error) {

	cli, err := client.NewEnvClient()
	args := []string{"bash", "-c", command}
	if err != nil {
		return "",err
	}
	execConfig:= types.ExecConfig{Tty:false, AttachStdout:true, AttachStderr:true, Cmd: args}
	respIdExecCreate,err := cli.ContainerExecCreate(context.Background(),containerID,execConfig)
	if err != nil {
		return  "", err
	}
	conection, err := cli.ContainerExecAttach(context.Background(),respIdExecCreate.ID,types.ExecConfig{})
	if err != nil {
		log.Errorf("Unable to execute %s", command)
		log.WithError(err)
	}
	conection.CloseWrite()
	stdoutput := new(bytes.Buffer)
	stderror := new(bytes.Buffer)
	stdcopy.StdCopy(stdoutput, stderror, conection.Reader)
	//ob, err := ioutil.ReadAll(conection.Reader)
	output := strings.TrimSpace(stdoutput.String())
	errorString := stderror.String()
	if (errorString != "") {
		err :=fmt.Errorf(errorString)
		log.Errorf("Unable to execute %s", command)
		log.WithError(err)
		return "", err
	}
	if err != nil {
		return "", err
	}

	return output, nil
}
