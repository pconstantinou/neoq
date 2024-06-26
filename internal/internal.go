package internal

import (
	"math"
	"math/rand"
	"regexp"
	"time"
)

type contextKey struct{}

const (
	JobStatusNew       = "new"
	JobStatusProcessed = "processed"
	JobStatusFailed    = "failed"
)

var JobCtxVarKey contextKey

// CalculateBackoff calculates the number of seconds to back off before the next retry
// this formula is unabashedly taken from Sidekiq because it is good.
func CalculateBackoff(retryCount int, runAfter time.Time) time.Time {
	now := time.Now().UTC()
	if runAfter.Before(now) {
		runAfter = now
	}
	const backoffExponent = 4
	const maxInt = 30
	p := int(math.Round(math.Pow(float64(retryCount), backoffExponent)))
	return runAfter.Add(time.Duration(p+15+RandInt(maxInt)*retryCount+1) * time.Second)
}

// RandInt returns a random integer up to max
func RandInt(max int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano())) // nolint: gosec
	return r.Intn(max)
}

// StripNonAlphanum strips nonalphanumeric/non underscore characters from a string and returns a new one
func StripNonAlphanum(s string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	return re.ReplaceAllString(s, "")
}
