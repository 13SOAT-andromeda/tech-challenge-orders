package clients

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/sony/gobreaker"
)

type contextKey string

const (
	ctxRequestID   contextKey = "x-request-id"
	ctxTraceparent contextKey = "traceparent"
)

// WithRequestHeaders injects outbound propagation headers into ctx.
// Call this in the HTTP middleware before passing ctx downstream.
func WithRequestHeaders(ctx context.Context, requestID, traceparent string) context.Context {
	ctx = context.WithValue(ctx, ctxRequestID, requestID)
	return context.WithValue(ctx, ctxTraceparent, traceparent)
}

type baseClient struct {
	http *http.Client
	cb   *gobreaker.CircuitBreaker
}

func newBaseClient(name string, timeout time.Duration) baseClient {
	return baseClient{
		http: &http.Client{Timeout: timeout},
		cb: gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:        name,
			MaxRequests: 1,
			Timeout:     30 * time.Second,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures >= 5
			},
		}),
	}
}

// execute wraps makeReq with retry (3×, exponential, 5xx/timeout only) and circuit breaker.
func (b *baseClient) execute(ctx context.Context, makeReq func() (*http.Request, error)) (*http.Response, error) {
	raw, err := b.cb.Execute(func() (any, error) {
		return backoff.Retry(ctx, func() (*http.Response, error) {
			req, err := makeReq()
			if err != nil {
				return nil, backoff.Permanent(err)
			}
			if id, ok := ctx.Value(ctxRequestID).(string); ok && id != "" {
				req.Header.Set("X-Request-ID", id)
			}
			if tp, ok := ctx.Value(ctxTraceparent).(string); ok && tp != "" {
				req.Header.Set("traceparent", tp)
			}
			resp, err := b.http.Do(req)
			if err != nil {
				return nil, err // network/timeout — retryable
			}
			if resp.StatusCode >= 500 {
				resp.Body.Close()
				return nil, fmt.Errorf("upstream %d", resp.StatusCode) // retryable
			}
			if resp.StatusCode >= 400 {
				resp.Body.Close()
				return nil, backoff.Permanent(fmt.Errorf("client error %d", resp.StatusCode))
			}
			return resp, nil
		}, backoff.WithMaxTries(3))
	})
	if err != nil {
		return nil, err
	}
	return raw.(*http.Response), nil
}
