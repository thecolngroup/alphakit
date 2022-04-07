package trend

import (
	"context"
	"testing"

	"github.com/colngroup/zero2algo/dec"
	"github.com/colngroup/zero2algo/market"
	"github.com/colngroup/zero2algo/ta"
	"github.com/stretchr/testify/assert"
)

func TestPredicter_ReceivePrice(t *testing.T) {
	var giveOsc, giveSD, giveMMI ta.MockIndicator
	giveOsc.On("Update", []float64{10}).Return(error(nil))
	giveSD.On("Update", []float64{10}).Return(error(nil))
	giveMMI.On("Update", []float64{7}).Return(error(nil))
	givePrice := market.Kline{C: dec.New(10)}
	givePrev := 3.0

	predicter := NewPredicter(&giveOsc, &giveSD, &giveMMI)
	predicter.prev = givePrev
	err := predicter.ReceivePrice(context.Background(), givePrice)

	assert.NoError(t, err)
	giveOsc.AssertExpectations(t)
	giveSD.AssertExpectations(t)
	giveMMI.AssertExpectations(t)
}

func TestPredicter_Predict(t *testing.T) {
	tests := []struct {
		name          string
		giveOscValues []float64
		giveSDValues  []float64
		giveMMIValues []float64
		want          float64
	}{
		{
			name:          "flat @ 0",
			giveOscValues: []float64{10, 10},
			giveSDValues:  []float64{0, 0},
			giveMMIValues: []float64{0, 0},
			want:          0,
		},
		{
			name:          "flat @ 0.1, MMI down-trend only",
			giveOscValues: []float64{10, 10},
			giveSDValues:  []float64{0, 0},
			giveMMIValues: []float64{75, 70},
			want:          0.1,
		},
		{
			name:          "long @ 1.0, cross up SD w/ MMI",
			giveOscValues: []float64{10, 20},
			giveSDValues:  []float64{15, 15},
			giveMMIValues: []float64{75, 70},
			want:          1.0,
		},
		{
			name:          "long @ 0.9, cross up SD w/no MMI",
			giveOscValues: []float64{10, 20},
			giveSDValues:  []float64{15, 15},
			giveMMIValues: []float64{70, 75},
			want:          0.9,
		},
		{
			name:          "long @ 0.7, cross up zero w/ MMI",
			giveOscValues: []float64{-10, 10},
			giveSDValues:  []float64{15, 15},
			giveMMIValues: []float64{75, 70},
			want:          0.7,
		},
		{
			name:          "long @ 0.6, cross up zero w/no MMI",
			giveOscValues: []float64{-10, 10},
			giveSDValues:  []float64{15, 15},
			giveMMIValues: []float64{70, 75},
			want:          0.6,
		},
		{
			name:          "short @ -1.0",
			giveOscValues: []float64{-10, -20},
			giveSDValues:  []float64{15, 15},
			giveMMIValues: []float64{75, 70},
			want:          -1.0,
		},
		{
			name:          "short @ -0.9",
			giveOscValues: []float64{-10, -20},
			giveSDValues:  []float64{15, 15},
			giveMMIValues: []float64{70, 75},
			want:          -0.9,
		},
		{
			name:          "short @ -0.7",
			giveOscValues: []float64{10, -10},
			giveSDValues:  []float64{15, 15},
			giveMMIValues: []float64{75, 70},
			want:          -0.7,
		},
		{
			name:          "short @ -0.6",
			giveOscValues: []float64{10, -10},
			giveSDValues:  []float64{15, 15},
			giveMMIValues: []float64{70, 70},
			want:          -0.6,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			predicter := NewPredicter(
				&ta.StubIndicator{Values: tt.giveOscValues},
				&ta.StubIndicator{Values: tt.giveSDValues},
				&ta.StubIndicator{Values: tt.giveMMIValues},
			)
			act := predicter.Predict()
			assert.Equal(t, tt.want, act)
		})
	}

}

func TestPredicter_Valid(t *testing.T) {
	predicter := NewPredicter(
		&ta.StubIndicator{IsValid: true},
		&ta.StubIndicator{IsValid: true},
		&ta.StubIndicator{IsValid: true},
	)

	want := true
	act := predicter.Valid()
	assert.Equal(t, want, act)
}
