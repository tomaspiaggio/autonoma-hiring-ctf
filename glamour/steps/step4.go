package steps

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dop251/goja"
)

var (
	codeStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#04B575")).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#626262")).
		Padding(1)
)

// Step4 is the fourth challenge step - fix async code
type Step4 struct {
	BaseStep
	textarea textarea.Model
	question string
	errorMsg string
	code     string
}

// NewStep4 creates a new Step4 instance
func NewStep4(sm *StepManager) *Step4 {
	// Broken JavaScript code with a promise-related bug
	brokenCode := `async function fetchUserData(users) {
  const userData = [];
  
  for (const user of users) {
    fetchUser(user).then(data => {
      userData.push(data);
    });
  }
  
  return userData;
}

// Mock function (don't modify)
function fetchUser(user) {
  return Promise.resolve({ id: user, name: 'User ' + user });
}`

	// Create a textarea for the code
	ta := textarea.New()
	ta.SetValue(brokenCode)
	ta.Focus()
	ta.ShowLineNumbers = true
	ta.Placeholder = "Fix the code here"
	ta.SetWidth(100)
	ta.SetHeight(20)

	return &Step4{
		BaseStep: NewBaseStep("Async JavaScript Debugging", sm),
		textarea: ta,
		question: "Fix the JavaScript function that fetches user data.\nThe current implementation has a bug where the returned list is always empty.",
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
			if s.evaluateCode(code) {
				s.MarkCompleted()
				s.errorMsg = "Well done! The async code is fixed."
				return s, nil
			}
			return s, nil
		}
	}

	var cmd tea.Cmd
	s.textarea, cmd = s.textarea.Update(msg)
	return s, cmd
}

// evaluateCode runs the user's JavaScript solution and checks if it works
func (s *Step4) evaluateCode(code string) bool {
	vm := goja.New()

	testCode := code + `
	// Test the function
	async function testFetchUserData() {
		const users = [1, 2, 3];
		const result = await fetchUserData(users);
		
		// Check if we have all users
		if (result.length !== 3) {
			throw new Error("Expected 3 users, got " + result.length);
		}
		
		// Check if all users are present
		for (let i = 0; i < users.length; i++) {
			const user = result.find(u => u.id === users[i]);
			if (!user) {
				throw new Error("User " + users[i] + " not found in results");
			}
		}
		
		return true;
	}
	
	// Run the test
	let success = false;
	testFetchUserData().then(result => {
		success = result;
	});
	`

	_, err := vm.RunString(testCode)
	if err != nil {
		s.errorMsg = "JavaScript error: " + err.Error()
		return false
	}

	// Check if the solution contains either Promise.all or await in the loop
	if !strings.Contains(code, "Promise.all") && !strings.Contains(code, "await fetchUser") {
		s.errorMsg = "Your solution should handle promises correctly. Think about how to wait for all promises to resolve."
		return false
	}

	s.errorMsg = "Your solution handles the promises correctly!"
	return true
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
