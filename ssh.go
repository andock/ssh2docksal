package ssh2docksal

import (
	"fmt"
	"github.com/apex/log"
	"github.com/common-nighthawk/go-figure"
	"github.com/gliderlabs/ssh"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/sftp"
	"strings"
	"time"
)

// DockerClientInterface for different docker clients
type dockerClientInterface interface {
	Execute(containerID string, s ssh.Session, c Config)
	Find(containerName string) (string, error)
	SftpHandler(containerID string, config Config) sftp.Handlers
}

// Config for ssh options
type Config struct {
	WelcomeMessage string
	DockerUser     string
	Cache          *cache.Cache
}

func (config *Config) getCache() *cache.Cache {
	if config.Cache != nil {
		return config.Cache
	}
	cache := cache.New(5*time.Minute, 10*time.Minute)
	config.Cache = cache
	return config.Cache
}
func getContainerNames(username string) (string, string) {

	var container string
	s := strings.Split(username, "---")
	projectName := s[0]

	if len(s) == 2 {
		container = s[1]
	} else if len(s) == 1 {
		container = "cli"
	}
	return projectName, container
}
func getContainerID(client dockerClientInterface, username string) (string, error) {
	projectName, container := getContainerNames(username)
	containerName := projectName + "_" + container + "_1"
	return client.Find(containerName)
}

// SSHHandler handles the ssh connection
func SSHHandler(sshHandler dockerClientInterface, config Config) {
	ssh.Handle(func(s ssh.Session) {
		log.Debugf("Looking for  container %s", s.User())
		c := config.getCache()
		var err error
		var existingContainer string

		cacheValue, found := c.Get(s.User())
		if !found {
			existingContainer, err = getContainerID(sshHandler, s.User())
			if existingContainer == "" {
				log.Errorf("Container %s lookup failed. Maybe the container is not up. Run fin up", s.User(), err)
			} else {
				c.Set(s.User(), existingContainer, cache.DefaultExpiration)
			}
			if err != nil {
				s.Exit(255)
				return
			}
		} else {
			existingContainer = cacheValue.(string)
		}

		if existingContainer == "" {
			log.Errorf("No container found for name %s", s.User())
			s.Exit(1)
			return
		}

		log.Debugf("Found container %s", existingContainer)

		if err != nil {
			log.Errorf(err.Error())
			s.Exit(1)
			return
		}
		projectName, container := getContainerNames(s.User())
		config.DockerUser = "root"
		if container == "cli" {
			config.DockerUser = "docker"
		}
		if s.Subsystem() == "sftp" {
			log.Debugf("Start sftp")
			sftpServer := sftp.NewRequestServer(s, sshHandler.SftpHandler(existingContainer, config))
			_ = sftpServer.Serve()

		} else {
			_, _, isPty := s.Pty()
			if config.WelcomeMessage != "" && isPty == true && len(s.Command()) == 0 {

				message := figure.NewFigure(config.WelcomeMessage, "", true).String()
				fmt.Fprintf(s, "\n\n%s\n\n\r", message)
				fmt.Fprintf(s, " Welcome to %s.\n\n\r", config.WelcomeMessage)
				fmt.Fprintf(s, " This is the %s service\n\r", container)
				fmt.Fprintf(s, " of environment %s.\n\n\r", projectName)
			}

			sshHandler.Execute(existingContainer, s, config)
		}

	})
}
