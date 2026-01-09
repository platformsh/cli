package auth

import (
	"context"
	"net/http"
)

// eventCtxKey is the context key for storing the event name.
type eventCtxKey struct{}

// WithEventName returns a new context that carries the provided event name.
func WithEventName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, eventCtxKey{}, name)
}

// EventNameFromContext retrieves an event name previously stored with WithEventName.
// It returns an empty string if none is set.
func EventNameFromContext(ctx context.Context) string {
	v, _ := ctx.Value(eventCtxKey{}).(string)
	return v
}

// EventTransport wraps an http.RoundTripper to add event tracking headers.
type EventTransport struct {
	// Base is the underlying RoundTripper to use for requests.
	Base http.RoundTripper

	// EventName is the event name to send in the X-CLI-Event header.
	// If empty, no header is added.
	EventName string

	// UserAgent is the User-Agent string to send.
	// If empty, or a User-Agent is already set on the request, no header is added.
	UserAgent string
}

// RoundTrip adds the X-CLI-Event and User-Agent headers to the request.
func (t *EventTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.EventName != "" {
		req.Header.Set("X-CLI-Event", t.EventName)
	}
	if t.UserAgent != "" && req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", t.UserAgent)
	}
	return t.Base.RoundTrip(req)
}
