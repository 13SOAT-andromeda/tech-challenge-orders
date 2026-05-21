package clients

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/sony/gobreaker"
)

type contextKey string

const (
	ctxRequestID   contextKey = "x-request-id"
	ctxTraceparent contextKey = "traceparent"
	ctxUserID      contextKey = "x-user-id"
	ctxUserEmail   contextKey = "x-user-email"
	ctxUserRole    contextKey = "x-user-role"
)

// WithRequestHeaders injects outbound propagation headers into ctx.
// Call this in the HTTP middleware before passing ctx downstream.
func WithRequestHeaders(ctx context.Context, requestID, traceparent string) context.Context {
	ctx = context.WithValue(ctx, ctxRequestID, requestID)
	return context.WithValue(ctx, ctxTraceparent, traceparent)
}

// WithUserHeaders injects the authenticated user headers into ctx so they are
// forwarded on outbound calls to upstream services.
func WithUserHeaders(ctx context.Context, userID, userEmail, userRole string) context.Context {
	ctx = context.WithValue(ctx, ctxUserID, userID)
	ctx = context.WithValue(ctx, ctxUserEmail, userEmail)
	return context.WithValue(ctx, ctxUserRole, userRole)
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
			if uid, ok := ctx.Value(ctxUserID).(string); ok && uid != "" {
				req.Header.Set("X-User-Id", uid)
			}
			if email, ok := ctx.Value(ctxUserEmail).(string); ok && email != "" {
				req.Header.Set("X-User-Email", email)
			}
			if role, ok := ctx.Value(ctxUserRole).(string); ok && role != "" {
				req.Header.Set("X-User-Role", role)
			}

			url := req.URL.String()
			log.Printf("[http-client] %s %s", req.Method, url)

			resp, err := b.http.Do(req)
			if err != nil {
				log.Printf("[http-client] %s %s: network error: %v", req.Method, url, err)
				return nil, err // network/timeout — retryable
			}
			if resp.StatusCode >= 500 {
				resp.Body.Close()
				log.Printf("[http-client] %s %s: upstream error %d", req.Method, url, resp.StatusCode)
				return nil, fmt.Errorf("upstream %d", resp.StatusCode) // retryable
			}
			if resp.StatusCode >= 400 {
				resp.Body.Close()
				log.Printf("[http-client] %s %s: client error %d", req.Method, url, resp.StatusCode)
				return nil, backoff.Permanent(fmt.Errorf("client error %d", resp.StatusCode))
			}
			return resp, nil
		}, backoff.WithMaxTries(3))
	})
	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
			log.Printf("[http-client] circuit breaker open: %s", b.cb.Name())
		}
		return nil, err
	}
	return raw.(*http.Response), nil
}
