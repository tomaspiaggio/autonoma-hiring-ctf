// main.go
package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/keygen"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/tomaspiaggio/autonoma-hiring-ctf/common"
	"github.com/tomaspiaggio/autonoma-hiring-ctf/glamour/steps"
)

const (
	host              = "localhost"
	port              = 2222
	challengeDuration = 25 * time.Minute
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	stepTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFA500")).
			Bold(true)

	promptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	timeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87")).
			Bold(true)

	emailStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	// Styles for the viewport
	borderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#626262")).
			Padding(1)
)

// Define key bindings
type keyMap struct {
	Quit key.Binding
	Help key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Help, k.Quit},
	}
}

var keys = keyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q/ctrl+c", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
}

// Main TUI model
type model struct {
	keys         keyMap
	help         help.Model
	width        int
	height       int
	startTime    time.Time
	emailInput   textinput.Model
	emailEntered bool
	stepManager  *steps.StepManager
	viewport     viewport.Model
	ready        bool
}

func initialModel() model {
	// Initialize email input
	ti := textinput.New()
	ti.Placeholder = "you@example.com"
	ti.Focus()
	ti.Width = 40
	ti.Prompt = "Email: "

	// Create all steps
	allSteps := []steps.Step{}

	// Create step manager
	sm := steps.NewStepManager(allSteps)

	// Create viewport (will be initialized properly in tea.WindowSizeMsg)
	vp := viewport.New(80, 24)
	vp.Style = borderStyle

	return model{
		keys:         keys,
		help:         help.New(),
		width:        80,
		height:       24,
		startTime:    time.Now(),
		emailInput:   ti,
		emailEntered: false,
		stepManager:  sm,
		viewport:     vp,
		ready:        false,
	}
}

