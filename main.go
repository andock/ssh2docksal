package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"
	"syscall"
	"unsafe"

	"os/exec"
	"strings"

	"github.com/apex/log"

	"github.com/gliderlabs/ssh"
	"github.com/kr/pty"
)

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
		cmd.Stdin = s
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
	if (!isScp) {
		s.Exit(0)
	}
	log.Infof("Run done")
}
func main() {
	ssh.Handle(func(s ssh.Session) {
		//io.WriteString(s, "authorizedKey")
		//authorizedKey := gossh.MarshalAuthorizedKey(s.PublicKey())
		log.Infof("Starting")

		var find *exec.Cmd
		var err error
		var entrypoint = "/bin/bash"
		var command []string
		var joinedArgs string

		command = s.Command()
		// checking if a container already exists for this user
		existingContainer := ""

		find = exec.Command("docker", "ps", fmt.Sprintf("--filter=name=%s_cli_1", s.User()), "--quiet", "--no-trunc")

		buf, err := find.CombinedOutput()
		if err != nil {
			log.Warnf("docker ps ... failed: %v", err)
			return
		}
		existingContainer = strings.TrimSpace(string(buf))

		if existingContainer == "" {
			log.Warnf("CLI Container: %s_cli_1 not found", s.User())
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
				isScp = (command[0] == "rsync");
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

	publicKeyOption := ssh.PublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {

		authorizedKeysBytes, err := ioutil.ReadFile("/home/cw/.ssh/authorized_keys")
		if err != nil {
			log.Fatalf("Failed to load authorized_keys, err: %v", err)
			return false
		}
		for len(authorizedKeysBytes) > 0 {
			pubKey, _, _, rest, err := ssh.ParseAuthorizedKey(authorizedKeysBytes)
			if err != nil {
				log.WithError(err)
			}
			if ssh.KeysEqual(key, pubKey) {
				return true
			}
			authorizedKeysBytes = rest
		}
		return false

	})

	log.Info("starting ssh server on port 2222...")
	log.WithError(ssh.ListenAndServe(":2222", nil, publicKeyOption))
}
