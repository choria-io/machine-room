// Copyright (c) R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package machineroom

import (
	"fmt"
	"sync"

	"github.com/choria-io/go-choria/aagent/data/machinedata"
	archivewatcher "github.com/choria-io/go-choria/aagent/watchers/archivewatcher"
	"github.com/choria-io/go-choria/aagent/watchers/execwatcher"
	"github.com/choria-io/go-choria/aagent/watchers/filewatcher"
	"github.com/choria-io/go-choria/aagent/watchers/kvwatcher"
	"github.com/choria-io/go-choria/aagent/watchers/metricwatcher"
	"github.com/choria-io/go-choria/aagent/watchers/nagioswatcher"
	"github.com/choria-io/go-choria/aagent/watchers/pluginswatcher"
	"github.com/choria-io/go-choria/aagent/watchers/schedulewatcher"
	"github.com/choria-io/go-choria/aagent/watchers/timerwatcher"
	"github.com/choria-io/go-choria/plugin"
	golangrpc "github.com/choria-io/go-choria/providers/agent/mcorpc/golang"
	provisioner "github.com/choria-io/go-choria/providers/agent/mcorpc/golang/provision"
	"github.com/choria-io/go-choria/providers/data/golang/choriadata"
	"github.com/choria-io/go-choria/scout/data/scoutdata"
	"github.com/choria-io/machine-room/internal/autoagents/machinesmanager"
	"github.com/sirupsen/logrus"
)

var (
	defaultPlugins = map[string]plugin.Pluggable{
		"choria_provision":        provisioner.ChoriaPlugin(),
		"agent_provider_golang":   golangrpc.ChoriaPlugin(),
		"watcher_archive":         archivewatcher.ChoriaPlugin(),
		"watcher_exec":            execwatcher.ChoriaPlugin(),
		"watcher_file":            filewatcher.ChoriaPlugin(),
		"watcher_kv":              kvwatcher.ChoriaPlugin(),
		"watcher_metric":          metricwatcher.ChoriaPlugin(),
		"watcher_nagios":          nagioswatcher.ChoriaPlugin(),
		"watcher_plugins":         pluginswatcher.ChoriaPlugin(),
		"watcher_schedule":        schedulewatcher.ChoriaPlugin(),
		"watcher_timer":           timerwatcher.ChoriaPlugin(),
		"machine_plugins_manager": machines_manager.ChoriaPlugin(),
		"data_choria":             choriadata.ChoriaPlugin(),
		"data_machine":            machinedata.ChoriaPlugin(),
		"data_scout":              scoutdata.ChoriaPlugin(),
	}

	mu     sync.Mutex
	loaded bool
)

func loadPlugins(opts *Options, log *logrus.Entry) error {
	mu.Lock()
	defer mu.Unlock()

	if loaded {
		return nil
	}

	loaded = true

	for k, v := range defaultPlugins {
		log.Infof("Registering plugin %s %s: %s", v.PluginType().String(), k, v.PluginName())
		err := plugin.Register(k, v)
		if err != nil {
			return fmt.Errorf("plugin %v failed: %v", v.PluginName(), err)
		}
	}

	for k, v := range opts.Plugins {
		err := plugin.Register(k, v)
		if err != nil {
			log.Infof("Registering plugin %s %s: %s", v.PluginType().String(), k, v.PluginName())
			return fmt.Errorf("plugin %v failed: %v", k, err)
		}
	}

	return nil
}
