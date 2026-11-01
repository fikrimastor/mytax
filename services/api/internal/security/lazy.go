package security

import (
 "context"
 "errors"
 "fmt"
 "sync"
 "time"
)

// ErrVerifierUnavailable means the identity provider could not be reached to
// build the verifier. It is NOT an authentication success: callers must fail
// closed (RequireAccess answers 503).
var ErrVerifierUnavailable = errors.New("token verifier unavailable")

// LazyVerifier builds the real verifier on first use instead of at boot, so
// the API can start before Keycloak is ready (for example when Coolify starts
// both together). Until construction succeeds every request is rejected.
// Failed attempts are rate-limited by minRetry and only one runs at a time.
// Configuration must be validated at startup (ValidateAccessConfig).
type LazyVerifier struct {
 factory  func(context.Context) (TokenVerifier, error)
 timeout  time.Duration
 minRetry time.Duration
 now      func() time.Time
 mu       sync.Mutex
 v        TokenVerifier
 inflight bool
 nextTry  time.Time
}

func NewLazyVerifier(factory func(context.Context) (TokenVerifier, error), timeout, minRetry time.Duration) (*LazyVerifier, error) {
 if factory == nil || timeout <= 0 || minRetry <= 0 {
  return nil, errors.New("factory, timeout and minRetry are required")
 }
 return &LazyVerifier{factory: factory, timeout: timeout, minRetry: minRetry, now: time.Now}, nil
}

func (l *LazyVerifier) get(ctx context.Context) (TokenVerifier, error) {
 l.mu.Lock()
 if l.v != nil {
  v := l.v
  l.mu.Unlock()
  return v, nil
 }
 if l.inflight || l.now().Before(l.nextTry) {
  l.mu.Unlock()
  return nil, ErrVerifierUnavailable
 }
 l.inflight = true
 l.mu.Unlock()

 // A client disconnect must not poison the shared attempt.
 attempt, cancel := context.WithTimeout(context.WithoutCancel(ctx), l.timeout)
 v, err := l.factory(attempt)
 cancel()

 l.mu.Lock()
 defer l.mu.Unlock()
 l.inflight = false
 if err != nil || v == nil {
  l.nextTry = l.now().Add(l.minRetry)
  return nil, fmt.Errorf("%w: %v", ErrVerifierUnavailable, err)
 }
 l.v = v
 return v, nil
}

func (l *LazyVerifier) Verify(ctx context.Context, raw string) (Identity, error) {
 v, err := l.get(ctx)
 if err != nil {
  return Identity{}, err
 }
 return v.Verify(ctx, raw)
}
