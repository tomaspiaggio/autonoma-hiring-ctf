package steps

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tomaspiaggio/autonoma-hiring-ctf/common"
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

// Step3 is a math challenge with 10 questions
type Step3 struct {
	BaseStep
	questions     []string
	answers       []int
	userAnswers   []int
	currentQ      int
	choices       []int
	cursor        int
	errorMsg      string
	timerStart    time.Time
	timeRemaining time.Duration
	finished      bool
}

// NewStep3 creates a new Step3 instance
func NewStep3(sm *StepManager) *Step3 {
	return &Step3{
		BaseStep: NewBaseStep("Math Challenge", sm),
		questions: []string{
			"What is 7 + 12?",
			"What is 15 + 4?",
			"What is 8 + 9?",
			"What is 3 + 5 - 2?",
			"What is 12 - 4 + 7?",
			"What is 3 × 5 + 2?",
			"What is 8 + 2 × 6?",
			"What is 4 × 3 - 7?",
			"What is 18 - 6 × 2?",
			"What is 3 × (4 + 2)?",
		},
		answers:       []int{19, 19, 17, 6, 15, 17, 20, 5, 6, 18},
		userAnswers:   make([]int, 10),
		currentQ:      0,
		choices:       []int{},
		cursor:        0,
		errorMsg:      "",
		timeRemaining: time.Minute,
		timerStart:    time.Now(),
	}
}

// Init initializes the step
func (s *Step3) Init() tea.Cmd {
	s.timerStart = time.Now()
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return common.TickMsg(t)
	})
}

// Update handles user input
func (s *Step3) Update(msg tea.Msg) (Step, tea.Cmd) {
	switch msg := msg.(type) {
	case common.TickMsg:
		elapsed := time.Since(s.timerStart)
		s.timeRemaining = time.Minute - elapsed

		if s.timeRemaining <= 0 && !s.finished {
			s.finished = true
			correct := 0
			for i, ans := range s.userAnswers {
				if ans == s.answers[i] {
					correct++
				}
			}
			if correct >= 7 {
				s.MarkCompleted()
			} else {
				s.errorMsg = fmt.Sprintf("Time's up! You got %d out of 10 correct. Need at least 7 to pass.")
				go func() {
					time.Sleep(1 * time.Second)
					s.sm.SetFailedStep(s.errorMsg)
				}()
			}
			return s, nil
		}

		// This forces the view to update every second
		return s, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return common.TickMsg(t)
		})

	case tea.KeyMsg:
		if s.finished {
			return s, nil
		}

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
			s.userAnswers[s.currentQ] = s.choices[s.cursor]
			s.currentQ++
			s.cursor = 0

			// Generate choices for the next question
			if s.currentQ < 10 {
				s.generateChoices()
			} else {
				s.finished = true
				correct := 0
				for i, ans := range s.userAnswers {
					if ans == s.answers[i] {
						correct++
					}
				}
				if correct >= 7 {
					s.MarkCompleted()
				} else {
					s.errorMsg = fmt.Sprintf("You got %d out of 10 correct. Need at least 7 to pass.", correct)
				}
			}
		}
	}

	return s, nil
}

// generateChoices creates 4 possible answers including the correct one
func (s *Step3) generateChoices() {
	correctAnswer := s.answers[s.currentQ]
	s.choices = []int{correctAnswer}

	// Add 3 incorrect options
	for len(s.choices) < 4 {
		// Generate a number within +/-5 of correct answer
		option := correctAnswer + (5 - (len(s.choices) * 2))

		// Ensure option is positive and not a duplicate
		if option > 0 && !contains(s.choices, option) {
			s.choices = append(s.choices, option)
		} else {
			option = correctAnswer - (3 - len(s.choices))
			if option > 0 && !contains(s.choices, option) {
				s.choices = append(s.choices, option)
			}
		}
	}

	// Shuffle the choices
	for i := range s.choices {
		j := i + (s.currentQ % (len(s.choices) - i))
		if j >= len(s.choices) {
			j = i
		}
		s.choices[i], s.choices[j] = s.choices[j], s.choices[i]
	}
}

// contains checks if a slice contains a value
func contains(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// View returns the view for this step
func (s *Step3) View() string {
	var sb strings.Builder

	minutes := int(s.timeRemaining.Seconds()) / 60
	seconds := int(s.timeRemaining.Seconds()) % 60

	sb.WriteString(fmt.Sprintf("\n  Time remaining: %02d:%02d", minutes, seconds))
	sb.WriteString(fmt.Sprintf("\n  Question %d of 10\n\n", s.currentQ+1))

	if s.currentQ < 10 && !s.finished {
		sb.WriteString("  " + s.questions[s.currentQ] + "\n\n")

		if len(s.choices) == 0 {
			s.generateChoices()
		}

		for i, choice := range s.choices {
			if i == s.cursor {
				sb.WriteString(fmt.Sprintf("  %s\n", selectedItemStyle.Render(fmt.Sprintf("> %d", choice))))
			} else {
				sb.WriteString(fmt.Sprintf("  %s\n", itemStyle.Render(fmt.Sprintf("  %d", choice))))
			}
		}

		sb.WriteString("\n  (Use arrow keys to navigate, Enter to select)")
	} else {
		if s.IsCompleted() {
			sb.WriteString("\n  Challenge completed! All questions answered correctly.")
		} else {
			sb.WriteString("\n  " + s.errorMsg)
		}
	}

	return sb.String()
}
