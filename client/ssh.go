package client

import (
	"fmt"
	"github.com/andock/ssh2docksal"
	"github.com/apex/log"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/gliderlabs/ssh"
	"github.com/pkg/sftp"
	"golang.org/x/net/context"
	"io"
	"strings"
	"time"
)

type DockerClient struct {
}

// SftpHandler returns the associated sftp docker handler.
func (a *DockerClient) SftpHandler(containerID string) sftp.Handlers {
	return DockerCliSftpHandler(containerID)
}

// Find lookups for container id  by given container name
func (a *DockerClient) Find(containerName string) (string, error) {

	cli, err := client.NewEnvClient()
	if err != nil {
		return "", err
	}
	args, err := filters.ParseFlag("name=" +containerName, filters.NewArgs())
	if err != nil {
		return "", err
	}
	options := types.ContainerListOptions{Filters: args}
	containers, err := cli.ContainerList(context.Background(), options)
	if err != nil {
		return "", err
	}
	if len(containers) == 1 {
		container := containers[0]
		if container.State != "running" {
			err = fmt.Errorf("Container %s is not running. Run fin up.", containerName)
			log.Errorf(err.Error())
			return "", err
		}
		return container.ID, nil
	} else if len(containers) > 1 {
		err = fmt.Errorf("Found more than one container. Name: %s.", containerName)
		log.Errorf(err.Error())
		return "", err
	} else {
		err = fmt.Errorf("Unable to access container %s. Propably the container is not up. Run fin up.", containerName)
		log.Errorf(err.Error())
		return "", err
	}

}

func dockerExec(containerID string, command string, cfg container.Config, sess ssh.Session) (status int, err error) {
	log.Debugf("SSH: Execute command: %s", command)
	status = 255
	ctx := context.Background()
	docker, err := client.NewEnvClient()
	if err != nil {
		log.Errorf("Couldn't connect to docker")
		return status, err
	}

	execStartCheck := types.ExecConfig{
		Tty: cfg.Tty,
	}

	ec := types.ExecConfig{
		AttachStdout: cfg.AttachStdout,
		AttachStdin:  cfg.AttachStdin,
		AttachStderr: cfg.AttachStderr,
		Detach:       false,
		Tty:          cfg.Tty,
	}
	ec.Cmd = append(ec.Cmd, "/bin/bash")

	if command != "" {
		ec.Cmd = append(ec.Cmd, "-lc")
		ec.Cmd = append(ec.Cmd, command)
	}

	ec.User = "docker"
	eresp, err := docker.ContainerExecCreate(context.Background(), containerID, ec)
	if err != nil {
		log.Errorf("docker.ContainerExecCreate: ", err)
		return
	}

	stream, err := docker.ContainerExecAttach(ctx, eresp.ID, execStartCheck)
	if err != nil {
		log.Errorf("docker.ContainerExecAttach: ", err)
		return
	}
	defer stream.Close()

	outputErr := make(chan error)

	go func() {
		var err error
		if cfg.Tty {
			_, err = io.Copy(sess, stream.Reader)
		} else {
			_, err = stdcopy.StdCopy(sess, sess.Stderr(), stream.Reader)
		}
		outputErr <- err
	}()

	go func() {
		defer stream.CloseWrite()
		io.Copy(stream.Conn, sess)
	}()

	if cfg.Tty {
		_, winCh, _ := sess.Pty()
		go func() {
			for win := range winCh {
				err := docker.ContainerExecResize(ctx, eresp.ID, types.ResizeOptions{
					Height: uint(win.Height),
					Width:  uint(win.Width),
				})
				if err != nil {
					log.WithError(err)
					break
				}
			}
		}()
	}
	for {
		inspect, err := docker.ContainerExecInspect(ctx, eresp.ID)
		if err != nil {
			log.WithError(err)
		}
		if !inspect.Running {
			status = inspect.ExitCode
			break
		}
		time.Sleep(time.Second)
	}
	return
}

// Execute executes commands
func (a *DockerClient) Execute(containerID string, s ssh.Session, c ssh2docksal.Config) {
	_, _, isPty := s.Pty()
	cfg := container.Config{AttachStdin: true, AttachStderr: true, AttachStdout: true, Tty: isPty}
	_, err := dockerExec(containerID, strings.Join(s.Command(), " "), cfg, s)
	if err != nil {
		s.Exit(255)
	}

}
