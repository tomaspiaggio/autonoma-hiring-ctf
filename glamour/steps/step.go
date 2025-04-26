package steps

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Step represents a challenge step in the CTF
type Step interface {
	// View returns the content to be displayed for this step
	View() string

	// Title returns the title of the step
	Title() string

	// IsCompleted checks if the step is completed
	IsCompleted() bool

	// Update handles user input and updates the step's state
	Update(msg tea.Msg) (Step, tea.Cmd)

	// Init initializes the step and returns any commands
	Init() tea.Cmd
}

// BaseStep contains common properties for all steps
type BaseStep struct {
	title     string
	completed bool
	sm        *StepManager
}

// NewBaseStep creates a new base step with the given title
func NewBaseStep(title string, sm *StepManager) BaseStep {
	return BaseStep{
		title:     title,
		completed: false,
		sm:        sm,
	}
}

// Title returns the step title
func (b BaseStep) Title() string {
	return b.title
}

// IsCompleted checks if the step is completed
func (b BaseStep) IsCompleted() bool {
	return b.completed
}

// MarkCompleted marks the step as completed
func (b *BaseStep) MarkCompleted() {
	b.completed = true
}

// StepManager handles progression between steps
type StepManager struct {
	Steps       []Step
	CurrentStep int
	startTime   time.Time
	StepFailed  bool
}

// NewStepManager creates a new step manager with the given steps
func NewStepManager(steps []Step, startTime time.Time) *StepManager {
	return &StepManager{
		Steps:       steps,
		CurrentStep: 0,
		startTime:   startTime,
		StepFailed:  false,
	}
}

// CurrentStepView returns the view of the current step
func (sm *StepManager) CurrentStepView() string {
	if sm.CurrentStep < len(sm.Steps) {
		return sm.Steps[sm.CurrentStep].View()
	}
	return "Challenge completed!"
}

// UpdateCurrentStep updates the current step with the given message
func (sm *StepManager) UpdateCurrentStep(msg tea.Msg) tea.Cmd {
	if sm.CurrentStep < len(sm.Steps) {
		var cmd tea.Cmd
		sm.Steps[sm.CurrentStep], cmd = sm.Steps[sm.CurrentStep].Update(msg)

		// Check if the current step is completed
		if sm.Steps[sm.CurrentStep].IsCompleted() && sm.CurrentStep < len(sm.Steps)-1 {
			sm.CurrentStep++
			// Initialize the next step
			return tea.Batch(cmd, sm.Steps[sm.CurrentStep].Init())
		}

		return cmd
	}
	return nil
}

func (sm *StepManager) SetCurrentStep(step int) {
	sm.CurrentStep = step
}

func (sm *StepManager) SetFailedStep(failureMsg string) {
	stepReached := sm.CurrentStep
	timeTaken := time.Since(sm.startTime)
	sm.Steps = []Step{
		NewFailedStep(stepReached, timeTaken, failureMsg, sm),
	}
	sm.CurrentStep = 0
	sm.StepFailed = true
}

func (sm *StepManager) GetCompletedSteps() int {
	return sm.CurrentStep
}

// Init initializes the step manager
func (sm *StepManager) Init() tea.Cmd {
	if len(sm.Steps) > 0 {
		return sm.Steps[0].Init()
	}
	return nil
}

func GenerateSteps(sm *StepManager) []Step {
	return []Step{
		NewStep1(sm),
		NewStep2(sm),
		NewStep3(sm),
		NewStep4(sm),
		NewStep5(sm),
		NewEndStep(sm),
	}
}
