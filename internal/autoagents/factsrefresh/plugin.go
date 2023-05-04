// Copyright (c) R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package factsrefresh

import (
	"fmt"
	"time"

	"github.com/choria-io/go-choria/aagent/machine"
	mp "github.com/choria-io/go-choria/aagent/plugin"
	"github.com/choria-io/go-choria/aagent/watchers"
	"github.com/choria-io/go-choria/plugin"
)

func Register(cmdPath string, version string, interval time.Duration, cfgFile string) error {
	if cmdPath == "" {
		return fmt.Errorf("no command path set in options")
	}

	m := &machine.Machine{
		MachineName:    "facts_refresh",
		MachineVersion: version,
		InitialState:   "GATHER",
		Transitions: []*machine.Transition{
			{
				Name:        "MAINTENANCE",
				From:        []string{"GATHER"},
				Destination: "MAINTENANCE",
			},
			{
				Name:        "RESUME",
				From:        []string{"MAINTENANCE"},
				Destination: "GATHER",
			},
		},
		WatcherDefs: []*watchers.WatcherDef{
			{
				Name:       "update_facts",
				Type:       "exec",
				Interval:   interval.String(),
				StateMatch: []string{"GATHER"},
				Properties: map[string]any{
					"command":              fmt.Sprintf("%s facts --config %s", cmdPath, cfgFile),
					"timeout":              "1m",
					"gather_initial_state": "true",
				},
			},
		},
	}

	return plugin.Register("facts_refresh", mp.NewMachinePlugin("facts_refresh", m))
}
