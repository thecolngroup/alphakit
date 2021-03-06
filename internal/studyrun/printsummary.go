// Copyright 2022 The Coln Group Ltd
// SPDX-License-Identifier: MIT

package studyrun

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/thecolngroup/alphakit/optimize"
	"github.com/thecolngroup/gou/conv"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

var _summaryHeader = []string{
	"PRR",
	"Vol",
	"MDD",
	"Wins",
	"CAGR",
	"Sharpe",
	"Calmar",
	"Kelly",
	"OptimalF",
	"Samples",
	"RoundTurns",
}

// printSummary prints a summary report to stdout.
func printSummary(report optimize.PhaseReport) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(_summaryHeader)
	table.Append([]string{
		fmt.Sprintf("%.2f", report.PRR),
		fmt.Sprintf("%.2f%%", report.HistVolAnn*100),
		fmt.Sprintf("%.2f%%", report.MDD*100),
		fmt.Sprintf("%.2f%%", report.WinPct*100),
		fmt.Sprintf("%.2f%%", report.CAGR*100),
		fmt.Sprintf("%.2f", report.Sharpe),
		fmt.Sprintf("%.2f", report.Calmar),
		fmt.Sprintf("%.2f", report.Kelly),
		fmt.Sprintf("%.2f", report.OptimalF),
		fmt.Sprintf("%d", report.SampleCount),
		fmt.Sprintf("%d", report.RoundTurnCount),
	})
	table.Render()

}

// printParams pretty prints a map.
func printParams(params map[string]any) {
	keys := maps.Keys(params)
	slices.Sort(keys)
	for _, k := range keys {
		fmt.Printf("%s: %s\n", k, conv.ToString(params[k]))
	}
}
