package ta

func Peak(series []float64) bool {
	t2, t1, t0 := Lookback(series, 2), Lookback(series, 1), Lookback(series, 0)
	prev := Slope(t2, t1)
	curr := Slope(t1, t0)
	return (prev == 1 || prev == 0) && curr == -1
}

func Valley(series []float64) bool {
	t2, t1, t0 := Lookback(series, 2), Lookback(series, 1), Lookback(series, 0)
	prev := Slope(t2, t1)
	curr := Slope(t1, t0)
	return (prev == -1 || prev == 0) && curr == 1
}
