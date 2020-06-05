package stress_testing

import (
	"context"
	"fmt"
	"github.com/paulbellamy/ratecounter"
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
	"math"
	"sync"
	"time"
)

const (
	rateCounterPeriod = 1 * time.Minute
)

type AdaptiveRateAdjuster func(currentLimit float64, currentRate float64, currentErrorFraction float64) float64

type ErrorFractionThresholdConfig struct {
	IncreaseRatio            float64
	IncreaseMaxErrorFraction float64
	DecreaseRatio            float64
	DecreaseMinErrorFraction float64
	MaxRate                  float64
	MinRate                  float64
}

func (conf *ErrorFractionThresholdConfig) RateAdjuster() AdaptiveRateAdjuster {
	if conf.IncreaseMaxErrorFraction < 0 || conf.IncreaseMaxErrorFraction > 1 {

	}
	return func(currentLimit float64, currentRate float64, errorFraction float64) float64 {
		if errorFraction > conf.DecreaseMinErrorFraction {
			// it's okay to keep decreasing the limit, even if the rate is far below the limit --
			// as long as the rate doesn't go below the min
			return math.Max(conf.MinRate, conf.DecreaseRatio*currentLimit)
		}
		// OTOH, don't keep increasing the limit if the rate is less than 80% of the limit
		if (errorFraction < conf.IncreaseMaxErrorFraction) && (currentRate > (currentLimit * 0.8)) {
			return math.Min(conf.MaxRate, conf.IncreaseRatio*currentLimit)
		}
		return currentLimit
	}
}

type BaseRateSetter = func(float64) float64

func Const(baseline float64) BaseRateSetter {
	return func(float64) float64 {
		return baseline
	}
}

func Linear(slope float64) BaseRateSetter {
	return func(seconds float64) float64 {
		return seconds * slope
	}
}

func Sinusoid(baseline float64, amplitude float64, period float64, phase float64) BaseRateSetter {
	if baseline-amplitude < 0 {
		panic(fmt.Sprintf("baseline of %f and amplitude of %f will produce a negative value", baseline, amplitude))
	}
	return func(seconds float64) float64 {
		// b + A * sin(phase + (2 pi t / period))
		return baseline + amplitude*math.Sin(phase+2*math.Pi*seconds/period)
	}
}

func ShiftRight(seconds float64, f BaseRateSetter) BaseRateSetter {
	return func(t float64) float64 {
		return f(t - seconds)
	}
}

func Add(fs ...BaseRateSetter) BaseRateSetter {
	return func(t float64) float64 {
		total := float64(0)
		for _, f := range fs {
			total += f(t)
		}
		return total
	}
}

type Piece struct {
	DurationSeconds float64
	BaseRateSetter  BaseRateSetter
}

func Piecewise(pieces []*Piece) BaseRateSetter {
	total := float64(0)
	for _, p := range pieces {
		total += p.DurationSeconds
	}
	return func(t float64) float64 {
		_, fraction := math.Modf(t / total)
		seconds := fraction * total
		runningTotal := float64(0)
		for ix, p := range pieces {
			runningTotal += p.DurationSeconds
			if seconds <= p.DurationSeconds {
				log.Tracef("piecewise: %f, %f, %f, %d, %f", t, fraction, seconds, ix, runningTotal)
				return p.BaseRateSetter(seconds)
			}
			seconds -= p.DurationSeconds
		}
		panic("intervals didn't add up correctly")
	}
}

func Spike(baseline float64, lowDuration float64, height float64, highDuration float64, ramp float64) BaseRateSetter {
	pieces := []*Piece{
		{DurationSeconds: lowDuration, BaseRateSetter: Const(0)},
		{DurationSeconds: ramp, BaseRateSetter: Linear(height / ramp)},
		{DurationSeconds: highDuration, BaseRateSetter: Const(height)},
		{DurationSeconds: ramp, BaseRateSetter: Add(Linear(-height/ramp), Const(height))},
	}
	return Add(Piecewise(pieces), Const(baseline))
}

func AmplitudeFromArray(vals []float64, durationSeconds float64) BaseRateSetter {
	totalLength := float64(len(vals)) * durationSeconds
	return func(t float64) float64 {
		// 1. handle wraparound if t is larger than total length
		_, fraction := math.Modf(t / totalLength)
		timeInPattern := fraction * totalLength
		// 2. figure out which bucket we're in
		bucket, _ := math.Modf(timeInPattern / durationSeconds)
		// 3. turn bucket into int by rounding down
		log.Infof("%f, %f, %f, %f, %d", t, fraction, timeInPattern, bucket, int(bucket))
		return vals[int(bucket)]
	}
}

