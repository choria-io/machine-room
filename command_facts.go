// Copyright (c) R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package machineroom

import (
	"context"
	"time"

	"github.com/choria-io/fisk"
)

func (c *cliInstance) factsCommand(_ *fisk.ParseContext) error {
	_, log, err := c.CommonConfigure()
	if err != nil {
		return err
	}

	to, cancel := context.WithTimeout(c.ctx, 30*time.Second)
	defer cancel()

	return saveFacts(to, *c.opts, log)
}
