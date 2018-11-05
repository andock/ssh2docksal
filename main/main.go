package main

import (
	"github.com/andock/ssh2docksal"
	"github.com/andock/ssh2docksal/docker_cli"
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
		authorizedKeyFile := c.String("authorized-key-file")
		log.Info("Used authorized_key_file: " + authorizedKeyFile)
		authorization = ssh2docksal.NoAuth()
	}
	if (authorization == nil) {
		log.Warn("No valid authenification type" + c.String("auth-type"))
		return
	}


	level := log.InfoLevel
	if c.Bool("verbose") {
		level = log.DebugLevel
	}
	log.SetLevel(level)

	adapter := &docker_cli.CliDockerClient{}
	ssh2docksal.SSHHandler(adapter, ssh2docksal.Config{Banner: c.String("banner")})
	bindPort := c.String("bind")
	log.Info("Starting ssh server on port " + bindPort)
	log.WithError(ssh.ListenAndServe(bindPort, nil, authorization))
	log.Info("Server started")
}

func main() {
	app := cli.NewApp()
	app.Author = "Christian Wiedemann"
	app.Email = "christian.wiedemann@key-tec.de"
	app.Version = "1.0"
	app.Usage = "SSH to docksal"

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose, V",
			Usage: "Enable verbose mode",
		},
		cli.StringFlag{
			Name:  "syslog-server",
			Usage: "Configure a syslog server, i.e: udp://localhost:514",
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

	}
	app.Action = StartServer
	app.Run(os.Args)
}
