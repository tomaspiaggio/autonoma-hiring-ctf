package steps

import (
	"fmt"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/tomaspiaggio/autonoma-hiring-ctf/email"
)

// EndStep is the final step shown upon successful completion.
type EndStep struct {
	BaseStep
	jwtToken      string
	generatedKey  string
	calLink       string
	instructions  string
	congratsStyle lipgloss.Style
	infoStyle     lipgloss.Style
	tokenStyle    lipgloss.Style
}

// JWTCustomClaims defines the payload for the JWT token.
type JWTCustomClaims struct {
	GoToThisLink string `json:"goToThisLink"`
	Instructions string `json:"instructions"`
	FollowMe     string `json:"followMe"`
	Key          string `json:"key"`
	jwt.RegisteredClaims
}

// TODO: Move this secret to a more secure location (e.g., environment variable)
var jwtSecretKey = []byte("a_very_secret_key_for_autonoma_ctf_shhh")

// NewEndStep creates a new EndStep instance.
func NewEndStep(sm *StepManager) *EndStep {
	generatedKey := uuid.NewString() // Generate a unique key
	calLink := "https://cal.com/tom-piaggio-autonoma/15min"
	instructions := "In the meeting description, please write the key provided below and briefly share your thoughts on the CTF."
	followMe := "@tomaspiaggio"

	// Create the claims
	claims := JWTCustomClaims{
		calLink,
		instructions,
		followMe,
		generatedKey,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // Token expires in 24 hours
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "AutonomaCTF",
			Subject:   "CandidateCompletion",
			ID:        uuid.NewString(),
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	signedToken, err := token.SignedString(jwtSecretKey)
	if err != nil {
		// In a real app, handle this error more gracefully
		signedToken = fmt.Sprintf("Error generating token: %v", err)
	}

	congratsStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#4CAF50")). // Green
		Padding(1, 0)

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Padding(0, 1)

	tokenStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD700")). // Gold
		Padding(1, 0).                         // Reduced padding
		Bold(true)                             // Make it stand out without borders

	return &EndStep{
		BaseStep:      NewBaseStep("Challenge Completed!", sm),
		jwtToken:      signedToken,
		generatedKey:  generatedKey,
		calLink:       calLink,
		instructions:  instructions,
		congratsStyle: congratsStyle,
		infoStyle:     infoStyle,
		tokenStyle:    tokenStyle,
	}
}

// Init initializes the step.
func (s *EndStep) Init() tea.Cmd {
	s.MarkCompleted()
	return nil
}

// Update handles messages for the end step.
func (s *EndStep) Update(msg tea.Msg) (Step, tea.Cmd) {
	switch msg.(type) {
	case tea.MouseMsg:
		// Handle mouse events if needed in the future
		return s, nil
	}

	if !s.sm.EmailSent {
		s.sm.EmailSent = true
		go func() {
			_, err := email.SendEndEmail(s.sm.Email, "Tom Piaggio", "tom@autonoma.app", s.jwtToken)
			if err != nil {
				log.Printf("Error sending end email for %s: %v\n", s.sm.Email, err)
			}
		}()
		go func() {
			_, err := s.sm.db.CreateAttempt(s.sm.Email, false, map[string]interface{}{
				"step": s.sm.CurrentStep,
				"time": time.Since(s.sm.startTime),
				"msg":  s.sm.FailureMsg,
			})
			if err != nil {
				log.Printf("Error creating attempt for %s: %v\n", s.sm.Email, err)
			}
			time.Sleep(5 * time.Second)
			s.sm.StepFailed = true
		}()
	}

	return s, nil
}

// View renders the final success message, JWT token, and instructions.
func (s *EndStep) View() string {
	return fmt.Sprintf(
		"%s\n\n%s\n\n%s\n\n%s",
		s.congratsStyle.Render("🎉 Congratulations! You've completed almost all of the challenges! 🎉"),
		s.infoStyle.Render("Check your inbox for the last challenge."),
		s.infoStyle.Render("If you overcome this challenge, you'll be able to add a calendar invite to meet with Tom"),
		s.infoStyle.Render("Click any key to exit. This message won't be shown again."),
	)
}

// Ensure EndStep implements the Step interface
var _ Step = (*EndStep)(nil)

// Helper to satisfy the interface, specific implementations are in BaseStep or overridden
func (s *EndStep) Title() string     { return s.BaseStep.Title() }
func (s *EndStep) IsCompleted() bool { return s.BaseStep.IsCompleted() }
func (s *EndStep) MarkCompleted()    { s.BaseStep.MarkCompleted() }
