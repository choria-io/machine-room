// Copyright (c) R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package machineroom

import (
	"context"
	"encoding/hex"
	"os"
	"time"

	"github.com/choria-io/go-choria/choria"
	"github.com/choria-io/tokens"
	"github.com/nats-io/nkeys"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/sirupsen/logrus"
)

func generateFacts(ctx context.Context, opts Options, log *logrus.Entry) (any, error) {
	data := map[string]map[string]any{
		"machine_room": {},
		"host":         {},
		"mem":          {},
		"swap":         {},
		"cpu":          {},
		"disk":         {},
		"net":          {},
	}

	machineRoomFacts(opts, data, log)
	additionalFacts(ctx, opts, data, log)
	standardFacts(ctx, opts, data, log)

	return data, nil
}

func additionalFacts(ctx context.Context, opts Options, data map[string]map[string]any, log *logrus.Entry) {
	if opts.AdditionalFacts == nil {
		return
	}

	extra, err := opts.AdditionalFacts(ctx, opts.roCopy(), log)
	if err != nil {
		log.Errorf("Could not gather additional facts: %v", err)
	} else {
		data["machine_room"]["additional_facts"] = extra
	}
}

func machineRoomFacts(opts Options, data map[string]map[string]any, log *logrus.Entry) {
	var err error

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

	if choria.FileExist(opts.NatsNeySeedFile) {
		pubNKey, err = loadNkeyPublic(opts)
		if err != nil {
			log.Warnf("Could not read nkey: %v", err)
		}
	}

	data["machine_room"] = map[string]any{
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
}

func loadNkeyPublic(opts Options) (string, error) {
	seed, err := os.ReadFile(opts.NatsNeySeedFile)
	if err != nil {
		return "", err
	}
	kp, err := nkeys.FromSeed(seed)
	if err != nil {
		return "", err
	}
	return kp.PublicKey()
}

func standardFacts(ctx context.Context, opts Options, data map[string]map[string]any, log *logrus.Entry) {
	if opts.NoStandardFacts {
		return
	}

	var err error

	if !opts.NoMemoryFacts {
		data["mem"]["virtual"], err = mem.VirtualMemoryWithContext(ctx)
		if err != nil {
			log.Warnf("Could not gather virtual memory information: %v", err)
		}

		data["swap"]["memory"], err = mem.SwapMemoryWithContext(ctx)
		if err != nil {
			log.Warnf("Could not gather swap information: %v", err)
		}
	}

	if !opts.NoCPUFacts {
		data["cpu"]["info"], err = cpu.InfoWithContext(ctx)
		if err != nil {
			log.Warnf("Could not gather CPU information: %v", err)
		}
	}

	if !opts.NoDiskFacts {
		parts, err := disk.PartitionsWithContext(ctx, false)
		if err != nil {
			log.Warnf("Could not gather Disk partitions: %v", err)
		}
		if len(parts) > 0 {
			matchedParts := []disk.PartitionStat{}
			usages := []*disk.UsageStat{}

			for _, part := range parts {
				matchedParts = append(matchedParts, part)
				u, err := disk.UsageWithContext(ctx, part.Mountpoint)
				if err != nil {
					log.Warnf("Could not get usage for partition %s: %v", part.Mountpoint, err)
					continue
				}
				usages = append(usages, u)
			}

			data["disk"]["partitions"] = matchedParts
			data["disk"]["usage"] = usages
		}
	}

	if !opts.NoHostFacts {
		data["host"]["info"], err = host.InfoWithContext(ctx)
		if err != nil {
			log.Warnf("Could not gather host information: %v", err)
		}
	}

	if !opts.NoNetworkFacts {
		data["net"]["interfaces"], err = net.InterfacesWithContext(ctx)
		if err != nil {
			log.Warnf("Could not gather network interfaces: %v", err)
		}
	}
}
