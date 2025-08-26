package commands

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/platformsh/cli/internal/auth"
	"github.com/platformsh/cli/internal/config"
	"golang.org/x/oauth2"
)

// newAPICurlCommand creates the `api:curl` command which performs an authenticated HTTP request
// against the configured API, using OAuth2 tokens from the credentials store and retrying once on 401.
func newAPICurlCommand(cnf *config.Config) *cobra.Command {
	var (
		method              string
		data                string
		jsonBody            string
		includeHeaders      bool
		headOnly            bool
		disableCompression  bool
		enableGlob          bool // accepted for compatibility; no effect
		noRetry401          bool
		failNoOutput        bool
		headerFlags         []string
		userSetFailExplicit bool
	)

	cmd := &cobra.Command{
		Use:   "api:curl [flags] [path]",
		Short: "Run an authenticated cURL request on the Upsun API",
		Args:  cobra.RangeArgs(0, 1),
		PreRun: func(cmd *cobra.Command, _ []string) {
			// Track if user explicitly set --fail to avoid overriding it dynamically below.
			if f := cmd.Flags().Lookup("fail"); f != nil && f.Changed {
				userSetFailExplicit = true
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cfg := config.FromContext(ctx)

			// Determine path/URL.
			var target string
			if len(args) > 0 {
				target = args[0]
			} else {
				target = "/"
			}

			// Build absolute URL if a path was provided.
			if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
				base := strings.TrimRight(cfg.API.BaseURL, "/")
				if !strings.HasPrefix(target, "/") {
					target = "/" + target
				}
				target = base + target
			}

			// Resolve method.
			m := strings.ToUpper(strings.TrimSpace(method))
			if m == "" {
				m = "GET"
			}
			if headOnly {
				m = "HEAD"
			}
			if m == "GET" && (data != "" || jsonBody != "") {
				m = "POST"
			}
			if data != "" && jsonBody != "" {
				return fmt.Errorf("cannot use --data and --json together")
			}

			// Dynamic default for --fail: true unless --no-retry-401, unless user specified explicitly.
			if !userSetFailExplicit {
				failNoOutput = !noRetry401
			}

			// Base transport: optionally disable compression.
			baseRT := http.DefaultTransport
			if t, ok := http.DefaultTransport.(*http.Transport); ok && disableCompression {
				clone := t.Clone()
				clone.DisableCompression = true
				baseRT = clone
			}

			var httpClient *http.Client
			if noRetry401 {
				// Use plain oauth2 transport (no 401 retry logic).
				ts, err := auth.NewTokenSource(ctx)
				if err != nil {
					return err
				}
				httpClient = &http.Client{Transport: &oauth2.Transport{Source: ts, Base: baseRT}}
			} else {
				// Use our retrying transport via NewClient and inject baseRT via context.
				ctxWithRT := auth.WithTransport(ctx, baseRT)
				legacyCLI := makeLegacyCLIWrapper(cfg, cmd.OutOrStdout(), cmd.ErrOrStderr(), cmd.InOrStdin())
				c, err := auth.NewLegacyCLIClient(ctxWithRT, legacyCLI)
				if err != nil {
					return err
				}
				httpClient = c
			}

			// Build request.
			var body io.Reader
			if jsonBody != "" {
				body = strings.NewReader(jsonBody)
			} else if data != "" {
				body = strings.NewReader(data)
			}
			req, err := http.NewRequestWithContext(ctx, m, target, body)
			if err != nil {
				return err
			}

			// Set headers.
			req.Header.Set("User-Agent", cfg.UserAgent())
			if jsonBody != "" {
				req.Header.Set("Content-Type", "application/json")
			} else if data != "" && req.Header.Get("Content-Type") == "" {
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}
			// Apply -H headers.
			for _, h := range headerFlags {
				h = strings.TrimSpace(h)
				if h == "" {
					continue
				}
				// Support "Name: value" and "Name=value" forms.
				var name, value string
				if strings.Contains(h, ":") {
					parts := strings.SplitN(h, ":", 2)
					name = strings.TrimSpace(parts[0])
					value = strings.TrimSpace(parts[1])
				} else if strings.Contains(h, "=") {
					parts := strings.SplitN(h, "=", 2)
					name = strings.TrimSpace(parts[0])
					value = strings.TrimSpace(parts[1])
				} else {
					return fmt.Errorf("invalid header format: %q", h)
				}
				if name == "" {
					return fmt.Errorf("invalid header: empty name in %q", h)
				}
				req.Header.Add(name, value)
			}

			// Execute request.
			resp, err := httpClient.Do(req)
			if err != nil {
				return err
			}
			defer func() { _ = resp.Body.Close() }()

			// Handle -f/--fail behavior.
			if failNoOutput && resp.StatusCode >= 400 {
				return httpStatusError(target, resp)
			}

			// Output.
			out := cmd.OutOrStdout()
			// For HEAD requests, always show headers (like curl -I). For --include, add headers before body.
			if includeHeaders || headOnly || strings.EqualFold(m, "HEAD") {
				// Status line.
				fmt.Fprintf(out, "%s %s\r\n", resp.Proto, resp.Status)
				// Headers.
				for k, vs := range resp.Header {
					for _, v := range vs {
						fmt.Fprintf(out, "%s: %s\r\n", k, v)
					}
				}
				fmt.Fprint(out, "\r\n")
			}

			if !headOnly && !strings.EqualFold(m, "HEAD") {
				if _, err := io.Copy(out, resp.Body); err != nil {
					// Swallow broken pipe errors when piping output.
					if !isBrokenPipe(err) {
						return err
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&method, "request", "X", "", "The request method to use")
	cmd.Flags().StringVarP(&data, "data", "d", "", "Data to send")
	cmd.Flags().StringVar(&jsonBody, "json", "", "JSON data to send")
	cmd.Flags().BoolVarP(&includeHeaders, "include", "i", false, "Include headers in the output")
	cmd.Flags().BoolVarP(&headOnly, "head", "I", false, "Fetch headers only")
	cmd.Flags().BoolVar(&disableCompression, "disable-compression", false, "Do not request compressed responses")
	cmd.Flags().BoolVar(&enableGlob, "enable-glob", false, "Enable curl globbing (no effect)")
	cmd.Flags().BoolVar(&noRetry401, "no-retry-401", false, "Disable automatic retry on 401 errors")
	cmd.Flags().BoolVarP(&failNoOutput, "fail", "f", false, "Fail with no output on an error response")
	cmd.Flags().StringArrayVarP(&headerFlags, "header", "H", nil, "Extra header(s) (multiple values allowed)")

	return cmd
}

// helpers
func cloneHeader(h http.Header) http.Header {
	out := make(http.Header, len(h))
	for k, v := range h {
		out[k] = append([]string(nil), v...)
	}
	return out
}

func readAllToString(r io.Reader) string {
	b, _ := io.ReadAll(r)
	return string(b)
}

func isBrokenPipe(err error) bool {
	if err == nil {
		return false
	}
	// This is a heuristic; on macOS broken pipe often contains this substring.
	return strings.Contains(strings.ToLower(err.Error()), "broken pipe")
}

// httpStatusError renders a minimal error similar to curl -f behavior.
func httpStatusError(u string, resp *http.Response) error {
	// Try to display a concise error with status and URL path.
	parsed, _ := url.Parse(u)
	target := u
	if parsed != nil {
		target = parsed.String()
	}
	// Do not dump body.
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	return fmt.Errorf("server returned HTTP %d for %s", resp.StatusCode, target)
}
