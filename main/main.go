package main

import (
	"github.com/andock/ssh2docksal"
	"github.com/andock/ssh2docksal/docker_client"
	"github.com/apex/log"
	"github.com/codegangsta/cli"
	"github.com/gliderlabs/ssh"
	"os"
)

// StartServer is the default cli action
func StartServer(c *cli.Context) {
	var authorization ssh.Option

	if c.String("auth-type") == "public-key" {
		authorizedKeyFile := c.String("authorized-key-file")
		log.Info("Used authorized_key_file: " + authorizedKeyFile)
		authorization = ssh2docksal.PublicKeyAuth(authorizedKeyFile)
	}

	if c.String("auth-type") == "noauth" {
		log.Info("Authorization: auth-type")
		authorization = ssh2docksal.NoAuth()
	}
	if authorization == nil {
		log.Warn("No valid authenification type" + c.String("auth-type"))
		return
	}

	level := log.InfoLevel
	if c.Bool("verbose") {
		level = log.DebugLevel
	}
	level = log.DebugLevel
	log.SetLevel(level)

	sshHandler := &docker_client.CliDockerHandler{}

	ssh2docksal.SSHHandler(sshHandler, ssh2docksal.Config{WelcomeMessage: c.String("welcome-message")})
	bindPort := c.String("bind")
	log.Info("Starting ssh server on port " + bindPort)
	log.WithError(ssh.ListenAndServe(bindPort, nil, authorization))
	log.Info("Server started")
}

func main() {
	app := cli.NewApp()
	app.Author = "Christian Wiedemann"
	app.Email = "christian.wiedemann@key-tec.de"
	app.Version = "0.0.10"
	app.Usage = "SSH to docksal"

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose, V",
			Usage: "Enable verbose mode",
		},
		cli.StringFlag{
			Name:  "bind, b",
			Value: ":2222",

			Usage: "Listen to address",
		},
		cli.StringFlag{
			Name:  "auth-type",
			Value: "public-key",
			Usage: "Authentification type: [public-key|noauth]",
		},
		cli.StringFlag{
			Name:  "authorized-key-file",
			Value: os.Getenv("HOME") + "/.ssh/authorized_keys",
			Usage: "Path to your authorized key file.",
		},
		cli.StringFlag{
			Name:  "welcome-message",
			Value: "docksal",
			Usage: "Welcome message",
		},
	}
	log.Infof("Welcome to ssh2docksal %s", app.Version)
	app.Action = StartServer
	app.Run(os.Args)
}
