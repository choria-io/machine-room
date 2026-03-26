// Copyright (c) R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package machineroom

import (
	"context"
	"encoding/hex"
	"os"
	"time"

	"github.com/choria-io/ccm/facts"
	"github.com/choria-io/ccm/manager"
	"github.com/choria-io/ccm/model"
	"github.com/choria-io/go-choria/choria"
	"github.com/choria-io/tokens"
	"github.com/nats-io/nkeys"
	"github.com/sirupsen/logrus"
)

func generateFacts(ctx context.Context, opts Options, log *logrus.Entry) (any, error) {
	cfg := model.NewFactsConfig()
	cfg.UserConfigDirectory = ""
	cfg.SystemConfigDirectory = ""
	cfg.NoHostFacts = opts.NoHostFacts
	cfg.NoNetworkFacts = opts.NoNetworkFacts
	cfg.NoSwapFacts = opts.NoSwapFacts
	cfg.NoMemoryFacts = opts.NoMemoryFacts
	cfg.NoCPUFacts = opts.NoCPUFacts
	cfg.NoPartitionFacts = opts.NoDiskFacts
	cfg.ExtraFactSources = append(cfg.ExtraFactSources, machineRoomFactProvider(opts, log))

	return facts.Gather(ctx, *cfg, manager.NewLogrusLogger(log))
}

func additionalFacts(ctx context.Context, opts Options, data map[string]any, log *logrus.Entry) {
	if opts.AdditionalFacts == nil {
		return
	}

	extra, err := opts.AdditionalFacts(ctx, opts.roCopy(), log)
	if err != nil {
		log.Errorf("Could not gather additional facts: %v", err)
		return
	}

	data["additional_facts"] = extra
}

func machineRoomFactProvider(opts Options, log *logrus.Entry) model.FactProvider {
	return func(ctx context.Context, _ model.FactsConfig, _ model.Logger) (map[string]any, error) {
		var err error

		fdata := map[string]any{}

		ext := tokens.MapClaims{}
		var provToken []byte
		if choria.FileExist(opts.ProvisioningJWTFile) {
			provToken, err = os.ReadFile(opts.ProvisioningJWTFile)
			if err == nil {
				t, err := tokens.ParseProvisionTokenUnverified(string(provToken))
				if err == nil {
					ext = t.Extensions
				}
			}
		}

		token := []byte{}
		pubKey := []byte{}
		pubNKey := ""

		if choria.FileExist(opts.ServerJWTFile) {
			token, err = os.ReadFile(opts.ServerJWTFile)
			if err != nil {
				log.Warnf("Could not read server token: %v", err)
			}
		}

		if choria.FileExist(opts.ServerSeedFile) {
			pubKey, _, err = choria.Ed25519KeyPairFromSeedFile(opts.ServerSeedFile)
			if err != nil {
				log.Warnf("Could not read server public key: %v", err)
			}
		}

		if choria.FileExist(opts.NatsNkeySeedFile) {
			pubNKey, err = loadNkeyPublic(opts)
			if err != nil {
				log.Warnf("Could not read nkey: %v", err)
			}
		}

		f := map[string]any{
			"identity":          opts.Identity,
			"timestamp":         time.Now(),
			"timestamp_seconds": time.Now().Unix(),
			"server": map[string]any{
				"token":       string(token),
				"public_key":  hex.EncodeToString(pubKey),
				"public_nkey": pubNKey,
			},
			"options": opts,
			"provisioning": map[string]any{
				"extended_claims": ext,
				"token":           string(provToken),
			},
		}

		if opts.AdditionalFacts != nil {
			additionalFacts(ctx, opts, f, log)
		}

		fdata["machine_room"] = f

		return fdata, nil
	}
}

func loadNkeyPublic(opts Options) (string, error) {
	seed, err := os.ReadFile(opts.NatsNkeySeedFile)
	if err != nil {
		return "", err
	}
	kp, err := nkeys.FromSeed(seed)
	if err != nil {
		return "", err
	}
	return kp.PublicKey()
}
