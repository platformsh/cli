package auth

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"

	"github.com/platformsh/cli/internal/legacy"
)

type refresher interface {
	refreshToken() error
	invalidateToken() error
}

// Transport is an HTTP RoundTripper similar to golang.org/x/oauth2.Transport.
// It injects Authorization headers using a savingSource and, on a 401 response,
// clears the cached token and retries the request once.
type Transport struct {
	// base is the underlying oauth2.Transport that adds the Authorization header.
	base http.RoundTripper

	// refresher is the savingSource used as the TokenSource for base; kept private
	// so we can clear its cached token on 401.
	refresher refresher

	wrapper *legacy.CLIWrapper
	logger  *log.Logger
}

// RoundTrip adds Authorization via the underlying oauth2.Transport. If the
// response is 401 Unauthorized, it clears the cached token and retries once.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Body = wrapReader(req.Body)

	resp, err := t.base.RoundTrip(req)

	// Retry on 401
	if resp != nil && resp.StatusCode == http.StatusUnauthorized {
		_ = t.log("The access token has been refreshed. Retrying request.")
		_ = t.refresher.invalidateToken()
		flushReader(resp.Body)
		resp, err = t.base.RoundTrip(req)
	}

	return resp, err
}

func (t *Transport) log(msg string, args ...any) error {
	if t.logger == nil {
		return nil
	}
	t.logger.Printf(msg, args...)
	return nil
}

// context key for storing a custom RoundTripper.
type transportCtxKey struct{}

// WithTransport returns a new context that carries the provided RoundTripper.
func WithTransport(ctx context.Context, rt http.RoundTripper) context.Context {
	return context.WithValue(ctx, transportCtxKey{}, rt)
}

// TransportFromContext retrieves a RoundTripper previously stored with
// WithTransport. It returns (nil, false) if none is set.
func TransportFromContext(ctx context.Context) (http.RoundTripper, bool) {
	v := ctx.Value(transportCtxKey{})
	if v == nil {
		return nil, false
	}
	rt, ok := v.(http.RoundTripper)
	if !ok || rt == nil {
		return nil, false
	}
	return rt, true
}

func wrapReader(r io.ReadCloser) io.ReadCloser {
	if r == nil {
		return nil
	}
	bodyBytes, _ := io.ReadAll(r)
	_ = r.Close()
	return io.NopCloser(bytes.NewBuffer(bodyBytes))
}

func flushReader(r io.ReadCloser) {
	if r == nil {
		return
	}
	_, _ = io.Copy(io.Discard, r)
	_ = r.Close()
}
