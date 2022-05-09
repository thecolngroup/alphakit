package optimize

import (
	"context"
	"errors"
	"math"
	"runtime"

	"github.com/colngroup/zero2algo/broker"
	"github.com/colngroup/zero2algo/market"
	"github.com/colngroup/zero2algo/perf"
	"github.com/colngroup/zero2algo/trader"
	"github.com/gammazero/workerpool"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// BruteOptimizer implements a 'brute-force peak objective' optimization approach which
// tests all given parameter combinations and selects the highest ranked (peak) param set.
// Optima is selected by the given ObjectiveRanker func.
// Optimization method in 3 stages:
// Prepare:
// - Accept 1 or more price data samples
// - Split sample price data into in-sample (training) and out-of-sample (validation) datasets
// - Generate 1 or more param sets using the cartesian product of given ranges that define the param space
// Train:
// - Execute each algo param set over the in-sample price data
// - Average the performance for each param set over the in-sample data
// - Rank the param sets based on the performance objective (Profit Factor, Sharpe etc)
// Validate:
// - Execute the highest ranked ("trained") algo param set over the out-of-sample price data
// - Accept or reject the hypothesis based on statistical significance of the study report
type BruteOptimizer struct {
	SampleSplitPct float64
	WarmupBarCount int
	MakeBot        trader.MakeFromConfig
	MakeDealer     broker.MakeSimulatedDealer
	Ranker         ObjectiveRanker

	MaxWorkers int

	study Study
}

type bruteOptimizerJob struct {
	ParamSet       ParamSet
	Sample         []market.Kline
	WarmupBarCount int
	MakeBot        trader.MakeFromConfig
	MakeDealer     broker.MakeSimulatedDealer
}

func NewBruteOptimizer() BruteOptimizer {
	return BruteOptimizer{
		SampleSplitPct: 0,
		WarmupBarCount: 0,
		Ranker:         SharpeRanker,
		MaxWorkers:     runtime.NumCPU(),
		study:          NewStudy(),
	}
}

func (o *BruteOptimizer) Prepare(in ParamMap, samples [][]market.Kline) (int, error) {

	products := CartesianBuilder(in)
	for i := range products {
		pSet := NewParamSet()
		pSet.Params = ParamMap(products[i])
		o.study.Training = append(o.study.Training, pSet)
	}

	for i := range samples {
		training, validation := splitSample(samples[i], o.SampleSplitPct)
		o.study.TrainingSamples = append(o.study.TrainingSamples, training)
		o.study.ValidationSamples = append(o.study.ValidationSamples, validation)
	}

	steps := len(o.study.Training) * len(samples) // Training phase
	steps += len(samples)                         // Validation phase for optimum

	return steps, nil
}

func (o *BruteOptimizer) Start(ctx context.Context) (<-chan OptimizerStep, error) {

	outCh := make(chan OptimizerStep)

	// Helper to append results to each phase
	appendResult := func(phase Phase, results map[ParamSetID]Report, pset ParamSet, backtest perf.PerformanceReport) {
		report, ok := results[pset.ID]
		if !ok {
			report = NewReport()
			report.Subject = pset
			report.Phase = phase
		}
		backtest.Properties = pset.Params
		report.Backtests = append(report.Backtests, backtest)
		results[pset.ID] = report
	}

	go func() {
		defer close(outCh)

		doneCh := make(chan struct{})
		defer close(doneCh)

		// Training phase
		trainigJobCh := o.enqueueJobs(o.study.Training, o.study.TrainingSamples)
		trainingOutCh := processBruteJobs(ctx, doneCh, trainigJobCh, o.MaxWorkers)
		for step := range trainingOutCh {
			step.Phase = Training
			outCh <- step
			if step.Err != nil {
				if errors.Is(step.Err, trader.ErrInvalidConfig) {
					continue
				} else {
					return
				}
			}
			appendResult(Training, o.study.TrainingResults, step.PSet, step.Result)
		}

		// Summarize backtest results for each param set
		for k := range o.study.TrainingResults {
			o.study.TrainingResults[k] = Summarize(o.study.TrainingResults[k])
		}

		// Select top ranked result for validation phase
		results := maps.Values(o.study.TrainingResults)
		slices.SortFunc(results, o.Ranker)
		optima := results[len(results)-1].Subject
		o.study.Validation = append(o.study.Validation, optima)

		// Validation phase
		validationJobCh := o.enqueueJobs(o.study.Validation, o.study.ValidationSamples)
		validationOutCh := processBruteJobs(ctx, doneCh, validationJobCh, o.MaxWorkers)
		for step := range validationOutCh {
			step.Phase = Validation
			outCh <- step
			if step.Err != nil {
				return
			}
			appendResult(Validation, o.study.ValidationResults, step.PSet, step.Result)
		}
		o.study.ValidationResults[optima.ID] = Summarize(o.study.ValidationResults[optima.ID])
	}()

	return outCh, nil
}

func (o *BruteOptimizer) Study() Study {
	return o.study
}

func (o *BruteOptimizer) enqueueJobs(pSets []ParamSet, samples [][]market.Kline) <-chan bruteOptimizerJob {

	// A buffered channel enables us to enqueue jobs and close the channel in a single function to simplify the call flow
	// Without a buffer the loop would block awaiting a ready receiver for the jobs
	jobCh := make(chan bruteOptimizerJob, len(pSets)*len(samples))
	defer close(jobCh)

	// Enqueue a job for each pset and price series combination
	for i := range pSets {
		for j := range samples {
			jobCh <- bruteOptimizerJob{
				ParamSet:       pSets[i],
				Sample:         samples[j],
				WarmupBarCount: o.WarmupBarCount,
				MakeBot:        o.MakeBot,
				MakeDealer:     o.MakeDealer,
			}
		}
	}
	return jobCh
}

func processBruteJobs(ctx context.Context, doneCh <-chan struct{}, jobCh <-chan bruteOptimizerJob, maxWorkers int) <-chan OptimizerStep {

	outCh := make(chan OptimizerStep)

	go func() {
		defer close(outCh)

		wp := workerpool.New(maxWorkers)
		//var wg sync.WaitGroup
		next := true

		for next {
			select {
			case <-ctx.Done():
				outCh <- OptimizerStep{Err: ctx.Err()}
				next = false
			case <-doneCh:
				next = false
			case job, ok := <-jobCh:
				if !ok {
					next = false
					break
				}
				//wg.Add(1)
				wp.Submit(
					func() {
						//defer wg.Done()
						dealer, err := job.MakeDealer()
						if err != nil {
							outCh <- OptimizerStep{PSet: job.ParamSet, Err: err}
							return
						}
						bot, err := job.MakeBot(job.ParamSet.Params)
						if err != nil {
							outCh <- OptimizerStep{PSet: job.ParamSet, Err: err}
							return
						}
						bot.SetDealer(dealer)
						if err := bot.Warmup(ctx, job.Sample[:job.WarmupBarCount]); err != nil {
							outCh <- OptimizerStep{PSet: job.ParamSet, Err: err}
							return
						}
						perf, err := runBacktest(ctx, bot, dealer, job.Sample[job.WarmupBarCount:])
						outCh <- OptimizerStep{PSet: job.ParamSet, Result: perf, Err: err}
					})
			}
		}
		//wg.
		wp.StopWait()
	}()

	return outCh
}

func runBacktest(ctx context.Context, bot trader.Bot, dealer broker.SimulatedDealer, prices []market.Kline) (perf.PerformanceReport, error) {
	var empty perf.PerformanceReport

	for i := range prices {
		price := prices[i]
		if err := dealer.ReceivePrice(ctx, price); err != nil {
			return empty, err
		}
		if err := bot.ReceivePrice(ctx, price); err != nil {
			return empty, err
		}
	}

	if err := bot.Close(ctx); err != nil {
		return empty, err
	}

	trades, _, err := dealer.ListTrades(context.Background(), nil)
	if err != nil {
		return empty, err
	}
	equity := dealer.EquityHistory()
	report := perf.NewPerformanceReport(trades, equity)

	return report, nil
}

func splitSample(sample []market.Kline, splitPct float64) (a, b []market.Kline) {
	if splitPct == 0 {
		return sample, sample
	}

	splitIndex := float64(len(sample)) * splitPct
	splitIndex = math.Ceil(splitIndex)
	return sample[:int(splitIndex)], sample[int(splitIndex):]
}
