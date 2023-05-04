// Copyright (c) R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package machineroom

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/choria-io/go-choria/build"
	"github.com/choria-io/go-choria/choria"
	"github.com/choria-io/go-choria/config"
	"github.com/choria-io/go-choria/providers/provtarget"
	cs "github.com/choria-io/go-choria/server"
	"github.com/choria-io/machine-room/internal/autoagents/factsrefresh"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
	"github.com/sirupsen/logrus"
)

type server struct {
	cfg    *config.Config
	bi     *build.Info
	fw     *choria.Framework
	opts   *Options
	inproc nats.InProcessConnProvider
	log    *logrus.Entry
}

func newServer(opts *Options, configFile string, inproc nats.InProcessConnProvider, log *logrus.Entry) (*server, error) {
	if configFile == "" {
		return nil, fmt.Errorf("configuration file is required")
	}

	var err error
	srv := &server{
		bi:   &build.Info{},
		opts: opts,
		log:  log.WithField("machine_room", "server"),
	}

	srv.bi.SetProvisionJWTFile(opts.ProvisioningJWTFile)
	srv.bi.SetProvisionUsingVersion2(false)
	srv.bi.EnableProvisionModeAsDefault()
	srv.bi.SetProvisionFacts(opts.FactsFile)
	build.Version = opts.Version // TODO: wrap in bi

	hasRequiredFiles := choria.FileExist(configFile) && choria.FileExist(opts.ServerJWTFile) && choria.FileExist(opts.ServerSeedFile) && choria.FileExist(opts.NatsNeySeedFile)

	switch {
	case hasRequiredFiles:
		srv.cfg, err = config.NewSystemConfig(configFile, true)
		if err != nil {
			log.Errorf("Could not parse configuration, forcing reprovision: %v", err)
		}

		if srv.shouldProvision() {
			provtarget.Configure(context.Background(), srv.cfg, srv.log.WithField("component", "provtarget"))

			log.Warnf("Switching to provisioning configuration due to build defaults and configuration settings")
			srv.cfg, err = srv.provisionConfig(configFile, srv.bi)
			if err != nil {
				return nil, err
			}
		} else {
			srv.cfg.CustomLogger = srv.log.Logger

			// auto agents are always on
			srv.cfg.Choria.MachineSourceDir = opts.MachinesDirectory
			srv.cfg.Choria.MachinesSignerPublicKey = opts.MachineSigningKey

			// standard status file always
			srv.cfg.Choria.StatusFilePath = opts.ServerStatusFile

			// message submit for auto agents etc
			srv.cfg.Choria.SubmissionSpoolMaxSize = opts.ServerSubmissionSpoolSize
			srv.cfg.Choria.SubmissionSpool = opts.ServerSubmissionDirectory

			// some settings we need to not forget in provisioning helper
			srv.cfg.Choria.UseSRVRecords = false
			srv.cfg.RegisterInterval = 300
			srv.cfg.RegistrationSplay = true
			srv.cfg.FactSourceFile = opts.FactsFile
			srv.cfg.Choria.InventoryContentRegistrationTarget = "choria.broadcast.agent.registration"
			srv.cfg.Registration = []string{"inventory_content"}
			srv.cfg.Collectives = []string{"choria"}
			srv.cfg.MainCollective = "choria"

			srv.cfg.Choria.SecurityProvider = "choria"
			srv.cfg.Choria.ChoriaSecurityTokenFile = opts.ServerJWTFile
			srv.cfg.Choria.ChoriaSecuritySeedFile = opts.ServerSeedFile

			err = os.MkdirAll(srv.cfg.Choria.MachineSourceDir, 0700)
			if err != nil {
				srv.log.Warnf("Could not create machine source directory: %v", err)
			}

			err = srv.saveCredentials()
			if err != nil {
				srv.log.Errorf("Could not save NATS credentials: %v", err)
			}

			err = factsrefresh.Register(opts.CommandPath, opts.Version, opts.FactsRefreshInterval, configFile)
			if err != nil {
				srv.log.Errorf("Could not register facts refresh autonomous agent: %v", err)
			}
		}

	default:
		err = srv.createServerNKey()
		if err != nil {
			return nil, err
		}

		srv.cfg, err = srv.provisionConfig(configFile, srv.bi)
		if err != nil {
			return nil, err
		}
		srv.cfg.CustomLogger = srv.log.Logger
		provtarget.Configure(context.Background(), srv.cfg, srv.log.WithField("component", "provtarget"))

		log.Warnf("Switching to provisioning configuration due to build defaults and missing %s", configFile)
	}

	srv.cfg.ApplyBuildSettings(srv.bi)

	srv.fw, err = choria.NewWithConfig(srv.cfg)
	if err != nil {
		return nil, err
	}

	if inproc != nil {
		srv.fw.SetInProcessConnProvider(inproc)
	}

	return srv, nil
}

