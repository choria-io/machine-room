// Copyright (c) R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package machineroom

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/choria-io/fisk"
)

func (c *cliInstance) resetCommand(_ *fisk.ParseContext) error {
	_, log, err := c.CommonConfigure()
	if err != nil {
		return err
	}

	opts := c.opts

	log.Warnf("Ensure that the process is stopped prior to resetting")

	if !c.force {
		var ok bool
		err = survey.AskOne(&survey.Confirm{
			Message: fmt.Sprintf("Really reset the %s agent", opts.Name),
		}, &ok)
		if err != nil {
			return err
		}

		if !ok {
			fmt.Println("Cancelling reset operation")
			return nil
		}
	}

	log.Warnf("Removing state storage directory %s", opts.ServerStorageDirectory)
	err = os.RemoveAll(opts.ServerStorageDirectory)
	if err != nil {
		log.Errorf("Could not remove storage directory: %v", err)
	}

	log.Warnf("Removing autonomous agent store %s", opts.MachinesDirectory)
	err = os.RemoveAll(opts.MachinesDirectory)
	if err != nil {
		log.Errorf("Could not remove autonomous agent store: %v", err)
	}

	log.Warnf("Removing instance facts file %s", opts.FactsFile)
	err = os.Remove(opts.FactsFile)
	if err != nil {
		log.Warnf("Could not remove facts file: %v", err)
	}

	log.Warnf("Removing JWT file %s", opts.ServerJWTFile)
	err = os.Remove(opts.ServerJWTFile)
	if err != nil {
		log.Errorf("Could not remove jwt file: %v", err)
	}

	log.Warnf("Removing Seed file %s", opts.ServerJWTFile)
	err = os.Remove(opts.ServerSeedFile)
	if err != nil {
		log.Errorf("Could not remove seed file: %v", err)
	}
	if opts.ConfigurationDirectory != "" {
		for _, f := range []string{defaultCaFile, defaultCertFile, defaultKeyFile, defaultNatsNkeyFile, defaultNatsCredentialFile} {
			path := filepath.Join(opts.ConfigurationDirectory, f)
			log.Warnf("Removing x509 file %v", path)
			err = os.Remove(path)
			if err != nil {
				log.Errorf("Could not remove %s: %v", f, err)
			}
		}
	}

	log.Warnf("Removing configuration file %v", c.cfgFile)
	err = os.Remove(c.cfgFile)
	if err != nil {
		log.Warnf("Could not remove configuration file: %v", err)
	}

	return nil
}