type RateLimiter struct {
	Name              string
	rateLimiter       *rate.Limiter
	stop              chan struct{}
	rateChangePeriod  time.Duration
	startRateCounter  *ratecounter.RateCounter
	finishRateCounter *ratecounter.RateCounter
	errorCounter      *ratecounter.RateCounter
	successCount      int
	errorCount        int
	jobsInProgress    int
	getRate           BaseRateSetter
	errorAdjuster     AdaptiveRateAdjuster
	mux               *sync.Mutex
	stopChan          chan struct{}
}

func NewRateLimiter(name string, getRate BaseRateSetter, rateChangePeriod time.Duration, errorAdjuster AdaptiveRateAdjuster) *RateLimiter {
	rl := &RateLimiter{
		Name:              name,
		rateLimiter:       rate.NewLimiter(rate.Limit(0), 1),
		stop:              make(chan struct{}),
		rateChangePeriod:  rateChangePeriod,
		startRateCounter:  ratecounter.NewRateCounter(rateCounterPeriod),
		finishRateCounter: ratecounter.NewRateCounter(rateCounterPeriod),
		errorCounter:      ratecounter.NewRateCounter(rateCounterPeriod),
		successCount:      0,
		errorCount:        0,
		jobsInProgress:    0,
		getRate:           getRate,
		errorAdjuster:     errorAdjuster,
		mux:               &sync.Mutex{},
		stopChan:          make(chan struct{}),
	}
	rl.setLimitForTime(0)
	rl.start()
	go rl.recordMetrics()
	return rl
}

func (rl *RateLimiter) Wait() {
	waitName := fmt.Sprintf("%s-wait", rl.Name)
	start := time.Now()

	// this should never error
	doOrDie(rl.rateLimiter.Wait(context.TODO()))

	recordDuration(waitName, time.Since(start))
	rl.startRateCounter.Incr(1)
	rl.mux.Lock()
	rl.jobsInProgress++
	rl.mux.Unlock()
}

func (rl *RateLimiter) Finish(desc string, err error) {
	recordEvent(desc, err)

	rl.mux.Lock()
	defer rl.mux.Unlock()
	rl.jobsInProgress--
	rl.finishRateCounter.Incr(1)
	if err != nil {
		rl.errorCounter.Incr(1)
		rl.errorCount++
	} else {
		rl.successCount++
	}
}

func (rl *RateLimiter) Stop() {
	log.Infof("stopping RateLimiter %s", rl.Name)
	close(rl.stop)
}

func (rl *RateLimiter) Limit() float64 {
	return float64(rl.rateLimiter.Limit())
}

func (rl *RateLimiter) errorFraction() float64 {
	errorRate := float64(rl.errorCounter.Rate())
	total := float64(rl.finishRateCounter.Rate())
	log.Debugf("error fraction: %s, %f, %f, %f", rl.Name, errorRate, total, errorRate/total)
	return errorRate / total
}

func (rl *RateLimiter) recordMetrics() {
	for {
		select {
		case <-rl.stopChan:
			return
		case <-time.After(20 * time.Second):
			recordNamedEventGauge("errorFraction", rl.Name, rl.errorFraction())
			rl.mux.Lock()
			recordNamedEventGauge("jobsInProgress", rl.Name, float64(rl.jobsInProgress))
			recordNamedEventBy("errorCount", rl.Name, rl.errorCount)
			recordNamedEventBy("successCount", rl.Name, rl.successCount)
			rl.mux.Unlock()
		}
	}
}

func (rl *RateLimiter) setLimitForTime(t int) {
	seconds := float64(t) * rl.rateChangePeriod.Seconds()
	baseLimit := rl.getRate(seconds)
	newLimit := baseLimit
	var adaptiveAdjustment = float64(1)
	if rl.errorAdjuster != nil {
		adaptiveAdjustment = rl.errorAdjuster(
			baseLimit,
			float64(rl.finishRateCounter.Rate())/rateCounterPeriod.Seconds(),
			rl.errorFraction())
	}

	newLimit *= adaptiveAdjustment
	log.Infof("updating RateLimiter %s limit to %f for t %d, seconds %f", rl.Name, newLimit, t, seconds)
	rl.rateLimiter.SetLimit(rate.Limit(newLimit))

	recordNamedEventGauge("rateLimit", rl.Name, newLimit)
	recordNamedEventGauge("baseLimit", rl.Name, baseLimit)
	recordNamedEventGauge("adaptiveAdjustment", rl.Name, adaptiveAdjustment)
}

func (rl *RateLimiter) start() {
	log.Infof("starting RateLimiter %s", rl.Name)
	go func() {
	ForLoop:
		for i := 1; ; i++ {
			select {
			case <-rl.stop:
				break ForLoop
			case <-time.After(rl.rateChangePeriod):
			}
			rl.setLimitForTime(i)
		}
	}()
}
