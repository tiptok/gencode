package mvcgen

import (
	"github.com/urfave/cli"
)

func Commands() []cli.Command {
	return []cli.Command{
		{
			Name:  "new",
			Usage: "Create a service template; example: gencode new -c Auth -m Login",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "c",
					Usage: "controller name for the service",
					Value: "Auth",
				},
				cli.StringFlag{
					Name:  "m",
					Usage: "controller handler name",
					Value: "Login",
				},
			},
			Action: Run,
		},
	}
}
