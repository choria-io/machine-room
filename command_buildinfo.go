// Copyright (c) R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package machineroom

import (
	"encoding/json"
	"fmt"

	"github.com/choria-io/fisk"
	"github.com/choria-io/go-choria/build"
)

func (c *cliInstance) buildInfoCommand(_ *fisk.ParseContext) error {
	_, _, err := c.CommonConfigure()
	if err != nil {
		return err
	}

	bi := build.Info{}

	nfo := map[string]any{
		"providers": map[string]any{
			"agent":    bi.AgentProviders(),
			"watchers": bi.MachineWatchers(),
			"data":     bi.DataProviders(),
		},
	}

	j, err := json.MarshalIndent(&nfo, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(j))

	return nil
}
