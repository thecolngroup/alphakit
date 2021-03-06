// Copyright 2022 The Coln Group Ltd
// SPDX-License-Identifier: MIT

package trend

import (
	"context"
	"encoding/csv"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thecolngroup/alphakit/broker/backtest"
	"github.com/thecolngroup/alphakit/market"
	"github.com/thecolngroup/alphakit/money"
	"github.com/thecolngroup/alphakit/perf"
	"github.com/thecolngroup/alphakit/risk"
	"github.com/thecolngroup/alphakit/ta"
	"github.com/thecolngroup/gou/dec"
)

func TestBotWithCrossPredicter(t *testing.T) {
	dealer := backtest.NewDealer()
	dealer.SetInitialCapital(dec.New(1000))

	predicter := NewCrossPredicter(
		ta.NewOsc(ta.NewALMA(32), ta.NewALMA(64)),
		ta.NewMMIWithSmoother(200, ta.NewALMA(200)))

	bot := Bot{
		EnterLong:  1,
		ExitLong:   -0.9,
		EnterShort: -1,
		ExitShort:  0.9,
		Asset:      market.NewAsset("BTCUSDT"),
		dealer:     dealer,
		Predicter:  predicter,
		Risker:     risk.NewFullRisker(),
		Sizer:      money.NewFixedSizer(dec.New(1000)),
	}

	file, err := os.Open("./testdata/BTCUSDT-1h-2021-Q1.csv")
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, file.Close())
	}()

	prices, _ := market.NewCSVKlineReader(csv.NewReader(file)).ReadAll()
	for _, price := range prices {
		if err := dealer.ReceivePrice(context.Background(), price); err != nil {
			t.Fatal(err)
		}
		if err := bot.ReceivePrice(context.Background(), price); err != nil {
			t.Fatal(err)
		}
	}

	assert.NoError(t, bot.Close(context.Background()))

	roundturns, _, _ := dealer.ListRoundTurns(context.Background(), nil)
	equity := dealer.EquityHistory()
	report := perf.NewPerformanceReport(roundturns, equity)
	perf.PrintSummary(report)
}
