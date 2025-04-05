package steps

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	codeStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#04B575")).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#626262")).
		Padding(1)
)

// Step4 is the fourth challenge step - fix the code
type Step4 struct {
	BaseStep
	textarea textarea.Model
	question string
	solution string
	errorMsg string
	code     string
}

// NewStep4 creates a new Step4 instance
func NewStep4() *Step4 {
	// Broken code with a bug
	brokenCode := `func reverseString(s string) string {
    r := []rune(s)
    for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
        // Something wrong here
        r[i] = r[j]
    }
    return string(r)
}`

	// Create a textarea for the code
	ta := textarea.New()
	ta.SetValue(brokenCode)
	ta.Focus()
	ta.ShowLineNumbers = true
	ta.Placeholder = "Fix the code here"

	return &Step4{
		BaseStep: NewBaseStep("Code Debugging"),
		textarea: ta,
		question: "Fix the Go function that attempts to reverse a string.\nThe current implementation has a bug.",
		solution: "r[i], r[j] = r[j], r[i]",
		errorMsg: "",
		code:     brokenCode,
	}
}

// Init initializes the step
func (s *Step4) Init() tea.Cmd {
	return textarea.Blink
}

// Update handles user input
func (s *Step4) Update(msg tea.Msg) (Step, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+s" || msg.String() == "ctrl+d" {
			// Check the solution
			code := s.textarea.Value()
			if strings.Contains(code, s.solution) {
				s.MarkCompleted()
				s.errorMsg = "Well done! The code is fixed."
				return s, nil
			} else {
				s.errorMsg = "The code is still not correct. Look at how you swap the characters."
				return s, nil
			}
		}
	}

	var cmd tea.Cmd
	s.textarea, cmd = s.textarea.Update(msg)
	return s, cmd
}

// View returns the view for this step
func (s *Step4) View() string {
	var sb strings.Builder

	sb.WriteString("\n  ")
	sb.WriteString(s.question)
	sb.WriteString("\n\n")

	// Show the textarea with the code
	codeView := s.textarea.View()
	sb.WriteString("  ")
	sb.WriteString(codeStyle.Render(codeView))

	if s.errorMsg != "" {
		sb.WriteString("\n\n  ")
		sb.WriteString(s.errorMsg)
	}

	sb.WriteString("\n\n  Press Ctrl+S or Ctrl+D to submit your solution")

	return sb.String()
}
