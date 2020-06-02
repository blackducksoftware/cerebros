package stress_testing

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
	"math"
	"time"
)

// TODO Amp is a horrible name.  Come up with something better
//   This type is for a function that takes in a time in seconds, and outputs an amplitude
type Amp = func(float64) float64

func Const(baseline float64) Amp {
	return func(float64) float64 {
		return baseline
	}
}

func Linear(slope float64) Amp {
	return func(seconds float64) float64 {
		return seconds * slope
	}
}

func Sinusoid(baseline float64, amplitude float64, period float64, phase float64) Amp {
	if baseline-amplitude < 0 {
		panic(fmt.Sprintf("baseline of %f and amplitude of %f will produce a negative value", baseline, amplitude))
	}
	return func(seconds float64) float64 {
		// b + A * sin(phase + (2 pi t / period))
		return baseline + amplitude*math.Sin(phase+2*math.Pi*seconds/period)
	}
}

func ShiftRight(seconds float64, f Amp) Amp {
	return func(t float64) float64 {
		return f(t - seconds)
	}
}

func Add(fs ...Amp) Amp {
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
	Amp             Amp
}

func Piecewise(pieces []*Piece) Amp {
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
				return p.Amp(seconds)
			}
			seconds -= p.DurationSeconds
		}
		panic("intervals didn't add up correctly")
	}
}

func Spike(baseline float64, lowDuration float64, height float64, highDuration float64, ramp float64) Amp {
	pieces := []*Piece{
		{DurationSeconds: lowDuration, Amp: Const(0)},
		{DurationSeconds: ramp, Amp: Linear(height / ramp)},
		{DurationSeconds: highDuration, Amp: Const(height)},
		{DurationSeconds: ramp, Amp: Add(Linear(-height/ramp), Const(height))},
	}
	return Add(Piecewise(pieces), Const(baseline))
}

func AmplitudeFromArray(vals []float64, durationSeconds float64) Amp {
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
	Name             string
	rateLimiter      *rate.Limiter
	stop             chan struct{}
	rateChangePeriod time.Duration
	getRate          Amp
}

func NewRateLimiter(name string, getRate Amp, rateChangePeriod time.Duration) *RateLimiter {
	rl := &RateLimiter{
		Name:             name,
		rateLimiter:      rate.NewLimiter(rate.Limit(0), 1),
		stop:             make(chan struct{}),
		rateChangePeriod: rateChangePeriod,
		getRate:          getRate,
	}
	rl.setLimitForTime(0)
	rl.start()
	return rl
}

func (rl *RateLimiter) Wait() error {
	return rl.rateLimiter.Wait(context.TODO())
}

func (rl *RateLimiter) Stop() {
	log.Infof("stopping RateLimiter %s", rl.Name)
	close(rl.stop)
}

func (rl *RateLimiter) Limit() float64 {
	return float64(rl.rateLimiter.Limit())
}

func (rl *RateLimiter) setLimitForTime(seconds int) {
	newLimit := rl.getRate(float64(seconds) * rl.rateChangePeriod.Seconds())
	log.Infof("updating RateLimiter %s limit to %f for time %d", rl.Name, newLimit, seconds)
	rl.rateLimiter.SetLimit(rate.Limit(newLimit))
	recordNamedEventGauge("rateLimit", rl.Name, newLimit)
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
