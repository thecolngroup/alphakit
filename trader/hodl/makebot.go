// Copyright 2022 The Coln Group Ltd
// SPDX-License-Identifier: MIT

package hodl

import (
	"github.com/thecolngroup/alphakit/trader"
	"github.com/thecolngroup/gou/conv"
)

// MakeBotFromConfig builds a valid Bot from a given set of config params.
func MakeBotFromConfig(config map[string]any) (trader.Bot, error) {
	var hodl Bot

	if _, ok := config["buybarindex"]; !ok {
		return nil, trader.ErrInvalidConfig
	}
	buyBarIndex := conv.ToInt(config["buybarindex"])

	if _, ok := config["sellbarindex"]; !ok {
		return nil, trader.ErrInvalidConfig
	}
	sellBarIndex := conv.ToInt(config["sellbarindex"])

	switch {
	case buyBarIndex == 0 && sellBarIndex == 0:
		break
	case buyBarIndex >= 0 && sellBarIndex == 0:
		break
	case buyBarIndex < 0 || sellBarIndex < 0:
		return nil, trader.ErrInvalidConfig
	case buyBarIndex >= sellBarIndex:
		return nil, trader.ErrInvalidConfig
	}

	hodl.BuyBarIndex = buyBarIndex
	hodl.SellBarIndex = sellBarIndex

	return &hodl, nil
}
