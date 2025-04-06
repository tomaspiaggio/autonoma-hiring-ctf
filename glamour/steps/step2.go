package steps

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Step2 is the password game challenge
type Step2 struct {
	BaseStep
	input       textinput.Model
	password    string
	constraints []constraint
	revealed    int
	errorMsg    string
}

type constraint struct {
	description string
	validate    func(string) (bool, string)
}

// NewStep2 creates a new Step2 instance
func NewStep2() *Step2 {
	input := textinput.New()
	input.Placeholder = "Enter your password"
	input.Focus()
	input.Width = 60

	return &Step2{
		BaseStep:    NewBaseStep("Password Game"),
		input:       input,
		constraints: buildConstraints(),
		revealed:    1,
		errorMsg:    "",
	}
}

func buildConstraints() []constraint {
	return []constraint{
		{
			description: "Password must be at least 8 characters long",
			validate: func(s string) (bool, string) {
				if len(s) < 8 {
					return false, fmt.Sprintf("Too short: %d/8 characters", len(s))
				}
				return true, ""
			},
		},
		{
			description: "Password must contain at least 1 number",
			validate: func(s string) (bool, string) {
				for _, char := range s {
					if unicode.IsDigit(char) {
						return true, ""
					}
				}
				return false, "No numbers found"
			},
		},
		{
			description: "Password must contain at least 1 special character",
			validate: func(s string) (bool, string) {
				specialChars := `!@#$%^&*()-_=+[]{};:'",.<>/?`
				for _, char := range s {
					if strings.ContainsRune(specialChars, char) {
						return true, ""
					}
				}
				return false, "No special characters found"
			},
		},
		{
			description: "The sum of all numbers must be 35",
			validate: func(s string) (bool, string) {
				sum := 0
				for _, char := range s {
					if unicode.IsDigit(char) {
						num, _ := strconv.Atoi(string(char))
						sum += num
					}
				}
				if sum != 35 {
					return false, fmt.Sprintf("Sum is %d, not 35", sum)
				}
				return true, ""
			},
		},
		{
			description: "Password must contain Roman numerals",
			validate: func(s string) (bool, string) {
				romanNumerals := "IVXLCDM"
				found := false
				for _, char := range s {
					if strings.ContainsRune(romanNumerals, char) {
						found = true
						break
					}
				}
				if !found {
					return false, "No Roman numerals found"
				}
				return true, ""
			},
		},
		{
			description: "The Roman numerals must sum to less than 100 and more than 10",
			validate: func(s string) (bool, string) {
				values := map[rune]int{'I': 1, 'V': 5, 'X': 10, 'L': 50, 'C': 100, 'D': 500, 'M': 1000}
				sum := 0
				for _, char := range s {
					if val, ok := values[char]; ok {
						sum += val
					}
				}
				if sum >= 100 || sum <= 10 {
					return false, fmt.Sprintf("Roman numeral sum is %d, must be < 100 and > 10", sum)
				}
				return true, ""
			},
		},
		{
			description: "Password must contain one of Autonoma's founders name in uppercase",
			validate: func(s string) (bool, string) {
				founders := []string{"SIMON", "TOMAS", "NICOLAS", "EUGENIO"}
				for _, name := range founders {
					if strings.Contains(s, name) {
						return true, ""
					}
				}
				return false, "No founder name found"
			},
		},
		{
			description: "Password must contain a programming language",
			validate: func(s string) (bool, string) {
				languages := []string{"PYTHON", "JAVA", "JAVASCRIPT", "C", "CPP", "CSHARP", "PHP", "RUBY", "GO", "SWIFT", "KOTLIN", "RUST", "SCALA", "PERL", "TYPESCRIPT"}
				for _, lang := range languages {
					if strings.Contains(strings.ToUpper(s), lang) {
						return true, ""
					}
				}
				return false, "No programming language found"
			},
		},
		{
			description: "Password must contain the area of a triangle with height 10 and base 4 (lowercase text)",
			validate: func(s string) (bool, string) {
				// Area = (base * height) / 2 = (4 * 10) / 2 = 20
				if strings.Contains(strings.ToLower(s), "twenty") {
					return true, ""
				}
				return false, "Missing the area of the triangle (twenty)"
			},
		},
	}
}

// Init initializes the step
func (s *Step2) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles user input
func (s *Step2) Update(msg tea.Msg) (Step, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return s, nil
		}
	}

	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)

	// Validate on every keystroke
	s.password = s.input.Value()

	// Validate against all revealed constraints
	allValid := true

	for i := 0; i < s.revealed; i++ {
		valid, errMsg := s.constraints[i].validate(s.password)
		if !valid {
			s.errorMsg = errMsg
			allValid = false
			break
		}
	}

	if allValid {
		if s.revealed < len(s.constraints) {
			s.revealed++
			s.errorMsg = "New constraint revealed!"
		} else {
			s.MarkCompleted()
		}
	}

	return s, cmd
}

// View returns the view for this step
func (s *Step2) View() string {
	var sb strings.Builder

	sb.WriteString("\n  Create a password that meets all constraints:\n\n")

	// Show revealed constraints
	for i := 0; i < s.revealed; i++ {
		valid, _ := s.constraints[i].validate(s.password)
		status := "❌"
		if valid {
			status = "✓"
		}
		sb.WriteString(fmt.Sprintf("  %s %s\n", status, s.constraints[i].description))
	}

	sb.WriteString("\n  ")
	sb.WriteString(s.input.View())

	if s.errorMsg != "" {
		sb.WriteString("\n\n  ")
		sb.WriteString(s.errorMsg)
	}

	return sb.String()
}
