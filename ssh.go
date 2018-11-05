package ssh2docksal

import (
	"github.com/apex/log"
	"github.com/gliderlabs/ssh"
	"strings"
)

// DockerClientInterface for differnt docker clients
type dockerClientInterface interface {
	Execute(containerName string, s ssh.Session, c Config)
	Find(containerName string) (string, error)
}

// Config for ssh options
type Config struct {
	Banner string
	Debug bool
}

func getContainerID(client dockerClientInterface, username string) (string, error) {

	var container string
	s := strings.Split(username,"---")
	projectName := s[0]

	if len(s) == 2 {
		container = s[1]
	} else if (len(s) == 1) {
		container = "cli"
	}

	containerName := projectName + "_" + container + "_1"
	return client.Find(containerName)
}

// SSHHandler handles the ssh connection
func SSHHandler(client dockerClientInterface, c Config) {
	ssh.Handle(func(s ssh.Session) {
		log.Debugf("Start connection")

		existingContainer, err := getContainerID(client, s.User())

		if (err != nil) {
			s.Exit(1)
			return
		}

		// Opening Docker process
		if existingContainer != "" {
			client.Execute(existingContainer, s, c)
		}
	})
}
