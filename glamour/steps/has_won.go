package steps

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tomaspiaggio/autonoma-hiring-ctf/common"
)

// HasWonStep informs the user they have already won the challenge.
type HasWonStep struct {
	BaseStep
	targetTime     time.Time // Time when the user can try again (midnight)
	ticker         *time.Ticker
	congratsStyle  lipgloss.Style
	infoStyle      lipgloss.Style
	countdownStyle lipgloss.Style
}

// NewHasWonStep creates a new HasWonStep instance.
func NewHasWonStep(sm *StepManager) *HasWonStep {
	ticker := time.NewTicker(1 * time.Second)

	now := time.Now()
	targetTime := now.Add(10 * time.Second)

	return &HasWonStep{
		BaseStep:     NewBaseStep("Already Victorious!", sm),
		targetTime:   targetTime,
		ticker:       ticker,
		congratsStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#4CAF50")). // Green
			Padding(1, 0),
		infoStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")). // White
			Padding(0, 1),
		countdownStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")). // Grey
			Italic(true),
	}
}

// Init marks the step as completed. No command needed as ticking is handled globally.
func (s *HasWonStep) Init() tea.Cmd {
	s.completed = true // Mark as "completed" in the sense that this is the only step
	return nil         // No need to start its own tick
}

// Update handles the countdown timer and key presses.
func (s *HasWonStep) Update(msg tea.Msg) (Step, tea.Cmd) {
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

// View displays the congratulatory message.
func (s *HasWonStep) View() string {
	remaining := time.Until(s.targetTime)
	if remaining < 0 {
		remaining = 0 // Ensure we don't show negative time
	}
	countdown := int(remaining.Seconds())
	msg := s.congratsStyle.Render("🎉 Congratulations Again! 🎉") + "\n\n"
	msg += s.infoStyle.Render("Our records show you have already successfully completed the CTF challenge.") + "\n"
	msg += s.infoStyle.Render("There's no need to attempt it again. We hope to speak with you soon!") + "\n\n"
	msg += s.countdownStyle.Render(fmt.Sprintf("Exiting in %d seconds...", countdown)) + "\n"
	msg += s.infoStyle.Render("(Press any key to exit now)")

	return "\n" + msg // Add some top padding
}

// IsCompleted indicates this is the final state for this session.
func (s *HasWonStep) IsCompleted() bool {
	return true
}
