// SPDX-License-Identifier: Apache-2.0

/*
   Package backoff contains some basic implementations and a selector by strategy name
*/
package backoff

import (
	"math/rand"
	"strings"
	"time"
)

// GetByName returns the WaitBeforeRetry function implementing the strategy
func GetByName(strategy string) TimeToWaitBeforeRetry {
	switch strings.ToLower(strategy) {
	case "linear":
		return LinearBackoff
	case "linear-jitter":
		return LinearJitterBackoff
	case "exponential":
		return ExponentialBackoff
	case "exponential-jitter":
		return ExponentialJitterBackoff
	}
	return DefaultBackoff
}

// TimeToWaitBeforeRetry returns the duration to wait before retrying for the
// given time
type TimeToWaitBeforeRetry func(int) time.Duration

// DefaultBackoffDuration is the duration returned by the DefaultBackoff
var DefaultBackoffDuration = time.Second

// DefaultBackoff always returns DefaultBackoffDuration
func DefaultBackoff(_ int) time.Duration {
	return DefaultBackoffDuration
}

// ExponentialBackoff returns ever increasing backoffs by a power of 2
func ExponentialBackoff(i int) time.Duration {
	return time.Duration(1<<uint(i)) * time.Second
}

// ExponentialJitterBackoff returns ever increasing backoffs by a power of 2
// with +/- 0-33% to prevent sychronized reuqests.
func ExponentialJitterBackoff(i int) time.Duration {
	return jitter(int(1 << uint(i)))
}

// LinearBackoff returns increasing durations, each a second longer than the last
func LinearBackoff(i int) time.Duration {
	return time.Duration(i) * time.Second
}

// LinearJitterBackoff returns increasing durations, each a second longer than the last
// with +/- 0-33% to prevent sychronized reuqests.
func LinearJitterBackoff(i int) time.Duration {
	return jitter(i)
}

var random *rand.Rand

func init() {
	random = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// jitter keeps the +/- 0-33% logic in one place
func jitter(i int) time.Duration {
	ms := i * 1000
	maxJitter := ms/3 + 1
	ms += random.Intn(2*maxJitter) - maxJitter
	if ms <= 0 {
		ms = 1
	}

	return time.Duration(ms) * time.Millisecond
}
