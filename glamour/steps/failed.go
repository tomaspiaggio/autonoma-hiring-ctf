package steps

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tomaspiaggio/autonoma-hiring-ctf/common"
)

// FailedStep indicates the user has failed the challenge.
type FailedStep struct {
	BaseStep
	stepReached    int           // Which step the user failed on (0-indexed)
	timeTaken      time.Duration // Total time spent
	countdown      int           // Seconds remaining before disconnect
	countdownStart int
	failureMsg     string // Optional specific reason for failure
	errorStyle     lipgloss.Style
	infoStyle      lipgloss.Style
}

// NewFailedStep creates a new FailedStep instance.
func NewFailedStep(stepReached int, timeTaken time.Duration, failureMsg string, sm *StepManager) *FailedStep {
	countdown := 10
	return &FailedStep{
		BaseStep:       NewBaseStep("Challenge Failed", sm),
		stepReached:    stepReached,
		timeTaken:      timeTaken, // Store actual time taken
		countdown:      countdown,
		countdownStart: countdown,
		failureMsg:     failureMsg,
		errorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87")). // Red/Pink
			Bold(true),
		infoStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")), // White
	}
}

// Init starts the countdown timer.
func (s *FailedStep) Init() tea.Cmd {
	s.completed = true // Mark as "completed" in the sense that the challenge run is over
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return common.TickMsg(t)
	})
}

// Update handles the countdown timer.
func (s *FailedStep) Update(msg tea.Msg) (Step, tea.Cmd) {
	switch msg.(type) {
	case common.TickMsg:
		if s.countdown > 0 {
			s.countdown--
			if s.countdown <= 0 {
				// Time's up, quit the application/session
				return s, tea.Quit
			}
			// Request another tick
			return s, tea.Tick(time.Second, func(t time.Time) tea.Msg {
				return common.TickMsg(t)
			})
		}
	}
	// Ignore other messages
	return s, nil
}

// View displays the failure message and stats.
func (s *FailedStep) View() string {
	// Format time taken
	totalMinutes := int(s.timeTaken.Minutes())
	totalSeconds := int(s.timeTaken.Seconds()) % 60
	timeStr := fmt.Sprintf("%02d:%02d", totalMinutes, totalSeconds)

	// Build the message
	msg := s.errorStyle.Render("Challenge Failed!") + "\n\n"
	if s.failureMsg != "" {
		msg += s.infoStyle.Render(s.failureMsg) + "\n\n"
	}
	// Use stepReached + 1 for user-friendly display (1-based index)
	msg += s.infoStyle.Render(fmt.Sprintf("You reached step %d.", s.stepReached+1)) + "\n"
	msg += s.infoStyle.Render(fmt.Sprintf("Total time: %s.", timeStr)) + "\n\n"
	msg += s.infoStyle.Render("Don't give up! Feel free to try again tomorrow.") + "\n\n"
	msg += s.infoStyle.Render("Click any key to exit")

	// Center the message (assuming windowStyle provides centering)
	// We might need to adjust this based on how it's rendered in main.go
	return "\n" + msg // Add some top padding
}

// IsCompleted indicates the challenge run is over.
func (s *FailedStep) IsCompleted() bool {
	return true // This step itself doesn't complete, but it signifies the end state.
}
