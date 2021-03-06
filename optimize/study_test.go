// Copyright 2022 The Coln Group Ltd
// SPDX-License-Identifier: MIT

package optimize

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thecolngroup/alphakit/perf"
)

func TestSummarize(t *testing.T) {

	give := PhaseReport{
		Trials: []perf.PerformanceReport{
			{
				TradeReport:     &perf.TradeReport{PRR: 2.0, TradeCount: 5},
				PortfolioReport: &perf.PortfolioReport{MaxDrawdown: 0.3, CAGR: 0.8, Sharpe: 1.0, Calmar: 2.0},
			},
			{
				TradeReport:     &perf.TradeReport{PRR: 4.0, TradeCount: 10},
				PortfolioReport: &perf.PortfolioReport{MaxDrawdown: 0.2, CAGR: 1.5, Sharpe: 2.0, Calmar: 2.0},
			},
		},
	}
	want := PhaseReport{
		PRR:            3.0,
		MDD:            0.25,
		CAGR:           1.15,
		Sharpe:         1.5,
		Calmar:         2.0,
		SampleCount:    2,
		RoundTurnCount: 15,
		Trials: []perf.PerformanceReport{
			{
				TradeReport:     &perf.TradeReport{PRR: 2.0, TradeCount: 5},
				PortfolioReport: &perf.PortfolioReport{MaxDrawdown: 0.3, CAGR: 0.8, Sharpe: 1.0, Calmar: 2.0},
			},
			{
				TradeReport:     &perf.TradeReport{PRR: 4.0, TradeCount: 10},
				PortfolioReport: &perf.PortfolioReport{MaxDrawdown: 0.2, CAGR: 1.5, Sharpe: 2.0, Calmar: 2.0},
			},
		},
	}

	act := Summarize(give)
	assert.Equal(t, want, act)
}
