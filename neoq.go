package neoq

import (
	"context"
	"time"

	"github.com/acaloiaro/neoq/backends/memory"
	"github.com/acaloiaro/neoq/config"
	"github.com/acaloiaro/neoq/logging"
	"github.com/acaloiaro/neoq/types"
)

// New creates a new backend instance for job processing.
//
// By default, neoq initializes [memory.Backend] if New() is called without a backend configuration option.
//
// Use [neoq.WithBackend] to initialize different backends.
//
// For available configuration options see [config.ConfigOption].
func New(ctx context.Context, opts ...config.Option) (b types.Backend, err error) {
	c := config.Config{}
	for _, opt := range opts {
		opt(&c)
	}

	if c.BackendInitializer == nil {
		c.BackendInitializer = memory.Backend
	}

	b, err = c.BackendInitializer(ctx, opts...)
	if err != nil {
		return
	}

	return
}

// WithBackend configures neoq to initialize a specific backend for job processing.
//
// Neoq provides two [config.BackendInitializer] that may be used with WithBackend
//   - [pkg/github.com/acaloiaro/neoq/backends/memory.Backend]
//   - [pkg/github.com/acaloiaro/neoq/backends/postgres.Backend]
func WithBackend(initializer config.BackendInitializer) config.Option {
	return func(c *config.Config) {
		c.BackendInitializer = initializer
	}
}

// WithJobCheckInterval configures the duration of time between checking for future jobs
func WithJobCheckInterval(interval time.Duration) config.Option {
	return func(c *config.Config) {
		c.JobCheckInterval = interval
	}
}

// WithLogLevel configures the log level for neoq's default logger. By default, log level is "INFO".
// if SetLogger is used, WithLogLevel has no effect on the set logger
func WithLogLevel(level logging.LogLevel) config.Option {
	return func(c *config.Config) {
		c.LogLevel = level
	}
}
