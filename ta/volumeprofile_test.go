// Copyright 2022 The Coln Group Ltd
// SPDX-License-Identifier: MIT

package ta

import (
	"encoding/csv"
	"os"
	"path"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"github.com/thecolngroup/alphakit/market"
	"golang.org/x/exp/slices"
)

const testdataPath string = "./testdata/"

func TestMarketProfileWithPriceFile(t *testing.T) {

	file, _ := os.Open(path.Join(testdataPath, "BTCUSDT-1m-2022-05-06.csv"))
	defer func() {
		assert.NoError(t, file.Close())
	}()

	prices, err := market.NewCSVKlineReader(csv.NewReader(file)).ReadAll()
	assert.NoError(t, err)
	var levels []VolumeLevel
	for i := range prices {

		hlc3 := HLC3(prices[i])
		vol := prices[i].Volume

		if hlc3 == 0 || vol == 0 {
			continue
		}

		levels = append(levels, VolumeLevel{
			Price:  hlc3,
			Volume: vol,
		})
	}
	slices.SortStableFunc(levels, func(i, j VolumeLevel) bool {
		return i.Price < j.Price
	})

	spew.Dump(prices[0].Start, prices[len(prices)-1].Start)
	vp := NewVolumeProfile(10, levels)

	assert.Equal(t, 35359.26666666667, vp.Low)
	assert.Equal(t, 35569.12777777779, vp.VAL)
	assert.Equal(t, 35988.850000000006, vp.POC)
	assert.Equal(t, 36408.572222222225, vp.VAH)
	assert.Equal(t, 36617.433333333334, vp.High)
}

func TestVolumeProfile_FixedBinWidth(t *testing.T) {
	file, _ := os.Open(path.Join(testdataPath, "BTCUSDT-1m-2022-05-06.csv"))
	defer func() {
		assert.NoError(t, file.Close())
	}()

	prices, err := market.NewCSVKlineReader(csv.NewReader(file)).ReadAll()
	assert.NoError(t, err)
	var levels []VolumeLevel
	for i := range prices {

		hlc3 := HLC3(prices[i])
		vol := prices[i].Volume

		if hlc3 == 0 || vol == 0 {
			continue
		}

		levels = append(levels, VolumeLevel{
			Price:  hlc3,
			Volume: vol,
		})
	}
	slices.SortStableFunc(levels, func(i, j VolumeLevel) bool {
		return i.Price < j.Price
	})

	spew.Dump(prices[0].Start, prices[len(prices)-1].Start)
	vp := NewVolumeProfileFixedBinWidth(10, levels)
	assert.Equal(t, 35359.26666666667, vp.Low)
	assert.Equal(t, 35747.090000000004, vp.VAL)
	assert.Equal(t, 36039.21666666667, vp.POC)
	assert.Equal(t, 36311.19666666667, vp.VAH)
	assert.Equal(t, 36617.433333333334, vp.High)
}
