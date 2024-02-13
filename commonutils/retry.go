package commonutils

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type RetryBackOffMode int

const (
	LinearBackoff RetryBackOffMode = iota
	ConstantBackoff
)

const (
	minDelay = 100 * time.Millisecond
	maxDelay = 10 * time.Minute
)

// Retry Function Does retry in case of error which is retryable.
// Internally it uses jitter to avoid thundering herd problem.
func Retry(attempts int, sleep time.Duration, f func() error, isRetryable func(err error) bool,
	r RetryBackOffMode) error {
	var err error
	if err := f(); err != nil {
		if !isRetryable(err) {
			return err
		}

		if attempts--; attempts > 0 {
			// Add some randomness to prevent creating a Thundering Herd
			jitter := time.Duration(rand.Int63n(int64(sleep)))
			sleep = sleep + jitter/2
			switch r {
			case ConstantBackoff:
				time.Sleep(sleep)

			case LinearBackoff:
				sleep = sleep * 2
				time.Sleep(sleep)

			default:
				sleep = sleep * 2
				time.Sleep(sleep)
			}
			return Retry(attempts, sleep, f, isRetryable, r)
		}
		return err
	}

	return err
}

var DefaultRetryableErrorFunction = func(err error) bool {
	return err != nil
}
