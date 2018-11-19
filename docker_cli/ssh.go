package docker_cli

import (
	"fmt"
	"github.com/andock/ssh2docksal"
	"github.com/apex/log"
	"github.com/gliderlabs/ssh"
	"github.com/kr/pty"
	"github.com/mbndr/figlet4go"
	"github.com/pkg/sftp"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

func setWinsize(f *os.File, w, h int) {
	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(h), uint16(w), 0, 0})))
}

// CliDockerClient is
type CliDockerHandler struct {
}

func (a *CliDockerHandler) SftpHandler(containerID string) sftp.Handlers {
	return DockerCliSftpHandler(containerID)
}

func (a *CliDockerHandler) Find(containerName string) (string, error) {

	findExecResult := exec.Command("docker", "ps", fmt.Sprintf("--filter=name=%s$", containerName), "--quiet", "--no-trunc")
	buf, err := findExecResult.CombinedOutput()
	if err != nil {
		log.Errorf("docker ps ... failed: %v", err)
		return "", err
	}
	existingContainer := strings.TrimSpace(string(buf))
	if existingContainer == "" {
		log.Errorf("Container: %s not found", containerName)
		return "", fmt.Errorf("container %s not found", containerName)
	}
	return existingContainer, nil
}

func (a *CliDockerHandler) Execute(containerID string, s ssh.Session, c ssh2docksal.Config) {
	var entrypoint = "/bin/bash"
	var command = s.Command()
	var joinedArgs string

	var isPty bool
	_, _, isPty = s.Pty()

	// Attaching to an existing container
	args := []string{"exec"}
	args = append(args, "-u")
	args = append(args, "docker")
	args = append(args, "-i")

	if isPty {
		args = append(args, "-t")
	}

	args = append(args, containerID)
	if entrypoint != "" {
		args = append(args, entrypoint)
	}
	isScp := false
	if len(command) != 0 {
		args = append(args, "-lc")
		args = append(args, strings.Join(command, " "))
		isScp = (command[0] == "rsync" || command[0] == "scp")
	}
	joinedArgs = strings.Join(args, " ")
	log.Debugf("Executing 'docker %s'", joinedArgs)
	if isPty {
		executePtyCommand("docker", args, s, c)
	} else {
		executeCommand("docker", args, s, isScp)
	}
	log.Debugf("Executing done")
}

func executeCommand(name string, args []string, s ssh.Session, isScp bool) {
	cmd := exec.Command(name, args...)
	log.Debugf("executeCommand started. Mode:  %s\n")

	if !isScp {
		cmd.Stdout = s
		cmd.Stderr = s
	}
	if isScp {
		stdin, err := cmd.StdinPipe()
		if err != nil {
			log.Warnf("Could not open stdin pipe of command: %s\n", err)
			s.Exit(255)
			return
		}

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Warnf("Could not open stdout pipe of command: %s\n", err)
			s.Exit(255)
			return
		}

		close := func() {
			s.Exit(0)
		}
		var once sync.Once
		go func() {
			io.Copy(stdin, s)
			once.Do(close)
		}()
		go func() {
			io.Copy(s, stdout)
			once.Do(close)
		}()
	}
	cmd.Run()
	log.Debugf("executeCommand completed")
}

func executePtyCommand(name string, args []string, s ssh.Session, c ssh2docksal.Config) {
	ptyReq, winCh, _ := s.Pty()
	cmd := exec.Command(name, args...)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setctty: true,
		Setsid:  true,
	}
	cmd.Stdout = s
	cmd.Stdin = s
	cmd.Stderr = s

	ascii := figlet4go.NewAsciiRender()
	renderStr, _ := ascii.Render(c.WelcomeMessage)
	fmt.Fprintf(s, "%s\n\r", renderStr)

	cmd.Env = append(cmd.Env, fmt.Sprintf("TERM=%s", ptyReq.Term))
	f, err := pty.Start(cmd)
	if err != nil {
		panic(err)
	}
	go func() {

		for win := range winCh {
			setWinsize(f, win.Width, win.Height)
		}
	}()
	go func() {
		io.Copy(f, s) // stdin
	}()
	io.Copy(s, f) // stdout
	cmd.Wait()
	s.Exit(0)
}
