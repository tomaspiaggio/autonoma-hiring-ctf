package steps

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dop251/goja"
)

// Step5 is the navigation grid challenge
type Step5 struct {
	BaseStep
	textarea textarea.Model
	question string
	errorMsg string
	code     string
	grid     [][]int
}

// NewStep5 creates a new Step5 instance
func NewStep5() *Step5 {
	// Initial JavaScript function template
	jsTemplate := `function hasPath(x, y, grid) {
    // Your implementation here
    // Navigate from (0,0) to (5,5) on the grid
    // You can only move right (R) or down (D)
    // Some cells are blocked (marked as 1)
    // Return the path as an array of moves ("R" or "D")
    // Return undefined if no path is found
}
`

	// Create a textarea for the code
	ta := textarea.New()
	ta.SetValue(jsTemplate)
	ta.Focus()
	ta.ShowLineNumbers = true
	ta.Placeholder = "Write your solution here"
	ta.SetWidth(100) // Increase width significantly
	ta.SetHeight(20) // Increase height significantly

	// Define the grid with obstacles (1 = obstacle, 0 = free)
	grid := [][]int{
		{0, 0, 1, 0, 0, 0},
		{0, 1, 0, 0, 1, 0},
		{0, 0, 0, 1, 0, 0},
		{1, 0, 0, 0, 0, 1},
		{0, 1, 0, 0, 1, 0},
		{0, 0, 0, 0, 0, 0},
	}

	return &Step5{
		BaseStep: NewBaseStep("Grid Navigation Challenge"),
		textarea: ta,
		question: "Navigate from (0,0) to (5,5) on a 6x6 grid.\nYou can only move right (R) or down (D).\nAvoid obstacles (marked as 1).\nImplement the hasPath function.",
		errorMsg: "",
		code:     jsTemplate,
		grid:     grid,
	}
}

// Init initializes the step
func (s *Step5) Init() tea.Cmd {
	return textarea.Blink
}

// Update handles user input
func (s *Step5) Update(msg tea.Msg) (Step, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+s" || msg.String() == "ctrl+d" {
			// Check the solution
			code := s.textarea.Value()
			if s.evaluateCode(code) {
				s.MarkCompleted()
				s.errorMsg = "Congratulations! Your solution successfully navigates the grid."
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
func (s *Step5) evaluateCode(code string) bool {
	vm := goja.New()

	// Execute the user's code
	_, err := vm.RunString(code)
	if err != nil {
		s.errorMsg = fmt.Sprintf("JavaScript error: %v", err)
		return false
	}

	// Get the function
	hasPathFn, ok := goja.AssertFunction(vm.Get("hasPath"))
	if !ok {
		s.errorMsg = "Could not find the hasPath function"
		return false
	}

	// Call the function with our grid
	res, err := hasPathFn(goja.Undefined(), vm.ToValue(0), vm.ToValue(0), vm.ToValue(s.grid))
	if err != nil {
		s.errorMsg = fmt.Sprintf("Error executing your function: %v", err)
		return false
	}

	// Check if result is undefined (no path)
	if res == nil || goja.IsUndefined(res) {
		s.errorMsg = "Your function returned undefined, but there is a valid path through the grid"
		return false
	}

	// Convert the result to a Go slice
	pathObj := res.Export()
	path, ok := pathObj.([]interface{})
	if !ok {
		s.errorMsg = "Your function should return an array of moves (\"R\" or \"D\")"
		return false
	}

	// Validate the path
	return s.validatePath(path)
}

// validatePath checks if the path is valid
func (s *Step5) validatePath(path []interface{}) bool {
	x, y := 0, 0

	// Convert path to string array
	moves := make([]string, len(path))
	for i, move := range path {
		moveStr, ok := move.(string)
		if !ok {
			s.errorMsg = "Path should contain only string values (\"R\" or \"D\")"
			return false
		}

		if moveStr != "R" && moveStr != "D" {
			s.errorMsg = "Path should contain only \"R\" or \"D\" moves"
			return false
		}

		moves[i] = moveStr
	}

	// Follow the path
	for _, move := range moves {
		switch move {
		case "R":
			x++
		case "D":
			y++
		}

		// Check bounds
		if x >= 6 || y >= 6 {
			s.errorMsg = "Path goes out of bounds"
			return false
		}

		// Check for obstacles
		if s.grid[y][x] == 1 {
			s.errorMsg = fmt.Sprintf("Path hits an obstacle at position (%d,%d)", x, y)
			return false
		}
	}

	// Check if we reached the destination
	if x == 5 && y == 5 {
		return true
	}

	s.errorMsg = fmt.Sprintf("Path ends at (%d,%d), not the destination (5,5)", x, y)
	return false
}

// View returns the view for this step
func (s *Step5) View() string {
	var sb strings.Builder

	sb.WriteString("\n  ")
	sb.WriteString(s.question)
	sb.WriteString("\n\n  Grid (1 = obstacle, 0 = free path):\n")

	// Display the grid
	for _, row := range s.grid {
		sb.WriteString("  ")
		for _, cell := range row {
			sb.WriteString(fmt.Sprintf("%d ", cell))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n  Your code:\n")
	codeView := s.textarea.View()
	sb.WriteString("  ")

	// Use a different style to avoid redeclaration issue
	codeBoxStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#04B575")).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#626262")).
		Padding(1)

	sb.WriteString(codeBoxStyle.Render(codeView))

	if s.errorMsg != "" {
		sb.WriteString("\n\n  ")
		sb.WriteString(s.errorMsg)
	}

	sb.WriteString("\n\n  Press Ctrl+S or Ctrl+D to submit your solution")

	return sb.String()
}
