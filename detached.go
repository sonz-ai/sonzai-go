package sonzai

import (
	"context"
	"log/slog"
	"time"
)

// DefaultDetachedTimeout is the upper bound applied to detached streaming
// calls when DetachOptions.Timeout is zero. AI generations rarely exceed a
// couple of minutes, but the 5-minute ceiling tolerates a slow LLM, a
// retrying upstream, or a long tool-use chain while still guaranteeing
// the call eventually returns instead of leaking a goroutine.
const DefaultDetachedTimeout = 5 * time.Minute

// DetachOptions tunes the *Detached streaming variants.
//
// The zero value is valid: detached calls run with DefaultDetachedTimeout
// and emit a warning via the default slog logger if the parent context
// is cancelled mid-stream.
type DetachOptions struct {
	// Timeout caps the detached call. Zero falls back to
	// DefaultDetachedTimeout. Pass a negative value to disable the
	// timeout entirely (rarely what you want — prefer an explicit cap).
	Timeout time.Duration

	// Logger overrides slog.Default() for the misuse warning.
	Logger *slog.Logger

	// OnParentCancel, if non-nil, is invoked instead of the slog warning
	// when the parent context is cancelled while the detached call is
	// still running. Useful for surfacing the condition to metrics or
	// structured tracing rather than logs alone.
	OnParentCancel func(parentErr error)
}

// detachContext returns a context derived from parent whose cancellation
// is decoupled from parent's, plus a cancel function the caller MUST
// invoke when the underlying network call returns.
//
// While the call is in flight, a watchdog goroutine listens for parent
// cancellation; if parent is cancelled before the call completes it logs
// a warning (or invokes opts.OnParentCancel) so accidental misuse is
// caught during dev without aborting the in-flight request.
//
// Use this when the streaming call must outlive its caller — e.g. inside
// a NATS message handler, a Watermill subscriber, or a short-lived HTTP
// request that returns to the client before the AI generation completes.
func detachContext(parent context.Context, opts DetachOptions) (context.Context, context.CancelFunc) {
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = DefaultDetachedTimeout
	}

	// context.WithoutCancel (Go 1.21+) preserves values (auth, tracing,
	// deadlines threaded via values) while severing the cancellation
	// chain from parent.
	base := context.WithoutCancel(parent)

	var (
		detached context.Context
		cancel   context.CancelFunc
	)
	if timeout > 0 {
		detached, cancel = context.WithTimeout(base, timeout)
	} else {
		detached, cancel = context.WithCancel(base)
	}

	// Done channel for the watchdog to exit cleanly once the call
	// returns (regardless of whether parent was cancelled).
	done := make(chan struct{})

	go func() {
		select {
		case <-parent.Done():
			// Parent cancelled while detached call is still in flight.
			// Surface it but DO NOT cancel the detached context — that
			// is the entire point of this helper.
			select {
			case <-done:
				// Call already returned; nothing to warn about.
				return
			default:
			}
			if opts.OnParentCancel != nil {
				opts.OnParentCancel(parent.Err())
				return
			}
			logger := opts.Logger
			if logger == nil {
				logger = slog.Default()
			}
			logger.Warn(
				"sonzai: parent context cancelled during detached streaming call; "+
					"call continues until completion or detached timeout — "+
					"this usually indicates the wrong helper is being used "+
					"(use the cancellation-honoring variant when the caller's "+
					"context lifetime exceeds the generation)",
				"parent_err", parent.Err(),
				"detached_timeout", timeout,
			)
		case <-done:
			// Call returned before parent was cancelled. No-op.
		}
	}()

	// Wrap cancel so callers calling it once both signals the watchdog
	// to exit and tears down the timeout/cancel chain.
	var closed bool
	wrappedCancel := func() {
		if !closed {
			closed = true
			close(done)
		}
		cancel()
	}

	return detached, wrappedCancel
}
