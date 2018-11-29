package ssh2docksal

import (
	"github.com/apex/log"
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
	SftpHandler(containerID string) sftp.Handlers
}

// Config for ssh options
type Config struct {
	WelcomeMessage string
	Cache *cache.Cache
}
func (config *Config) getCache() *cache.Cache {
	if config.Cache != nil {
		return config.Cache
	}
	cache := cache.New(5*time.Minute, 10*time.Minute)
	config.Cache = cache
	return config.Cache
}
func getContainerID(client dockerClientInterface, username string) (string, error) {

	var container string
	s := strings.Split(username, "---")
	projectName := s[0]

	if len(s) == 2 {
		container = s[1]
	} else if len(s) == 1 {
		container = "cli"
	}

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
				log.Errorf("container %s lookup failed. Maybe the container is not up. Run fin up", s.User())
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
			s.Exit(1)
			return
		}

		if s.Subsystem() == "sftp" {
			log.Debugf("Start sftp")
			sftpServer := sftp.NewRequestServer(s, sshHandler.SftpHandler(existingContainer))
			_ = sftpServer.Serve()

		} else {
			sshHandler.Execute(existingContainer, s, config)
		}

	})
}
