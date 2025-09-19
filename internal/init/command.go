package init

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/alecthomas/chroma/v2/quick"
	"github.com/fatih/color"
	"github.com/upsun/whatsun/pkg/digest"
	"github.com/upsun/whatsun/pkg/files"
	"golang.org/x/sync/errgroup"
	"golang.org/x/term"

	"github.com/platformsh/cli/internal/config"
	"github.com/platformsh/cli/internal/init/streaming"
)

type Options struct {
	OnlyShowDigest bool
	ExtraContext   string

	HTTPClient     *http.Client
	AIServiceURL   string
	UserAgent      string
	RequestTimeout time.Duration // Defaults to 10 minutes

	OrganizationID string
	ProjectID      string

	IsInteractive bool
	Yes           bool
	IsDebug       bool
	DebugLogFunc  func(fmt string, args ...any)
}

// RunAIConfig runs the command to generate configuration for a project using AI.
func RunAIConfig(
	ctx context.Context,
	cnf *config.Config,
	dg *digest.Digest,
	path string,
	opts *Options,
	stdout, stderr io.Writer,
) error {
	if opts.AIServiceURL == "" {
		return fmt.Errorf("no AI service URL available")
	}
	if opts.HTTPClient == nil {
		opts.HTTPClient = http.DefaultClient
	}

	spinr := defaultSpinner(stderr)
	defer spinr.Stop()

	if !files.IsLocal(path) {
		return fmt.Errorf("only local file paths are supported")
	}

	var (
		configRelativePath = filepath.Join(".upsun", "config.yaml")
		configAbsPath      = filepath.Join(path, configRelativePath)
	)
	if res, err := confirmOverwrite(opts, configAbsPath, stderr); err != nil {
		return err
	} else if !res {
		return nil
	}

	bodyBytes, err := json.Marshal(Input{
		Digest:         dg,
		ExtraContext:   opts.ExtraContext,
		OrganizationID: opts.OrganizationID,
		ProjectID:      opts.ProjectID,
		Debug:          opts.IsDebug,
	})
	if err != nil {
		return err
	}

	u, err := url.Parse(opts.AIServiceURL)
	if err != nil {
		return err
	}
	u.Path = "/ai/generate-configuration"
	urlStr := u.String()

	if opts.DebugLogFunc != nil {
		opts.DebugLogFunc("Sending %d bytes to URL: %s", len(bodyBytes), urlStr)
	}

	printWithSpinner(spinr, defaultMessageColor, "Calling the AI API")

	if opts.RequestTimeout == 0 {
		opts.RequestTimeout = 10 * time.Minute
	}
	reqCtx, cancelReq := context.WithTimeout(ctx, opts.RequestTimeout)
	defer cancelReq()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, urlStr, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", opts.UserAgent)
	resp, err := opts.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusTooManyRequests:
		if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
			return fmt.Errorf("rate limit exceeded: please try again after %s seconds", retryAfter)
		}
		return fmt.Errorf("rate limit exceeded: please try again later")
	case http.StatusBadRequest:
		var errObj struct {
			Error string `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errObj); err == nil && errObj.Error != "" {
			return fmt.Errorf("invalid request: %s", errObj.Error)
		}
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %d from %s", resp.StatusCode, u.String())
	}

	// Process the API response.
	var (
		apiOutput  Output
		handleData = dataHandler(func(data json.RawMessage, key string) error {
			if key == "output" {
				return json.Unmarshal(data, &apiOutput)
			}
			return fmt.Errorf("unexpected data key: %s", key)
		})
		eg, egCtx = errgroup.WithContext(ctx)
		msgChan   = make(chan streaming.Message, 10) // Buffered to prevent deadlock if handleMessage errors.
	)
	eg.Go(func() error {
		for msg := range msgChan {
			if err := handleMessage(&msg, stdout, stderr, spinr, handleData); err != nil {
				return err
			}
		}
		return nil
	})
	eg.Go(func() error {
		defer close(msgChan)
		return streaming.HandleResponse(egCtx, resp, msgChan)
	})
	if err := eg.Wait(); err != nil {
		return err
	}

	fmt.Fprintln(stderr)

	if !apiOutput.Valid || apiOutput.ConfigYAML == "" {
		return fmt.Errorf("no valid configuration received")
	}

	yamlContent := strings.TrimSpace(apiOutput.ConfigYAML)

	// Check if stdout is a terminal and supports color
	if f, ok := stdout.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		if err := quick.Highlight(stdout, yamlContent+"\n", "yaml", "terminal", "bw"); err != nil {
			// Fall back to plain text if highlighting fails
			fmt.Fprintln(stdout, yamlContent)
		}
	} else {
		// Non-interactive or redirected output - use plain text
		fmt.Fprintln(stdout, yamlContent)
	}

	fmt.Fprintln(stderr)
	fmt.Fprintln(stderr,
		color.YellowString("AI can make mistakes."),
		"Please check and modify the configuration to fit your needs.")

	// Determine whether to save the configuration to the file system (if --save was specified).
	// First, offer to save the configuration interactively.
	if opts.IsInteractive {
		confirmSave, err := confirm(stderr, fmt.Sprintf(
			"\nDo you want to save this configuration to %s?", color.GreenString(configRelativePath),
		))
		if err != nil {
			return err
		}
		if !confirmSave {
			return nil
		}
	}

	configYAML := fmt.Sprintf(
		"# %s configuration, generated using AI at: %v\n",
		cnf.Service.Name, time.Now().Format(time.RFC3339),
	) +
		"# AI can make mistakes. Please modify this file to suit your needs.\n" +
		apiOutput.ConfigYAML

	if err := saveConfiguration(opts, configAbsPath, configYAML, stderr); err != nil {
		return err
	}

	fmt.Fprintln(stderr, color.GreenString("\nYou can now deploy your project to %s.", cnf.Service.Name))

	fmt.Fprintf(stderr, "\nTo do so, commit the new configuration file and push:\n\n")
	fmt.Fprintf(stderr, "  git add %s\n", configRelativePath)
	fmt.Fprintf(stderr, "  git commit -m 'Add %s configuration'\n", cnf.Service.Name)
	fmt.Fprintf(stderr, "  %s project:set-remote\n", cnf.Application.Executable)
	fmt.Fprintf(stderr, "  %s push\n", cnf.Application.Executable)

	return nil
}

// confirm asks the user a basic yes/no question.
// TODO refactor this to a shared internal package
func confirm(stderr io.Writer, promptText string) (result bool, err error) {
	var renderer survey.Renderer
	renderer.WithStdio(terminal.Stdio{Err: stderr})
	prompt := &survey.Confirm{
		Renderer: renderer,
		Message:  promptText,
		Default:  true,
	}
	err = survey.AskOne(prompt, &result)
	return
}

func confirmOverwrite(opts *Options, absPath string, stderr io.Writer) (bool, error) {
	if _, err := os.Stat(absPath); err == nil {
		if !opts.IsInteractive && !opts.Yes {
			return false, fmt.Errorf("the configuration file already exists: %s", absPath)
		}

		fmt.Fprintln(stderr, "The configuration file already exists:", color.YellowString(absPath))

		if opts.Yes {
			return true, nil
		}

		if res, err := confirm(stderr, "Are you sure you want to overwrite it?"); err != nil {
			return false, err
		} else if res {
			fmt.Fprintln(stderr)
			return true, nil
		}
		return false, nil
	}
	return true, nil
}

func saveConfiguration(_ *Options, absPath, content string, stderr io.Writer) error {
	if err := ensureDirAndWriteFile(absPath, content); err != nil {
		return err
	}
	fmt.Fprintln(stderr)
	fmt.Fprintf(stderr, "%s Configuration saved to: %s\n",
		color.GreenString("✓"), color.GreenString(absPath))

	return nil
}

func ensureDirAndWriteFile(dest, content string) error {
	parentDir := filepath.Dir(dest)
	if err := os.MkdirAll(parentDir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", parentDir, err)
	}

	if err := os.WriteFile(dest, []byte(content), 0o600); err != nil {
		return fmt.Errorf("failed to write file %s: %w", dest, err)
	}

	return nil
}
