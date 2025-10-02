package tui

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Custom Upsun spinner character sets.
var (
	// UpsunSpinner - half-circle sun, using Braille characters.
	UpsunSpinner = spinner.Spinner{
		Frames: []string{
			`     `,
			`  ⢀  `,
			`  ⣀  `,
			` ⢀⣀⡀ `,
			` ⣀⣀⣀ `,
			`⢀⣀⣤⣀⡀`,
			`⣀⣀⣤⣀⣀`,
			`⣀⣤⣶⣤⣀`,
			`⣤⣾⣿⣷⣤`,
			`⣾⣿⣿⣿⣷`,
			`⣾⣿⣿⣿⣷`,
			`⣤⣾⣿⣷⣤`,
			`⣀⣤⣶⣤⣀`,
			`⣀⣀⣤⣀⣀`,
			`⢀⣀⣤⣀⡀`,
			` ⣀⣀⣀ `,
			` ⢀⣀⡀ `,
			`  ⣀  `,
			`  ⡀  `,
			`     `,
		},
		FPS: time.Second / 10,
	}

	// UpsunViolet is the Upsun brand color (#6046FF).
	UpsunViolet = lipgloss.Color("#6046FF")
)

// Spinner wraps bubbles/spinner to provide Start/Stop semantics.
type Spinner struct {
	spinner spinner.Model
	program *tea.Program
	writer  io.Writer
	Suffix  string
	style   lipgloss.Style
}

// spinnerModel is the Bubble Tea model for the spinner.
type spinnerModel struct {
	spinner  spinner.Model
	suffix   string
	quitting bool
}

func (m *spinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *spinnerModel) View() string {
	if m.quitting {
		return ""
	}
	return m.spinner.View() + m.suffix
}

// New creates a new Spinner with the given spinner type and writer.
func New(s spinner.Spinner, w io.Writer) *Spinner {
	model := spinner.New()
	model.Spinner = s
	return &Spinner{
		spinner: model,
		writer:  w,
		style:   lipgloss.NewStyle(),
	}
}

// NewDefault creates a new Spinner with the default Upsun spinner.
func NewDefault(w io.Writer) *Spinner {
	s := New(UpsunSpinner, w)
	_ = s.Color("violet")
	return s
}

// Color sets the spinner color. Supports standard color names and hex codes.
func (s *Spinner) Color(colorName string) error {
	// Map common color names to lipgloss colors.
	colorMap := map[string]lipgloss.Color{
		"magenta": lipgloss.Color("magenta"),
		"violet":  UpsunViolet,
		"green":   lipgloss.Color("green"),
		"cyan":    lipgloss.Color("cyan"),
		"yellow":  lipgloss.Color("yellow"),
		"red":     lipgloss.Color("red"),
		"blue":    lipgloss.Color("blue"),
		"white":   lipgloss.Color("white"),
	}

	if color, ok := colorMap[colorName]; ok {
		s.style = s.style.Foreground(color)
		s.spinner.Style = s.style
	} else if strings.HasPrefix(colorName, "#") {
		// Support hex colors.
		s.style = s.style.Foreground(lipgloss.Color(colorName))
		s.spinner.Style = s.style
	}

	return nil
}

// Start starts the spinner in a goroutine.
func (s *Spinner) Start() {
	if s.program != nil {
		return // Already running.
	}

	model := &spinnerModel{
		spinner: s.spinner,
		suffix:  s.Suffix,
	}

	s.program = tea.NewProgram(
		model,
		tea.WithOutput(s.writer),
		tea.WithoutSignalHandler(),
	)

	go func() {
		if _, err := s.program.Run(); err != nil {
			fmt.Fprintf(s.writer, "Error running spinner: %v\n", err)
		}
	}()

	// Give the program a moment to start.
	time.Sleep(10 * time.Millisecond)
}

// Stop stops the spinner.
func (s *Spinner) Stop() {
	if s.program != nil {
		s.program.Quit()
		s.program = nil
		// Clear the spinner line.
		fmt.Fprint(s.writer, "\r\033[K")
	}
}
