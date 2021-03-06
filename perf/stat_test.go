// Copyright 2022 The Coln Group Ltd
// SPDX-License-Identifier: MIT

package perf

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thecolngroup/gou/num"
)

func TestCAGR(t *testing.T) {
	giveInitial := 1000.0
	giveFinal := 2500.0
	giveDays := 190
	want := 4.81
	act := CAGR(giveInitial, giveFinal, giveDays)
	assert.Equal(t, want, num.Round2DP(act))
}

func TestKellyCriterion(t *testing.T) {
	giveProfitFactor := 1.6
	giveWinP := 0.7
	want := 0.51
	act := KellyCriterion(giveProfitFactor, giveWinP)
	assert.Equal(t, want, num.Round2DP(act))
}

func TestSharpeRatio(t *testing.T) {
	give := []float64{0.1, 0.2, -0.15, 0.1, 0.8, -0.3, 0.2}
	exp := 6.2
	act := SharpeRatio(give, SharpeDefaultAnnualRiskFreeRate)
	assert.Equal(t, exp, num.Round2DP(act))
}

func TestHistVolAnn(t *testing.T) {
	give := []float64{0.1, 0.2, -0.15, 0.1, 0.8, -0.3, 0.2}
	exp := 5.51
	act := HistVolAnn(give)
	assert.Equal(t, exp, num.Round2DP(act))
}
