// Copyright (c) R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package machineroom

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/choria-io/go-choria/build"
	"github.com/choria-io/go-choria/choria"
	"github.com/nats-io/nkeys"
	"github.com/sirupsen/logrus"
)

// CommonConfigure parses the configuration file, prepares logging etc and should be called early in any action
func (c *cliInstance) CommonConfigure() (RuntimeOptions, *logrus.Entry, error) {
	var err error

	c.cfgFile, err = filepath.Abs(c.cfgFile)
	if err != nil {
		return nil, nil, err
	}

	c.opts.StartTime = time.Now().UTC()
	c.opts.ConfigurationDirectory = filepath.Dir(c.cfgFile)
	c.opts.ServerSeedFile = filepath.Join(c.opts.ConfigurationDirectory, defaultServerSeedFileName)
	c.opts.ServerJWTFile = filepath.Join(c.opts.ConfigurationDirectory, defaultServerJwtFileName)
	c.opts.MachinesDirectory = filepath.Join(c.opts.ConfigurationDirectory, defaultMachineStore)
	c.opts.ServerStatusFile = defaultServerStatusFile
	c.opts.ServerSubmissionDirectory = defaultSubmissionSpool
	c.opts.ServerSubmissionSpoolSize = defaultSubmissionSpoolSize
	c.opts.ProvisioningJWTFile = filepath.Join(c.opts.ConfigurationDirectory, defaultProvisioningTokenFile)
	c.opts.FactsFile = filepath.Join(c.opts.ConfigurationDirectory, defaultFactsFile)
	c.opts.ServerStorageDirectory = defaultStorageDirectory
	c.opts.NatsNeySeedFile = filepath.Join(c.opts.ConfigurationDirectory, defaultNatsNkeyFile)
	c.opts.NatsCredentialsFile = filepath.Join(c.opts.ConfigurationDirectory, defaultNatsCredentialFile)

	build.ProvisionJWTFile = c.opts.ProvisioningJWTFile

	log := logrus.New()
	switch {
	case strings.ToLower(c.logfile) == "discard":
		log.SetOutput(io.Discard)

	case c.logfile != "":
		log.Formatter = &logrus.JSONFormatter{}

		file, err := os.OpenFile(c.logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0660)
		if err != nil {
			return nil, nil, fmt.Errorf("could not set up logging: %s", err)
		}

		log.SetOutput(file)
	}
	c.log = logrus.NewEntry(log)

	switch c.loglevel {
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	case "fatal":
		log.SetLevel(logrus.FatalLevel)
	default:
		log.SetLevel(logrus.WarnLevel)
	}

	if c.debug {
		log.SetLevel(logrus.DebugLevel)
	}

	err = loadPlugins(c.opts, log.WithField("stage", "plugins"))
	if err != nil {
		return nil, nil, fmt.Errorf("loading plugins failed: %v", err)
	}

	go c.interruptWatcher()

	return c.opts.roCopy(), c.log, nil
}

func (c *cliInstance) validateOptions() error {
	if c.opts.Help == "" {
		c.opts.Help = defaultHelp
	}

	if c.opts.Version == "" {
		c.opts.Version = version
	}

	if c.opts.Name == "" {
		c.opts.Name = defaultName
	}

	if c.opts.FactsRefreshInterval < time.Minute {
		c.opts.FactsRefreshInterval = defaultFactsRefresh
	}

	var err error
	if c.opts.CommandPath == "" {
		c.opts.CommandPath, err = filepath.Abs(os.Args[0])
		if err != nil {
			return fmt.Errorf("could not determine path to command: %v", err)
		}
	}

	if c.opts.MachineSigningKey == "" {
		return fmt.Errorf("autonomous agent signing key is required")
	}
	pk, err := hex.DecodeString(c.opts.MachineSigningKey)
	if err != nil {
		return fmt.Errorf("invalid autonomous agent signing key: %v", err)
	}
	if len(pk) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid autonomous agent signing key: incorrect length")
	}

	return nil
}

func (c *cliInstance) forceQuit() {
	<-time.After(defaultShutdownGrace)

	c.log.Errorf("Forcing shut-down after 10 second grace window")

	os.Exit(1)
}

func (c *cliInstance) interruptWatcher() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case sig := <-sigs:
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				go c.forceQuit()

				c.log.Warnf("Shutting down on interrupt")

				c.cancel()
				return
			}

		case <-c.ctx.Done():
			return
		}
	}
}

func (c *cliInstance) createServerNKey() error {
	if c.opts.NatsNeySeedFile == "" {
		return fmt.Errorf("no nkey seed configured")
	}
	if choria.FileExist(c.opts.NatsNeySeedFile) {
		return nil
	}

	ukp, err := nkeys.CreateUser()
	if err != nil {
		return fmt.Errorf("could not generate user nkey: %v", err)
	}
	ukps, err := ukp.Seed()
	if err != nil {
		return fmt.Errorf("could not generate user nkey: %v", err)
	}
	err = os.WriteFile(c.opts.NatsNeySeedFile, ukps, 0400)
	if err != nil {
		return fmt.Errorf("could not generate user nkey: %v", err)
	}

	return nil
}
