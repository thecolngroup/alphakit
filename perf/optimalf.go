// Copyright 2022 The Coln Group Ltd
// SPDX-License-Identifier: MIT

package perf

import (
	"gonum.org/v1/gonum/floats"
)

// OptimalF is a function that returns the 'OptimalF' for a series of trade returns as defined by Ralph Vince.
// It is a method for sizing positions to maximize geometric return whilst accounting for biggest trading loss.
// See: https://www.investopedia.com/terms/o/optimalf.asp
// Param roundturns is the series of profits (-ve amount for losses) for each trade
func OptimalF(roundturns []float64) float64 {

	maxLoss := floats.Min(roundturns)
	var maxTWR, optimalF float64

	for i := 1.0; i <= 100.0; i++ {
		twr := 1.0
		f := i / 100
		for j := range roundturns {
			if roundturns[j] == 0 {
				continue
			}
			hpr := 1 + f*(-roundturns[j]/maxLoss)
			twr *= hpr
		}
		if twr > maxTWR {
			maxTWR = twr
			optimalF = f
		}
	}

	return optimalF
}
