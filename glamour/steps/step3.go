package steps

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FAFAFA")).
				Background(lipgloss.Color("#7D56F4")).
				Padding(0, 1)

	itemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Padding(0, 1)
)

// Step3 is the third challenge step with a multiple choice question
type Step3 struct {
	BaseStep
	choices  []string
	cursor   int
	question string
	answer   int
	errorMsg string
}

// NewStep3 creates a new Step3 instance
func NewStep3() *Step3 {
	return &Step3{
		BaseStep: NewBaseStep("Security Concepts"),
		choices:  []string{"Malware", "SQL Injection", "Cross-Site Scripting (XSS)", "Phishing"},
		cursor:   0,
		question: "Which of the following is NOT a web application vulnerability?",
		answer:   3, // Phishing
		errorMsg: "",
	}
}

// Init initializes the step
func (s *Step3) Init() tea.Cmd {
	return nil
}

// Update handles user input
func (s *Step3) Update(msg tea.Msg) (Step, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if s.cursor > 0 {
				s.cursor--
			}
		case "down", "j":
			if s.cursor < len(s.choices)-1 {
				s.cursor++
			}
		case "enter", " ":
			if s.cursor == s.answer {
				s.MarkCompleted()
			} else {
				s.errorMsg = "Incorrect choice. Try again!"
			}
		}
	}
	return s, nil
}

// View returns the view for this step
func (s *Step3) View() string {
	var sb strings.Builder

	sb.WriteString("\n  ")
	sb.WriteString(s.question)
	sb.WriteString("\n\n")

	for i, choice := range s.choices {
		if i == s.cursor {
			sb.WriteString(fmt.Sprintf("  %s\n", selectedItemStyle.Render("> "+choice)))
		} else {
			sb.WriteString(fmt.Sprintf("  %s\n", itemStyle.Render("  "+choice)))
		}
	}

	if s.errorMsg != "" {
		sb.WriteString("\n  ")
		sb.WriteString(s.errorMsg)
	}

	sb.WriteString("\n\n  (Use arrow keys to navigate, Enter to select)")

	return sb.String()
}
