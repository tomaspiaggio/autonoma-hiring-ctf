// main.go
package main

import (
	"fmt"
	"log"
	"net/mail"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/keygen"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/joho/godotenv"
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

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87"))

	// Tab styles
	highlightColor = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	windowStyle    = lipgloss.NewStyle().BorderForeground(highlightColor).Padding(2, 0).Align(lipgloss.Center).Border(lipgloss.NormalBorder()).UnsetBorderTop()
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
	emailError   string
	stepManager  *steps.StepManager
	activeTab    int
	ready        bool
}

func initialModel() model {
	// Initialize email input
	ti := textinput.New()
	ti.Placeholder = "you@example.com"
	ti.Focus()
	ti.Width = 40
	ti.Prompt = promptStyle.Render("Email: ")

	// Create all steps
	allSteps := []steps.Step{}

	// Create step manager
	startTime := time.Now()
	sm := steps.NewStepManager(allSteps, startTime)

	return model{
		keys:         keys,
		help:         help.New(),
		width:        80,
		height:       24,
		startTime:    startTime,
		emailInput:   ti,
		emailEntered: false,
		emailError:   "",
		stepManager:  sm,
		activeTab:    0,
		ready:        false,
	}
}

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

var (
	inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder   = tabBorderWithBottom("┘", " ", "└")
	inactiveTabStyle  = lipgloss.NewStyle().Border(inactiveTabBorder, true).BorderForeground(highlightColor).Padding(0, 1)
	activeTabStyle    = inactiveTabStyle.Border(activeTabBorder, true)
)

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

// isValidEmail checks if the email address has a valid format.
func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case common.TickMsg:
		m.stepManager.UpdateCurrentStep(msg)
		if m.emailEntered {
			m.stepManager.UpdateCurrentStep(msg)
		}
		return m, tickEvery()

	case tea.KeyMsg:
		// Global keybindings
		if key.Matches(msg, m.keys.Quit) {
			return m, tea.Quit
		}

		// If email not entered yet, handle email input
		if !m.emailEntered {
			if msg.String() == "enter" {
				email := m.emailInput.Value()
				m.stepManager.SetEmail(email)
				if email != "" && isValidEmail(email) {
					m.emailEntered = true
					m.emailError = ""
					m.emailInput.Prompt = emailStyle.Render("Email: ")
					m.stepManager.Steps = steps.GenerateSteps(m.stepManager)
					cmds = append(cmds, m.stepManager.Init())
					return m, tea.Batch(cmds...)
				} else {
					m.emailError = "Invalid email format. Please try again."
					m.emailInput.Prompt = errorStyle.Render("Email: ")
					return m, nil
				}
			}

			if m.emailError != "" && msg.Type != tea.KeyEnter {
				m.emailError = ""
				m.emailInput.Prompt = promptStyle.Render("Email: ")
			}

			m.emailInput, cmd = m.emailInput.Update(msg)
			cmds = append(cmds, cmd)
		} else {
			// Pass input to the current step
			stepCmd := m.stepManager.UpdateCurrentStep(msg)
			if stepCmd != nil {
				cmds = append(cmds, stepCmd)
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		m.ready = true
	}

	if m.stepManager.StepFailed {
		return m, tea.Quit
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// Calculate time left
	timeLeft := challengeDuration - time.Since(m.startTime)
	var timeLeftStr string
	if timeLeft <= 0 {
		timeLeftStr = "Time's up!"
		m.stepManager.SetFailedStep("Time's up. You run out of time. You had 25 minutes to complete the challenge.")
	} else {
		minutes := int(timeLeft.Minutes())
		seconds := int(timeLeft.Seconds()) % 60
		timeLeftStr = fmt.Sprintf("Time left: %02d:%02d", minutes, seconds)
	}

	// Before user enters email, show email input prompt
	if !m.emailEntered {
		title := titleStyle.Render("Welcome to Autonoma CTF Challenge")
		timeDisplay := timeStyle.Render(timeLeftStr)

		descriptionList := []string{
			"The following is a CTF made by @tomaspiaggio, CTO at Autonoma.",
			"Anyone can participate, but if you get to the end, you are eligible for a job at Autonoma.",
			"The CTF is a series of challenges that you can solve by coding, reading, and thinking.",
			"You can tweet 'ssh ctf.autonoma.app' if you liked this.",
		}

		description := emailStyle.Render(strings.Join(descriptionList, "\n "))

		emailView := m.emailInput.View()
		if m.emailError != "" {
			emailView += "\n  " + errorStyle.Render(m.emailError)
		}

		s := fmt.Sprintf("\n\n  %s\t%s\n\n %s\n\n %s\n %s\n\n  %s\n\n",
			title,
			timeDisplay,
			description,
			promptStyle.Render("If you get in, you get hired."),
			helpStyle.Render("Note that we'll use that email to contact you."),
			emailView,
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

		return s
	}

	// After email entered, show the tabbed UI
	doc := strings.Builder{}

	// Header with title and time
	title := titleStyle.Render("Autonoma CTF Challenge")
	timeDisplay := timeStyle.Render(timeLeftStr)
	emailDisplay := emailStyle.Render("Email: " + m.emailInput.Value())

	header := fmt.Sprintf("\n  %s\t%s\t%s\n",
		title,
		emailDisplay,
		timeDisplay,
	)
	doc.WriteString(header)

	// Create tabs
	var renderedTabs []string
	var tabNames []string

	// Get current progress to know which steps are unlocked
	completedSteps := m.stepManager.GetCompletedSteps()

	for i, step := range m.stepManager.Steps {
		// Show step name or locked indicator
		tabName := step.Title()
		if i > completedSteps {
			tabName = "****"
		}
		tabNames = append(tabNames, tabName)
	}

	for i, t := range tabNames {
		var style lipgloss.Style

		// Only allow selecting unlocked tabs
		m.activeTab = completedSteps
		m.stepManager.SetCurrentStep(m.activeTab)

		isFirst := i == 0
		isLast := i == len(tabNames)-1
		isActive := i == m.activeTab

		if isActive {
			style = activeTabStyle
		} else {
			style = inactiveTabStyle
		}

		// If tab is locked, use a different style
		if i > completedSteps {
			style = style.Foreground(lipgloss.Color("#626262"))
		}

		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst && !isActive {
			border.BottomLeft = "├"
		} else if isLast && isActive {
			border.BottomRight = "│"
		} else if isLast && !isActive {
			border.BottomRight = "┤"
		}
		style = style.Border(border)
		renderedTabs = append(renderedTabs, style.Render(t))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	doc.WriteString("\n")

	// Set current step based on active tab
	// m.stepManager.SetCurrentStep(m.activeTab)

	// Content area
	tabContent := m.stepManager.CurrentStepView()
	doc.WriteString(windowStyle.Width(m.width - 4).Height(m.height - 10).Render(tabContent))

	return doc.String()
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
		tea.WithMouseAllMotion(),
		tea.WithMouseCellMotion(),
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	// Local mode (command line)
	if len(os.Args) > 1 && os.Args[1] == "local" {
		p := tea.NewProgram(
			initialModel(),
			tea.WithAltScreen(),
			tea.WithMouseAllMotion(),
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
