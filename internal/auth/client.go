package auth

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/platformsh/cli/internal/legacy"
)

type LegacyCLIClient struct {
	HTTPClient  *http.Client
	tokenSource oauth2.TokenSource
}

func (c *LegacyCLIClient) EnsureAuthenticated(_ context.Context) error {
	_, err := c.tokenSource.Token()
	return err
}

// NewLegacyCLIClient creates an HTTP client authenticated through the legacy CLI.
// The wrapper argument must be a dedicated wrapper, not used by other callers.
func NewLegacyCLIClient(ctx context.Context, wrapper *legacy.CLIWrapper) (*LegacyCLIClient, error) {
	ts, err := NewLegacyCLITokenSource(ctx, wrapper)
	if err != nil {
		return nil, fmt.Errorf("oauth2: create token source: %w", err)
	}

	refresher, ok := ts.(refresher)
	if !ok {
		return nil, fmt.Errorf("token source does not implement refresher")
	}
	baseRT := http.DefaultTransport
	if rt, ok := TransportFromContext(ctx); ok && rt != nil {
		baseRT = rt
	}

	httpClient := &http.Client{
		Transport: &Transport{
			refresher: refresher,
			base: &oauth2.Transport{
				Source: ts,
				Base:   baseRT,
			},
		},
	}

	return &LegacyCLIClient{
		HTTPClient:  httpClient,
		tokenSource: ts,
	}, nil
}
