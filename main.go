package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
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
func main() {
	ssh.Handle(func(s ssh.Session) {
		//io.WriteString(s, "authorizedKey")
		//authorizedKey := gossh.MarshalAuthorizedKey(s.PublicKey())
		log.Infof("Starting")

		var cmd *exec.Cmd
		var err error
		var entrypoint = "/bin/bash"
		var command []string
		var isPty bool
		ptyReq, winCh, isPty := s.Pty()
		command = s.Command()
		// checking if a container already exists for this user
		existingContainer := ""

		cmd = exec.Command("docker", "ps", fmt.Sprintf("--filter=name=%s_cli_1", s.User()), "--quiet", "--no-trunc")

		buf, err := cmd.CombinedOutput()
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

			if len(command) != 0 {
				args = append(args, "-c")
				args = append(args, strings.Join(command, " "))
			}

			log.Infof("Executing 'docker %s'", strings.Join(args, " "))
			cmd = exec.Command("docker", args...)
			//cmd.Stdout = s
			//cmd.Stdin = s
			//cmd.Stderr = s
			cmd.SysProcAttr = &syscall.SysProcAttr{
				Setctty: isPty,
				Setsid:  true,
			}

			if isPty {
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
			} else {
				log.Infof("No tty")

				log.Infof("Copy pipe")

				log.Infof("Start")
				err = cmd.Start()

				if err != nil {
					log.Warnf("cmd.Start failed: %v", err)
					return
				}
				log.Infof("Wait")

				log.Infof("Wait done")
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
