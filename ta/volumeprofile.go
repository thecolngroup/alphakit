// Copyright 2022 The Coln Group Ltd
// SPDX-License-Identifier: MIT

package ta

import (
	"math"

	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/stat"
)

// DefaultValueAreaPercentage is the percentage of the total volume used to calculate the value area.
const DefaultValueAreaPercentage = 0.68

// VolumeProfile is a histogram of market price and volume.
// Intent is to show the price points with most volume during a period.
// The profile gives key features such as:
//
// Point of control (POC)
//
// Value area high (VAH)
//
// Value area low (VAL)
//
// Session High
//
// Session Low
type VolumeProfile struct {

	// Bins is the histogram bins.
	Bins []float64

	// Hist is the histogram values.
	Hist []float64

	// POC is the point of control.
	POC float64

	// VAH is the value area high.
	VAH float64

	// VAL is the value area low.
	VAL float64

	// High is the highest price in the profile.
	High float64

	// Low is the lowest price in the profile.
	Low float64
}

// VolumeLevel is a price and volume pair used to build a volume profile.
type VolumeLevel struct {

	// Price is the market price, typically the high/low average of the kline.
	Price float64

	// Volume is the total buy and sell volume at the price.
	Volume float64
}

// NewVolumeProfileFixedBinWidth creates a new volume profile with a variable number of bins dictated by binWidth.
// Returns nil if binWidth is not positive.
func NewVolumeProfileFixedBinWidth(binWidth float64, levels []VolumeLevel) *VolumeProfile {

	if binWidth <= 0 {
		return nil
	}

	var sortedPrices []float64
	for _, level := range levels {
		sortedPrices = append(sortedPrices, level.Price)
	}
	max := floats.Max(sortedPrices)
	min := floats.Min(sortedPrices)
	nBins := int(math.Floor((max-min)/binWidth)) + 1

	return NewVolumeProfile(nBins, levels)
}

// NewVolumeProfile creates a new profile for the price and volume series given by levels.
// nBins is the number of bins to use for the profile histogram.
func NewVolumeProfile(nBins int, levels []VolumeLevel) *VolumeProfile {

	var vp VolumeProfile

	var sortedPrices, volumes []float64
	for _, level := range levels {
		sortedPrices = append(sortedPrices, level.Price)
		volumes = append(volumes, level.Volume)
	}

	vp.High = floats.Max(sortedPrices)
	vp.Low = floats.Min(sortedPrices)
	vp.Bins = make([]float64, nBins)
	vp.Bins = floats.Span(vp.Bins, vp.Low, vp.High+1)
	vp.Hist = stat.Histogram(nil, vp.Bins, sortedPrices, volumes)

	pocIdx := floats.MaxIdx(vp.Hist)
	vp.POC = midBin(vp.Bins, pocIdx)

	// Calculate Value Area with POC as the centre point
	vaTotalVol := floats.Sum(volumes) * DefaultValueAreaPercentage
	vaCumVol := vp.Hist[pocIdx]
	var vahVol, valVol float64
	vahIdx, valIdx := pocIdx+1, pocIdx-1
	stepVAH, stepVAL := true, true

	for (vaCumVol <= vaTotalVol) &&
		(vahIdx <= len(vp.Hist)-1 && valIdx >= 0) {

		if stepVAH {
			vahVol = 0
			for vahVol == 0 && vahIdx+1 < len(vp.Hist)-1 {
				vahVol = vp.Hist[vahIdx] + vp.Hist[vahIdx+1]
				vahIdx += 2
			}
			stepVAH = false
		}

		if stepVAL {
			valVol = 0
			for valVol == 0 && valIdx-1 >= 0 {
				valVol = vp.Hist[valIdx] + vp.Hist[valIdx-1]
				valIdx -= 2
			}
			stepVAL = false
		}

		switch {
		case vahVol > valVol:
			vaCumVol += vahVol
			stepVAH, stepVAL = true, false
		case vahVol < valVol:
			vaCumVol += valVol
			stepVAH, stepVAL = false, true
		case vahVol == valVol:
			vaCumVol += valVol + vahVol
			stepVAH, stepVAL = true, true
		}

		if vahIdx >= len(vp.Hist)-1 {
			stepVAH = false
		}

		if valIdx <= 0 {
			stepVAL = false
		}
	}

	vp.VAH = midBin(vp.Bins, vahIdx)
	vp.VAL = midBin(vp.Bins, valIdx)

	return &vp
}

func midBin(bins []float64, idx int) float64 {

	if len(bins) == 0 {
		return math.NaN()
	}

	if idx >= len(bins)-1 {
		return bins[len(bins)-1]
	}

	if idx < 0 {
		return bins[0]
	}

	return stat.Mean([]float64{bins[idx], bins[idx+1]}, nil)

}
