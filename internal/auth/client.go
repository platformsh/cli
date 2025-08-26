package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/platformsh/cli/internal/legacy"
	"golang.org/x/oauth2"
)

func NewLegacyCLIClient(ctx context.Context, wrapper *legacy.CLIWrapper) (*http.Client, error) {
	ts, err := NewLegacyCLITokenSource(ctx, wrapper)
	if err != nil {
		return nil, fmt.Errorf("oauth2: create token source: %w", err)
	}

	refresher := ts.(refresher)
	baseRT := http.DefaultTransport
	if rt, ok := TransportFromContext(ctx); ok && rt != nil {
		baseRT = rt
	}
	return &http.Client{
		Transport: &Transport{
			refresher: refresher,
			base: &oauth2.Transport{
				Source: ts,
				Base:   baseRT,
			},
			wrapper: wrapper,
		},
	}, nil
}
