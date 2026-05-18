package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/session"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/spf13/cobra"
)

var TUICmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive TUI (Text User Interface)",
	Long: `Launch an interactive text user interface for managing canvases and widgets.

The TUI provides a visual way to:
  - Browse canvases
  - View widgets on selected canvas
  - Create, edit, and delete resources
  - Navigate with keyboard shortcuts

Keyboard shortcuts:
  - ↑/↓ or j/k:  Navigate lists
  - Enter:       Select/View details
  - c:           Create new canvas/widget
  - d:           Delete selected item
  - r:           Refresh data
  - q or Ctrl+C: Quit`,
	RunE: runTUI,
}

type model struct {
	session        *canvus.Session
	canvases       []canvus.Canvas
	widgets        []canvus.Widget
	selectedCanvas *canvus.Canvas
	cursor         int
	mode           string // "canvas" or "widget"
	err            error
	width          int
	height         int
}

func initialModel(sess *canvus.Session) model {
	return model{
		session:  sess,
		canvases: []canvus.Canvas{},
		widgets:  []canvus.Widget{},
		cursor:   0,
		mode:     "canvas",
	}
}

func (m model) Init() tea.Cmd {
	return m.loadCanvases()
}

func (m model) loadCanvases() tea.Cmd {
	return func() tea.Msg {
		canvases, err := m.session.ListCanvases(context.Background(), nil)
		if err != nil {
			return errMsg{err}
		}
		return canvasesLoadedMsg{canvases}
	}
}

func (m model) loadWidgets(canvasID string) tea.Cmd {
	return func() tea.Msg {
		widgets, err := m.session.ListWidgets(context.Background(), canvasID, nil)
		if err != nil {
			return errMsg{err}
		}
		return widgetsLoadedMsg{widgets}
	}
}

type canvasesLoadedMsg struct {
	canvases []canvus.Canvas
}

type widgetsLoadedMsg struct {
	widgets []canvus.Widget
}

type errMsg struct {
	err error
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.mode == "canvas" {
				if m.cursor < len(m.canvases)-1 {
					m.cursor++
				}
			} else if m.mode == "widget" {
				if m.cursor < len(m.widgets)-1 {
					m.cursor++
				}
			}

		case "enter":
			if m.mode == "canvas" && len(m.canvases) > 0 {
				// Select canvas and load its widgets
				m.selectedCanvas = &m.canvases[m.cursor]
				m.mode = "widget"
				m.cursor = 0
				return m, m.loadWidgets(m.selectedCanvas.ID)
			}

		case "esc", "backspace":
			if m.mode == "widget" {
				m.mode = "canvas"
				m.selectedCanvas = nil
				m.widgets = nil
				m.cursor = 0
			}

		case "r":
			// Refresh data
			if m.mode == "canvas" {
				return m, m.loadCanvases()
			} else if m.mode == "widget" && m.selectedCanvas != nil {
				return m, m.loadWidgets(m.selectedCanvas.ID)
			}
		}

	case canvasesLoadedMsg:
		m.canvases = msg.canvases
		return m, nil

	case widgetsLoadedMsg:
		m.widgets = msg.widgets
		return m, nil

	case errMsg:
		m.err = msg.err
		return m, nil
	}

	return m, nil
}

func (m model) View() string {
	var s strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("12")).
		Padding(0, 1)

	s.WriteString(headerStyle.Render("Canvus CLI - Interactive Mode"))
	s.WriteString("\n\n")

	// Error display
	if m.err != nil {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Bold(true)
		s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		s.WriteString("\n\n")
	}

	// Current context
	if m.selectedCanvas != nil {
		contextStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Italic(true)
		s.WriteString(contextStyle.Render(fmt.Sprintf("Canvas: %s (ID: %s)", m.selectedCanvas.Name, m.selectedCanvas.ID)))
		s.WriteString("\n\n")
	}

	// Main content
	if m.mode == "canvas" {
		s.WriteString(m.renderCanvasList())
	} else if m.mode == "widget" {
		s.WriteString(m.renderWidgetsList())
	}

	// Footer with help
	s.WriteString("\n\n")
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8"))

	if m.mode == "canvas" {
		s.WriteString(helpStyle.Render("↑/↓: navigate | Enter: select canvas | r: refresh | q: quit"))
	} else {
		s.WriteString(helpStyle.Render("↑/↓: navigate | Esc: back to canvases | r: refresh | q: quit"))
	}

	return s.String()
}

func (m model) renderCanvasList() string {
	var s strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("14"))

	s.WriteString(titleStyle.Render(fmt.Sprintf("Canvases (%d)", len(m.canvases))))
	s.WriteString("\n\n")

	if len(m.canvases) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true)
		s.WriteString(emptyStyle.Render("No canvases found. Create one to get started."))
		return s.String()
	}

	for i, canvas := range m.canvases {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		itemStyle := lipgloss.NewStyle()
		if i == m.cursor {
			itemStyle = itemStyle.
				Foreground(lipgloss.Color("12")).
				Bold(true)
		}

		// Truncate name if too long
		name := canvas.Name
		if len(name) > 50 {
			name = name[:47] + "..."
		}

		line := fmt.Sprintf("%s %s", cursor, name)
		s.WriteString(itemStyle.Render(line))
		s.WriteString("\n")
	}

	return s.String()
}

func (m model) renderWidgetsList() string {
	var s strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("14"))

	s.WriteString(titleStyle.Render(fmt.Sprintf("Widgets (%d)", len(m.widgets))))
	s.WriteString("\n\n")

	if len(m.widgets) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true)
		s.WriteString(emptyStyle.Render("No widgets on this canvas. Press 'c' to create one."))
		return s.String()
	}

	for i, widget := range m.widgets {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		itemStyle := lipgloss.NewStyle()
		if i == m.cursor {
			itemStyle = itemStyle.
				Foreground(lipgloss.Color("12")).
				Bold(true)
		}

		// Get widget type and truncate if needed
		widgetType := widget.WidgetType
		widgetID := widget.ID
		if len(widgetID) > 20 {
			widgetID = widgetID[:17] + "..."
		}

		line := fmt.Sprintf("%s [%s] %s", cursor, widgetType, widgetID)
		s.WriteString(itemStyle.Render(line))
		s.WriteString("\n")
	}

	return s.String()
}

func runTUI(cmd *cobra.Command, args []string) error {
	// Get session from context
	sess := session.MustGetSession(cmd.Context())

	// Create and run the TUI program
	p := tea.NewProgram(initialModel(sess))
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	return nil
}