func tickEvery() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return common.TickMsg(t)
	})
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tickEvery(),
		textinput.Blink,
		m.stepManager.Init(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case common.TickMsg:
		m.stepManager.UpdateCurrentStep(msg)

		// when i receive a tick, i'll schedule the next one
		return m, tickEvery()

	case tea.KeyMsg:
		// Global keybindings
		if key.Matches(msg, m.keys.Quit) {
			return m, tea.Quit
		}

		// If email not entered yet, handle email input
		if !m.emailEntered {
			if msg.String() == "enter" && m.emailInput.Value() != "" {
				m.emailEntered = true
				m.viewport.GotoTop()
				m.stepManager.Steps = steps.GenerateSteps()
				return m, nil
			}

			m.emailInput, cmd = m.emailInput.Update(msg)
			cmds = append(cmds, cmd)
		} else {
			// Pass input to the current step
			stepCmd := m.stepManager.UpdateCurrentStep(msg)
			if stepCmd != nil {
				cmds = append(cmds, stepCmd)
			}

			// If it's a viewport navigational key, handle it
			switch msg.String() {
			case "up", "k", "ctrl+u", "pgup":
				m.viewport.LineUp(1)
			case "down", "j", "ctrl+d", "pgdown":
				m.viewport.LineDown(1)
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width

		if !m.ready {
			// Initialize viewport now that we know the terminal size
			headerHeight := 6 // title + time + step title
			footerHeight := 2 // help text
			verticalMargins := 4

			m.viewport = viewport.New(
				msg.Width-2, // account for borders
				msg.Height-headerHeight-footerHeight-verticalMargins,
			)
			m.viewport.Style = borderStyle
			m.viewport.SetContent(m.stepManager.CurrentStepView())
			m.ready = true
		} else {
			m.viewport.Width = msg.Width - 4
			m.viewport.Height = msg.Height - 12
		}
	}

	// Update viewport
	if m.ready && m.emailEntered {
		m.viewport.SetContent(m.stepManager.CurrentStepView())
		var viewportCmd tea.Cmd
		m.viewport, viewportCmd = m.viewport.Update(msg)
		cmds = append(cmds, viewportCmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	var s string

	// Calculate time left
	timeLeft := challengeDuration - time.Since(m.startTime)
	var timeLeftStr string
	if timeLeft <= 0 {
		timeLeftStr = "Time's up!"
	} else {
		minutes := int(timeLeft.Minutes())
		seconds := int(timeLeft.Seconds()) % 60
		timeLeftStr = fmt.Sprintf("Time left: %02d:%02d", minutes, seconds)
	}

	// Before user enters email, show email input prompt
	if !m.emailEntered {
		title := titleStyle.Render("Welcome to Autonoma CTF Challenge")
		tweet := titleStyle.Render("Tweet 'ssh ctf.autonoma.app' to get 10 extra minutes")
		timeDisplay := timeStyle.Render(timeLeftStr)

		s = fmt.Sprintf("\n\n  %s\t%s\n\n  %s\n\n  %s\n\n  %s\n\n",
			title,
			timeDisplay,
			tweet,
			promptStyle.Render("If you get in, you get hired"),
			m.emailInput.View(),
		)

		// Add challenge rules
		rules := []string{
			"IMPORTANT:",
			"- You can only do this test once a day",
			"- If you exit or run out of time, you're done",
			"- Challenges become more difficult as you go along",
			"- Some challenges are time based and require extra concentration",
			"",
			"Good luck!",
		}

		rulesText := helpStyle.Render("  " + strings.Join(rules, "\n  "))
		s += "\n" + rulesText + "\n"

		// Add help at the bottom
		help := helpStyle.Render("  (Enter to submit, Esc to quit)")
		s += "\n" + help

		return s
	}

	// After email entered, show the main UI with current step
	title := titleStyle.Render("Autonoma CTF Challenge")
	timeDisplay := timeStyle.Render(timeLeftStr)
	stepTitle := stepTitleStyle.Render(m.stepManager.CurrentStepTitle())
	emailDisplay := emailStyle.Render("Email: " + m.emailInput.Value())

	// Header
	s = fmt.Sprintf("\n  %s\t%s\t%s\n\n  %s\n\n",
		title,
		emailDisplay,
		timeDisplay,
		stepTitle,
	)

	// Main content area with viewport
	s += m.viewport.View() + "\n\n"

	// Help footer
	helpView := helpStyle.Render(fmt.Sprintf("  %s • %s",
		"↑/↓: Navigate",
		"q: Quit",
	))
	s += helpView

	return s
}

// teaHandler creates a new bubbletea program for each ssh session
func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, active := s.Pty()
	if !active {
		fmt.Println("No active terminal, size will be 80x24")
		pty.Window.Width = 80
		pty.Window.Height = 24
	}

	m := initialModel()

	return m, []tea.ProgramOption{
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	}
}

func main() {
	// Local mode (command line)
	if len(os.Args) > 1 && os.Args[1] == "local" {
		p := tea.NewProgram(
			initialModel(),
			tea.WithAltScreen(),
			tea.WithMouseCellMotion(),
		)

		if _, err := p.Run(); err != nil {
			log.Fatal("Error running program:", err)
			os.Exit(1)
		}
		return
	}

	// Create ssh directory if it doesn't exist
	os.MkdirAll(".ssh", 0700)

	// Create host key if it doesn't exist
	keyPath := ".ssh/id_ed25519"
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		log.Println("Generating new SSH host key...")

		// Generate a new key pair
		_, err := keygen.New(keyPath, keygen.WithKeyType(keygen.Ed25519))
		if err != nil {
			log.Fatalf("Error generating SSH host key: %v", err)
		}

		log.Println("SSH host key generated successfully")
	}

	// Set up ssh server
	s, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf("%s:%d", host, port)),
		wish.WithHostKeyPath(keyPath),
		wish.WithMiddleware(
			bm.Middleware(teaHandler),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Fatalln(err)
	}

	// Start ssh server
	fmt.Printf("Starting Autonoma CTF challenge SSH server on %s:%d...\n", host, port)
	fmt.Println("Connect with: ssh localhost -p 2222")
	fmt.Println("Or run in local mode: go run main.go local")
	log.Fatalln(s.ListenAndServe())
}
