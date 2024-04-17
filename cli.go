// Copyright (c) R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package machineroom

import (
	"context"
	"os"

	"github.com/choria-io/fisk"
	"github.com/sirupsen/logrus"
)

var (
	version     = "development"
	defaultName = "machine-room"
	defaultHelp = "Management Agent"
)

type cliInstance struct {
	opts *Options

	log *logrus.Entry
	cli *fisk.Application

	logfile  string
	loglevel string
	debug    bool
	cfgFile  string
	isLeader bool
	force    bool

	ctx    context.Context
	cancel context.CancelFunc
}

func newCli(o Options) (*cliInstance, error) {
	app := &cliInstance{opts: &o}

	err := app.validateOptions()
	if err != nil {
		return nil, err
	}

	app.cli = app.createCli()

	return app, nil
}

// Application expose the command line framework allowing new commands to be added to it at compile time
func (c *cliInstance) Application() *fisk.Application {
	return c.cli
}

// Run parses and executes the command
func (c *cliInstance) Run(ctx context.Context) error {
	c.ctx, c.cancel = context.WithCancel(ctx)

	args := os.Args[1:]
	if c.opts.Args != nil {
		args = c.opts.Args
	}
	c.cli.MustParseWithUsage(args)

	return nil
}

func (c *cliInstance) createCli() *fisk.Application {
	cli := fisk.New(c.opts.Name, c.opts.Help)
	cli.Author(c.opts.Contact)
	cli.Version(c.opts.Version)
	cli.HelpFlag.Short('h')

	cli.Flag("debug", "Enables debug logging").Default("false").UnNegatableBoolVar(&c.debug)

	run := cli.Commandf("run", "Runs the management agent").Action(c.runCommand)
	run.Flag("config", "Configuration file to use").Required().StringVar(&c.cfgFile)

	reset := cli.Commandf("reset", "Restores the agent to factory defaults").Action(c.resetCommand)
	reset.Flag("config", "Configuration file to use").Required().StringVar(&c.cfgFile)
	reset.Flag("force", "Force reset without prompting").UnNegatableBoolVar(&c.force)

	// generates and saves facts, will be called from auto agents to
	// update facts on a schedule hidden as it's basically a private api
	facts := cli.Commandf("facts", "Save facts about this node to a file").Action(c.factsCommand).Hidden()
	facts.Flag("config", "Configuration file to use").Required().StringVar(&c.cfgFile)

	cli.Commandf("buildinfo", "Shows build information").Action(c.buildInfoCommand).Hidden()

	return cli
}
