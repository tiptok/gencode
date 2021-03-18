package cmd

import (
	"github.com/tiptok/gencode/cmd/dddgen"
	"github.com/urfave/cli"
	"os"
)

type Cmd interface {
	App() *cli.App
	Init(opts ...Option) error
	// Options set within this command
	Options() Options
}
type Option func(o *Options)
type Options struct {
	// For the Command Line itself
	Name        string
	Description string
	Version     string
}

func Name(s string) Option {
	return func(o *Options) {
		o.Name = s
	}
}
func Description(s string) Option {
	return func(o *Options) {
		o.Description = s
	}
}
func Version(s string) Option {
	return func(o *Options) {
		o.Version = s
	}
}

type cmd struct {
	opts Options
	app  *cli.App
}

func (c *cmd) App() *cli.App {
	return cli.NewApp()
}
func (c *cmd) Init(opts ...Option) error {
	for _, o := range opts {
		o(&c.opts)
	}
	c.app.Name = c.opts.Name
	c.app.Version = c.opts.Version
	c.app.HideVersion = len(c.opts.Version) == 0
	c.app.Usage = c.opts.Description

	c.app.Commands = append(c.app.Commands, Commands()...)
	return nil
}
func (c *cmd) Options() Options {
	return c.opts
}

var DefaultCmd *cmd

func newCmd() *cmd {
	return &cmd{}
}
func Init(opts ...Option) {
	DefaultCmd = newCmd()
	DefaultCmd.app = cli.NewApp()
	DefaultCmd.Init(opts...)
	err := DefaultCmd.app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}

func Commands() []cli.Command {
	var commands []cli.Command
	//commands = append(commands, mvcgen.Commands()...)
	commands = append(commands, dddgen.Commands()...)
	return commands
}
