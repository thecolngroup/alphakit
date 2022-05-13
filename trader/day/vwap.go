package day

import (
<<<<<<< Updated upstream
	"math"

=======
	"github.com/colngroup/zero2algo/internal/util"
>>>>>>> Stashed changes
	"github.com/colngroup/zero2algo/market"
	"github.com/colngroup/zero2algo/ta"
	"github.com/davecgh/go-spew/spew"
)

//var _ ta.Indicator = (*VWAP)(nil)

// VWAP is a volume weighted average price.
type VWAP struct {
	cumPV  float64
	cumVol float64
	series []float64
}

// NewVWAP creates a new VWAP indicator with default parameters.
func NewVWAP() *VWAP {
	return &VWAP{}
}

// Update updates the indicator with the next value(s).
func (ind *VWAP) Update(prices ...market.Kline) error {

	for i := range prices {
<<<<<<< Updated upstream
		avgPrice := ta.HLC3(prices[i])
		vol := prices[i].Volume

		if avgPrice == 0 || vol == 0 {
			continue
		}

		ind.cumPV += avgPrice * vol
		ind.cumVol += vol

		vwap := ind.cumPV / ind.cumVol

=======

		hlc3 := ta.HLC3(prices[i])
		vol := util.NNZ(prices[i].Volume, 0.1)

		ind.cumPV += hlc3 * vol
		ind.cumVol += vol
		vwap := ind.cumPV / ind.cumVol

>>>>>>> Stashed changes
		ind.series = append(ind.series, vwap)

		if math.IsNaN(vwap) || math.IsInf(vwap, 0) {
			spew.Dump(ind)
			panic("NaN/inf")
		}
	}

	return nil
}

// Valid returns true if the indicator is valid.
// An indicator is invalid if it hasn't received enough values yet.
func (ind *VWAP) Valid() bool {
	return true
}

// Value returns the current value of the indicator.
func (ind *VWAP) Value() float64 {
	return ta.Lookback(ind.series, 0)
}

// History returns the historical values of the indicator.
func (ind *VWAP) History() []float64 {
	return ind.series
}
