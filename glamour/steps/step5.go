package steps

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	successStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#04B575")).
		Bold(true)
)

// Step5 is the fifth and final challenge step
type Step5 struct {
	BaseStep
	input      textinput.Model
	challenge  string
	solution   string
	errorMsg   string
	completed  bool
	successMsg string
}

// NewStep5 creates a new Step5 instance
func NewStep5() *Step5 {
	// Encode a congratulatory message
	decodedMessage := "Y29uZ3JhdHVsYXRpb25zISB5b3UgY29tcGxldGVkIHRoZSBjdGYh"

	input := textinput.New()
	input.Placeholder = "Enter decoded message here"
	input.Focus()
	input.Width = 60

	return &Step5{
		BaseStep:   NewBaseStep("Final Challenge"),
		input:      input,
		challenge:  "Decode this Base64 message: " + decodedMessage,
		solution:   "congratulations! you completed the ctf!",
		errorMsg:   "",
		completed:  false,
		successMsg: "Great job! You've successfully completed all the challenges!",
	}
}

// Init initializes the step
func (s *Step5) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles user input
func (s *Step5) Update(msg tea.Msg) (Step, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			userAnswer := strings.TrimSpace(strings.ToLower(s.input.Value()))
			if userAnswer == s.solution {
				s.MarkCompleted()
				s.completed = true
				return s, nil
			} else {
				s.errorMsg = "That's not the correct decoded message. Try again!"
				return s, nil
			}
		case "tab":
			// Provide a hint
			s.errorMsg = "Hint: Use online tools or command line to decode base64"
			return s, nil
		}
	}

	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)
	return s, cmd
}

// View returns the view for this step
func (s *Step5) View() string {
	var sb strings.Builder

	if s.completed {
		sb.WriteString("\n\n  ")
		sb.WriteString(successStyle.Render(s.successMsg))
		sb.WriteString("\n\n  ")
		sb.WriteString("You can now exit the challenge with 'q' or Ctrl+C")
		return sb.String()
	}

	sb.WriteString("\n  ")
	sb.WriteString(s.challenge)
	sb.WriteString("\n\n  ")
	sb.WriteString(s.input.View())

	if s.errorMsg != "" {
		sb.WriteString("\n\n  ")
		sb.WriteString(s.errorMsg)
	}

	sb.WriteString("\n\n  Press TAB for a hint")

	return sb.String()
}
