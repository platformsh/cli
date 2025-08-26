package auth

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/platformsh/cli/internal/legacy"
	"golang.org/x/oauth2"
)

type legacyCLITokenSource struct {
	ctx     context.Context
	cached  *oauth2.Token
	wrapper *legacy.CLIWrapper
	mu      sync.Mutex
}

func (ts *legacyCLITokenSource) getLegacyCLIToken() (*oauth2.Token, error) {
	bt := bytes.NewBuffer(nil)
	ts.wrapper.Stdout = bt
	if err := ts.wrapper.Exec(ts.ctx, "auth:token", "-W"); err != nil {
		return nil, fmt.Errorf("cannot retrieve token: %w", err)
	}

	at, err := unsafeParseJWT(bt.String())

	if err != nil {
		return nil, fmt.Errorf("cannot parse token: %w", err)
	}

	return &oauth2.Token{
		AccessToken: bt.String(),
		TokenType:   "Bearer",
		Expiry:      time.Unix(at.ExpiresAt, 0),
	}, nil
}

func (ts *legacyCLITokenSource) refreshToken() error {
	ts.cached = nil
	ts.wrapper.Stdout = io.Discard
	if err := ts.wrapper.Exec(ts.ctx, "auth:info", "--refresh"); err != nil {
		return fmt.Errorf("cannot refresh token: %w", err)
	}

	return nil
}

func (ts *legacyCLITokenSource) invalidateToken() error {
	if ts.cached != nil {
		ts.cached.AccessToken = ""
	}

	return nil
}

func (ts *legacyCLITokenSource) Token() (*oauth2.Token, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.cached == nil {
		tok, err := ts.getLegacyCLIToken()
		if err != nil {
			return nil, err
		}
		ts.cached = tok
	}

	if ts.cached != nil && ts.cached.Valid() {
		return ts.cached, nil
	}

	if err := ts.refreshToken(); err != nil {
		return nil, err
	}

	tok, err := ts.getLegacyCLIToken()
	if err != nil {
		return nil, err
	}

	ts.cached = tok
	return ts.cached, nil
}

func NewLegacyCLITokenSource(ctx context.Context, wrapper *legacy.CLIWrapper) (oauth2.TokenSource, error) {
	wrapper.ForceColor = true
	wrapper.DisableInteraction = true
	return &legacyCLITokenSource{
		ctx:     ctx,
		wrapper: wrapper,
	}, nil
}