func (s *server) Start(ctx context.Context, wg *sync.WaitGroup) error {
	s.fw.ConfigureProvisioning(ctx)

	instance, err := cs.NewInstance(s.fw)
	if err != nil {
		return fmt.Errorf("could not create Choria Machine Room Server instance: %s", err)
	}

	if s.opts.ReadyFunc != nil {
		instance.RegisterReadyCallback(func(ctx context.Context) {
			s.opts.ReadyFunc(ctx, s.opts.roCopy(), s.log.WithField("machine_room", "readyfunc"))
		})
	}

	wg.Add(1)
	go func() {
		err := instance.Run(ctx, wg)
		if err != nil {
			s.log.Errorf("Server instance failed to start: %v", err)
		}
	}()

	return nil
}

func (s *server) IsProvisioning() bool {
	return s.fw.ProvisionMode()
}

func (s *server) shouldProvision() bool {
	if s.cfg == nil {
		return true
	}

	should := true
	if s.cfg.HasOption("plugin.choria.server.provision") {
		should = s.cfg.Choria.Provision
	}

	return should
}

func (s *server) saveCredentials() error {
	njwt := s.cfg.Option(configKeySourceNatsJwt, "")
	if njwt == "" {
		return nil
	}

	nkBytes, err := os.ReadFile(s.opts.NatsNeySeedFile)
	if err != nil {
		return err
	}

	cred, err := jwt.FormatUserConfig(njwt, nkBytes)
	if err != nil {
		return err
	}

	return os.WriteFile(s.opts.NatsCredentialsFile, cred, 0600)
}

func (s *server) createServerNKey() error {
	if s.opts.NatsNeySeedFile == "" {
		return fmt.Errorf("no nkey seed configured")
	}
	if choria.FileExist(s.opts.NatsNeySeedFile) {
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
	err = os.WriteFile(s.opts.NatsNeySeedFile, ukps, 0600)
	if err != nil {
		return fmt.Errorf("could not generate user nkey: %v", err)
	}

	return nil
}

func (s *server) provisionConfig(f string, bi *build.Info) (*config.Config, error) {
	if !choria.FileExist(bi.ProvisionJWTFile()) {
		return nil, fmt.Errorf("provisioming token not found in %s", bi.ProvisionJWTFile())
	}

	err := s.createServerNKey()
	if err != nil {
		return nil, fmt.Errorf("could not create nkey: %w", err)
	}

	cfg, err := config.NewDefaultSystemConfig(true)
	if err != nil {
		return nil, fmt.Errorf("could not create default configuration for provisioning: %s", err)
	}

	cfg.ConfigFile = f

	// set this to avoid calling into puppet on non puppet machines
	// later ConfigureProvisioning() will do all the right things
	cfg.Choria.SecurityProvider = "file"

	// in provision mode we do not yet have certs and stuff so we disable these checks
	cfg.DisableSecurityProviderVerify = true

	cfg.Choria.UseSRVRecords = false

	return cfg, nil
}

func saveFacts(ctx context.Context, opts Options, log *logrus.Entry) error {
	data, err := generateFacts(ctx, opts, log)
	if err != nil {
		return err
	}

	j, err := json.Marshal(data)
	if err != nil {
		return err
	}

	log.Infof("Writing facts to %v", opts.FactsFile)

	return os.WriteFile(opts.FactsFile, j, 0600)
}
