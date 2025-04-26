package steps

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tomaspiaggio/autonoma-hiring-ctf/common"
)

// HasFailedStep informs the user they have already failed today.
type HasFailedStep struct {
	BaseStep
	targetTime     time.Time // Time when the user can try again (midnight)
	ticker         *time.Ticker
	errorStyle     lipgloss.Style
	infoStyle      lipgloss.Style
	countdownStyle lipgloss.Style
}

// NewHasFailedStep creates a new HasFailedStep instance.
func NewHasFailedStep(sm *StepManager) *HasFailedStep {
	now := time.Now()
	// Calculate the next midnight
	midnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	// Add a small buffer to avoid edge cases right at midnight
	targetTime := midnight.Add(1 * time.Second)

	ticker := time.NewTicker(1 * time.Second)

	return &HasFailedStep{
		BaseStep:   NewBaseStep("Try Again Tomorrow", sm),
		targetTime: targetTime,
		ticker:     ticker,
		errorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87")). // Red/Pink
			Bold(true),
		infoStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")), // White
		countdownStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00")). // Yellow
			Bold(true),
	}
}

// Init marks the step as completed. No command needed as ticking is handled globally.
func (s *HasFailedStep) Init() tea.Cmd {
	s.completed = true // Mark as "completed" in the sense that this is the only step
	return nil         // No need to start its own tick
}

// Update handles the countdown timer and key presses.
func (s *HasFailedStep) Update(msg tea.Msg) (Step, tea.Cmd) {
	switch msg.(type) {
	case common.TickMsg:
		// Check if the target time has been reached
		if time.Now().After(s.targetTime) {
			s.ticker.Stop()
			return s, tea.Quit // Quit when countdown time is reached
		}
		// No need to request another tick, main loop handles it
		return s, nil
	case tea.KeyMsg:
		s.ticker.Stop()
		return s, tea.Quit // Quit on any key press
	}
	// Ignore other messages
	return s, nil
}

// View displays the message and countdown.
func (s *HasFailedStep) View() string {
	// Calculate remaining time dynamically
	remaining := time.Until(s.targetTime)
	if remaining < 0 {
		remaining = 0 // Ensure we don't show negative time
	}

	hours := int(remaining.Hours())
	minutes := int(remaining.Minutes()) % 60
	seconds := int(remaining.Seconds()) % 60
	timeLeftStr := fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)

	msg := s.errorStyle.Render("Access Denied!") + "\n\n"
	msg += s.infoStyle.Render("You have already attempted the challenge today and did not succeed.") + "\n"
	msg += s.infoStyle.Render("Please try again tomorrow.") + "\n\n"
	msg += s.infoStyle.Render("Time until next attempt: ") + s.countdownStyle.Render(timeLeftStr) + "\n\n"
	msg += s.infoStyle.Render("Press any key to exit.")

	return "\n" + msg // Add some top padding
}

// IsCompleted indicates this is the final state for this session.
func (s *HasFailedStep) IsCompleted() bool {
	return true
}
