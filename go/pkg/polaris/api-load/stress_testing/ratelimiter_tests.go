/*
Copyright (C) 2020 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownership. The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied. See the License for the
specific language governing permissions and limitations
under the License.
*/
package stress_testing

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"math"
	"time"
)

func RunRateLimiterTests() {
	Describe("RateLimiter", func() {
		It("const", func() {
			Expect(Const(3)(25)).To(Equal(float64(3)))
		})

		It("sinusoid", func() {
			sin := Sinusoid(2, 2, 1, 0)
			Expect(sin(0)).Should(BeNumerically("~", 2, 0.01))
			Expect(sin(0.25)).Should(BeNumerically("~", 4, 0.01))
			Expect(sin(0.5)).Should(BeNumerically("~", 2, 0.01))
			Expect(sin(0.75)).Should(BeNumerically("~", 0, 0.01))
			Expect(sin(1)).Should(BeNumerically("~", 2, 0.01))

			longerSin := Sinusoid(8, 2, 5, 0)
			Expect(longerSin(0)).Should(BeNumerically("~", 8, 0.01))
			Expect(longerSin(1.25)).Should(BeNumerically("~", 10, 0.01))
			Expect(longerSin(2.5)).Should(BeNumerically("~", 8, 0.01))
			Expect(longerSin(3.75)).Should(BeNumerically("~", 6, 0.01))
			Expect(longerSin(5)).Should(BeNumerically("~", 8, 0.01))

			phase := Sinusoid(20, 5, 10, math.Pi/2)
			Expect(phase(-2.5)).Should(BeNumerically("~", 20, 0.01))
			Expect(phase(0)).Should(BeNumerically("~", 25, 0.01))
			Expect(phase(2.5)).Should(BeNumerically("~", 20, 0.01))
			Expect(phase(5)).Should(BeNumerically("~", 15, 0.01))
			Expect(phase(7.5)).Should(BeNumerically("~", 20, 0.01))
		})

		It("spike", func() {
			spike := Spike(8, 10, 12, 2, 1)
			Expect(spike(0)).Should(BeNumerically("~", 8, 0.01))
			Expect(spike(5)).Should(BeNumerically("~", 8, 0.01))
			Expect(spike(10)).Should(BeNumerically("~", 8, 0.01))
			Expect(spike(10.25)).Should(BeNumerically("~", 11, 0.01))
			Expect(spike(10.5)).Should(BeNumerically("~", 14, 0.01))
			Expect(spike(10.75)).Should(BeNumerically("~", 17, 0.01))
			Expect(spike(11)).Should(BeNumerically("~", 20, 0.01))
			Expect(spike(11.5)).Should(BeNumerically("~", 20, 0.01))
			Expect(spike(12)).Should(BeNumerically("~", 20, 0.01))
			Expect(spike(13)).Should(BeNumerically("~", 20, 0.01))
			Expect(spike(13.25)).Should(BeNumerically("~", 17, 0.01))
			Expect(spike(13.5)).Should(BeNumerically("~", 14, 0.01))
			Expect(spike(13.75)).Should(BeNumerically("~", 11, 0.01))
			Expect(spike(14)).Should(BeNumerically("~", 8, 0.01))

			s2 := Spike(2, 20, 4, 5, 2.5)
			Expect(s2(19.9)).Should(BeNumerically("~", 2, 0.01))
			Expect(s2(20.1)).Should(BeNumerically("~", 2.16, 0.01))
			Expect(s2(20.5)).Should(BeNumerically("~", 2.8, 0.01))
			Expect(s2(20.9)).Should(BeNumerically("~", 3.44, 0.01))
			Expect(s2(21.7)).Should(BeNumerically("~", 4.72, 0.01))
			Expect(s2(22.4)).Should(BeNumerically("~", 5.84, 0.01))
			Expect(s2(22.5)).Should(BeNumerically("~", 6, 0.01))
			Expect(s2(22.6)).Should(BeNumerically("~", 6, 0.01))
			Expect(s2(27.4)).Should(BeNumerically("~", 6, 0.01))
			Expect(s2(27.5)).Should(BeNumerically("~", 6, 0.01))
			Expect(s2(27.6)).Should(BeNumerically("~", 5.84, 0.01))
			Expect(s2(27.9)).Should(BeNumerically("~", 5.36, 0.01))
			Expect(s2(28.2)).Should(BeNumerically("~", 4.88, 0.01))
			Expect(s2(28.9)).Should(BeNumerically("~", 3.76, 0.01))
			Expect(s2(29.3)).Should(BeNumerically("~", 3.12, 0.01))
			Expect(s2(29.5)).Should(BeNumerically("~", 2.8, 0.01))
			Expect(s2(29.9)).Should(BeNumerically("~", 2.16, 0.01))
			Expect(s2(30)).Should(BeNumerically("~", 2, 0.01))
			Expect(s2(30.1)).Should(BeNumerically("~", 2, 0.01))
			Expect(s2(30.4)).Should(BeNumerically("~", 2, 0.01))
		})

		It("AmplitudeFromArray", func() {
			afa := AmplitudeFromArray([]float64{1, 5, 0.5, 0.6, 8}, 1.5)
			Expect(afa(0)).Should(BeNumerically("~", 1, 0.01))
			Expect(afa(1)).Should(BeNumerically("~", 1, 0.01))
			Expect(afa(1.49)).Should(BeNumerically("~", 1, 0.01))

			Expect(afa(1.5)).Should(BeNumerically("~", 5, 0.01))
			Expect(afa(1.51)).Should(BeNumerically("~", 5, 0.01))
			Expect(afa(2.99)).Should(BeNumerically("~", 5, 0.01))

			Expect(afa(3)).Should(BeNumerically("~", 0.5, 0.01))
			Expect(afa(3.01)).Should(BeNumerically("~", 0.5, 0.01))
			Expect(afa(7.49)).Should(BeNumerically("~", 8, 0.01))

			Expect(afa(7.5)).Should(BeNumerically("~", 1, 0.01))
			Expect(afa(7.51)).Should(BeNumerically("~", 1, 0.01))

			Expect(afa(74.99)).Should(BeNumerically("~", 8, 0.01))

			Expect(afa(75)).Should(BeNumerically("~", 1, 0.01))
			Expect(afa(75.01)).Should(BeNumerically("~", 1, 0.01))
		})

		It("RateLimiter on Const", func() {
			limit := 50
			rl := NewRateLimiter("test-const", Const(float64(limit)), 5*time.Second, nil)

			start := time.Now()
			for i := 0; i < limit-1; i++ {
				rl.Wait()
			}
			Expect(time.Since(start)).To(BeNumerically("<", 1*time.Second))

			start2 := time.Now()
			for j := 0; j < 3*limit; j++ {
				rl.Wait()
			}
			Expect(time.Since(start2)).To(BeNumerically(">", 2500*time.Millisecond))
		})

		It("RateLimiter on linear increase -- this is a pretty fragile test", func() {
			rl := NewRateLimiter("test-linear", Linear(1), 100*time.Millisecond, nil)

			for i := 1; i <= 10; i++ {
				time.Sleep(110 * time.Millisecond)
				actual := rl.Limit()
				expected := float64(i) / 10
				Expect(actual).To(BeNumerically("~", expected, 0.1))
				log.Infof("currently at %f, expected %f", actual, expected)
			}
		})
	})

	Describe("AdaptiveRateAdjuster", func() {
		f := (&ErrorFractionThresholdConfig{
			IncreaseRatio:            1.1,
			IncreaseMaxErrorFraction: 0.05,
			DecreaseRatio:            0.5,
			DecreaseMinErrorFraction: 0.1,
			MaxRate:                  10,
			MinRate:                  0.5,
		}).RateAdjuster()

		It("should increase, stay the same, and decrease depending on error rate", func() {
			Expect(f(1, .9, 0)).To(BeNumerically("~", 1.1, 0.01))
			Expect(f(1, .9, 0.06)).To(BeNumerically("~", 1, 0.01))
			Expect(f(1, .9, 0.11)).To(BeNumerically("~", 0.5, 0.01))
		})

		It("should not raise rates when less than 0.8 of the current limit", func() {
			Expect(f(1, .80, 0.01)).To(BeNumerically("~", 1.1, 0.01))
			Expect(f(1, .79, 0.01)).To(BeNumerically("~", 1, 0.01))
		})

		It("should raise rates even when below the min", func() {
			Expect(f(0.25, .24, 0)).To(BeNumerically("~", 0.275, 0.01))
			Expect(f(0.275, .27, 0)).To(BeNumerically("~", 0.3025, 0.01))
		})
	})
}
