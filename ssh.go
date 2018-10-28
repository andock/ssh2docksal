package ssh2docksal

import (
	"fmt"
	"github.com/apex/log"
	"github.com/gliderlabs/ssh"
	"github.com/kr/pty"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)
var findExecCommand = exec.Command

func setWinsize(f *os.File, w, h int) {
	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(h), uint16(w), 0, 0})))
}

func executePtyCommand(name string, args []string, s ssh.Session) {
	ptyReq, winCh, _ := s.Pty()
	cmd := exec.Command(name, args...)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setctty: true,
		Setsid:  true,
	}
	cmd.Stdout = s
	cmd.Stdin = s
	cmd.Stderr = s

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
func executeCommand(name string, args []string, s ssh.Session, isScp bool) {
	cmd := exec.Command(name, args...)

	if (!isScp) {
		cmd.Stdout = s
		cmd.Stderr = s
	}
	if (isScp) {
		stdin, err := cmd.StdinPipe()
		if err != nil {
			log.Infof("Could not open stdin pipe of command: %s\n", err)
		}

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Infof("Could not open stdout pipe of command: %s\n", err)
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
	log.Infof("Run done")
}

func getContainerID(username string) (string, error) {

	var err error
	existingContainer := ""
	var container string
	s := strings.Split(username,"--")
	projectName := s[0]

	if len(s) == 2 {
		container = s[1]
	} else if (len(s) == 1){
		container = "cli"
	}

	containerName := projectName+"_"+container +"_1"
	findExecResult := findExecCommand("docker", "ps", fmt.Sprintf("--filter=name=%s", containerName), "--quiet", "--no-trunc")

	buf, err := findExecResult.CombinedOutput()
	if err != nil {
		log.Warnf("docker ps ... failed: %v", err)
		return "", err
	}
	existingContainer = strings.TrimSpace(string(buf))

	if existingContainer == "" {
		log.Warnf("Container: %s not found", containerName)
		return "", fmt.Errorf("container %s not found", containerName)
	}
	return existingContainer, nil
}

// SSHHandler handles the ssh connection
func SSHHandler() {
	ssh.Handle(func(s ssh.Session) {

		log.Infof("Starting")

		var entrypoint = "/bin/bash"
		var command []string
		var joinedArgs string

		command = s.Command()
		existingContainer, err := getContainerID(s.User())

		if (err != nil) {
			s.Exit(1)
			return
		}

		// Opening Docker process
		if existingContainer != "" {
			var isPty bool
			_, _, isPty = s.Pty()

			log.Infof("Found container %s", existingContainer)
			// Attaching to an existing container
			args := []string{"exec"}
			args = append(args, "-u")
			args = append(args, "docker")
			args = append(args, "-i")

			if isPty {
				args = append(args, "-t")
			}

			args = append(args, existingContainer)
			if entrypoint != "" {
				args = append(args, entrypoint)
			}
			isScp := false
			if len(command) != 0 {
				args = append(args, "-lc")
				args = append(args, strings.Join(command, " "))
				isScp = (command[0] == "rsync" || command[0] == "scp");
			}
			joinedArgs = strings.Join(args, " ")
			log.Infof("Executing 'docker %s'", joinedArgs)
			if isPty {
				executePtyCommand("docker", args, s)
			} else {
				executeCommand("docker", args, s, isScp)
			}
			log.Infof("Executing done")
		}
	})
}
