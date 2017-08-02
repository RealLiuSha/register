package main

import (
	"fmt"
	"os"
	"time"

	"git.wolaidai.com/DevOps/register/pkg"
	"git.wolaidai.com/DevOps/register/pkg/g"
	"git.wolaidai.com/DevOps/register/pkg/utils"
	"git.wolaidai.com/DevOps/register/pkg/utils/log"
	"github.com/urfave/cli"
)

func main() {
	app := &cli.App{
		Name:     "consul-register",
		Usage:    "Docker event automatically listens and synchronizes consule",
		Version:  g.VERSION,
		Compiled: time.Now(),
		Authors: []cli.Author{
			{
				Name:  "freedie.liu",
				Email: "freedie.liu@wolaidai.com",
			},
		},
		Before: func(c *cli.Context) error {
			fmt.Fprintf(c.App.Writer, utils.StripIndent(
				`
			#####  ######  ####  #  ####  ##### ###### #####
			#    # #      #    # # #        #   #      #    #
			#    # #####  #      #  ####    #   #####  #    #
			#####  #      #  ### #      #   #   #      #####
			#   #  #      #    # # #    #   #   #      #   #
			#    # ######  ####  #  ####    #   ###### #    #

			`))
			return nil
		},
		Commands: []cli.Command{
			{
				Name:  "start",
				Usage: "start a new consul-register",
				Action: func(ctx *cli.Context) error {
					log.LogLevel(ctx.String("log.level"))
					pkg.Start()
					return nil
				},
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:        "concurrency",
						Value:       10,
						Usage:       "concurrency number",
						Destination: &g.CONCURRENCY,
					},
					&cli.StringFlag{
						Name:        "docker.endpoint",
						Value:       "unix:///var/run/docker.sock",
						Usage:       "Docker Conn EndPoint",
						Destination: &g.DOCKER_ENDPOINT,
					},
					&cli.StringFlag{
						Name:        "consul.host",
						Value:       "",
						Usage:       "consul server host",
						Destination: &g.CONSUL_HOST,
					},
					&cli.StringFlag{
						Name:        "consul.port",
						Value:       "8500",
						Usage:       "consul server port",
						Destination: &g.CONSUL_PORT,
					},
					&cli.StringFlag{
						Name:        "marathon.host",
						Value:       "192.168.20.2",
						Usage:       "marathon server host",
						Destination: &g.MARATHON_HOST,
					},
					&cli.StringFlag{
						Name:        "marathon.port",
						Value:       "8080",
						Usage:       "marathon server port",
						Destination: &g.MARATHON_PORT,
					},
					&cli.StringFlag{
						Name:        "marathon.dyups",
						Value:       "/base-service/loadbalance/orange",
						Usage:       "marathon of orange api url",
						Destination: &g.MARATHON_DYUPS_URL,
					},
					&cli.StringFlag{
						Name:        "dyups.port",
						Value:       "18081",
						Usage:       "orange server port",
						Destination: &g.DYUPS_PORT,
					},
					&cli.StringFlag{
						Name:        "dyups.url",
						Value:       "/upstream-admin",
						Usage:       "orange upstream url",
						Destination: &g.DYUPS_URL,
					},
					&cli.Int64Flag{
						Name:        "check.second",
						Value:       10,
						Usage:       "timer check upstream duration",
						Destination: &g.UPSTREAM_CHECK_SECONDS,
					},
					&cli.BoolFlag{
						Name:        "check.enable",
						Usage:       "timer check upstream is enable",
						Destination: &g.UPSTREAM_CHECK_ENABLE,
					},
					&cli.StringFlag{
						Name:        "log.level",
						Value:       "info",
						Usage:       "Only log messages with the given severity or above. Valid levels: [debug, info, warn, error, fatal]",
						Destination: &g.LOG_LEVEL,
					},
				},
			},
		},
	}

	app.Run(os.Args)
}
