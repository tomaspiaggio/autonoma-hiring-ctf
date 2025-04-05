package steps

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Step1 is the first challenge step
type Step1 struct {
	BaseStep
	answer         string
	errorMsg       string
	guesses        []string
	currentGuess   string
	currentRow     int
	maxGuesses     int
	correctStyle   lipgloss.Style
	partialStyle   lipgloss.Style
	incorrectStyle lipgloss.Style
	emptyStyle     lipgloss.Style
	activeStyle    lipgloss.Style
	highLight      lipgloss.Style
	completed      bool
	argentineWords []string
	hints          map[string]string
}

// NewStep1 creates a new Step1 instance
func NewStep1() *Step1 {
	correctStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#ffffff")).
		Background(lipgloss.Color("#6aaa64")).
		Align(lipgloss.Center).
		Width(5).
		Margin(0, 1, 0, 0)

	partialStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#ffffff")).
		Background(lipgloss.Color("#c9b458")).
		Align(lipgloss.Center).
		Width(5).
		Margin(0, 1, 0, 0)

	incorrectStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#ffffff")).
		Background(lipgloss.Color("#787c7e")).
		Align(lipgloss.Center).
		Width(5).
		Margin(0, 1, 0, 0)

	emptyStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#333333")).
		Width(5).
		Height(1).
		Align(lipgloss.Center).
		Margin(0, 1, 0, 0)

	activeStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#ffffff")).
		Width(5).
		Height(1).
		Align(lipgloss.Center).
		Margin(0, 1, 0, 0)

	highLight := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	argentineWords := []string{
		"asado", "birra", "morfi", "guita", "pibes",
		"chori", "locro", "garca", "mango", "piola",
		"yerba", "flaco", "posta", "amigo", "cheto",
		"mates",
	}

	// Create hints map with Argentinian Spanish phrases
	hints := map[string]string{
		"asado": "Ahora salimos a disfrutar olores",
		"birra": "Bajá inmediatamente Ricardo! Retrasas amigos",
		"morfi": "Mirá, Oscar recién freía ingredientes",
		"guita": "Gastamos últimamente ingresos tantos, amigo",
		"pibes": "Papá invita bebidas esta semana",
		"chori": "Compramos hamburguesas ¡o rica inquisición!",
		"locro": "Llevamos ollas con rica ofrenda",
		"garca": "Gastón ahora reclama comida ajena",
		"mango": "Mamá anduvo negociando ganancias obvias",
		"piola": "Pablo invita otra linda aventura",
		"yerba": "Ya estamos reuniendo bebidas argentinas",
		"flaco": "Fernando llegó a comprar ovejas",
		"posta": "Pablo ordena sequía, tormenta aparece",
		"amigo": "Alguien mencionó interesantes grandes obstáculos",
		"cheto": "Cada hermano evita tomar ómnibus",
		"mates": "Muchos argentinos toman esta semana",
	}

	// Seed random number generator
	randInt := rand.New(rand.NewSource(time.Now().UnixNano()))
	selectedWord := argentineWords[randInt.Intn(len(argentineWords))]

	return &Step1{
		BaseStep:       NewBaseStep("Wordle Challenge"),
		answer:         selectedWord,
		guesses:        make([]string, 6),
		currentGuess:   "",
		currentRow:     0,
		maxGuesses:     6,
		correctStyle:   correctStyle,
		partialStyle:   partialStyle,
		incorrectStyle: incorrectStyle,
		emptyStyle:     emptyStyle,
		activeStyle:    activeStyle,
		argentineWords: argentineWords,
		highLight:      highLight,
		hints:          hints,
		errorMsg:       "",
	}
}

// Init initializes the step
func (s *Step1) Init() tea.Cmd {
	return nil
}

// submitGuess submits the current guess and handles game state changes
func (s *Step1) submitGuess() {
	// Store the guess
	s.guesses[s.currentRow] = s.currentGuess

	// Check if the guess is correct
	if s.currentGuess == s.answer {
		s.completed = true
		s.MarkCompleted()
		return
	}

	s.currentRow++
	s.currentGuess = ""

	// Check if all guesses have been used
	if s.currentRow >= s.maxGuesses {
		s.errorMsg = fmt.Sprintf("You've run out of guesses! The word was: %s", s.answer)
		s.completed = true
	}
}

// Update handles user input
func (s *Step1) Update(msg tea.Msg) (Step, tea.Cmd) {
	if s.completed {
		return s, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "backspace":
			if len(s.currentGuess) > 0 {
				s.currentGuess = s.currentGuess[:len(s.currentGuess)-1]
				s.errorMsg = ""
			}

		default:
			// Only accept letters and limit to 5 characters
			if len(msg.String()) == 1 && msg.String() >= "a" && msg.String() <= "z" && len(s.currentGuess) < 5 {
				s.currentGuess += msg.String()
				s.errorMsg = ""

				// Auto-submit when 5 letters are entered
				if len(s.currentGuess) == 5 {
					s.submitGuess()
				}
			}
			if len(msg.String()) == 1 && msg.String() >= "A" && msg.String() <= "Z" && len(s.currentGuess) < 5 {
				s.currentGuess += strings.ToLower(msg.String())
				s.errorMsg = ""

				// Auto-submit when 5 letters are entered
				if len(s.currentGuess) == 5 {
					s.submitGuess()
				}
			}
		}
	}

	return s, nil
}

// View returns the view for this step
func (s *Step1) View() string {
	var sb strings.Builder

	// Instructions with subtle hint
	sb.WriteString("\n  Guess the 5-letter word. Green = correct letter & position, Yellow = correct letter, wrong position.")
	sb.WriteString("\n  (Hint: Think of unique cultural terms from the land of tango and mate.)")

	// Add the special hint with first letters
	if hint, ok := s.hints[s.answer]; ok {
		sb.WriteString(fmt.Sprintf("\n\n  %s", s.highLight.Render(fmt.Sprintf("Hint: %s", hint))))
	}

	sb.WriteString("\n\n")

	// Wordle grid
	for i := 0; i < s.maxGuesses; i++ {
		var row []string

		if i < s.currentRow {
			// This row has a completed guess
			guess := s.guesses[i]
			for j, ch := range guess {
				letterStr := string(ch)

				// Check if the letter is in the correct position
				if j < len(s.answer) && letterStr == string(s.answer[j]) {
					row = append(row, s.correctStyle.Render(letterStr))
				} else if strings.Contains(s.answer, letterStr) {
					row = append(row, s.partialStyle.Render(letterStr))
				} else {
					row = append(row, s.incorrectStyle.Render(letterStr))
				}
			}
		} else if i == s.currentRow {
			// Current active row
			currentGuessLen := len(s.currentGuess)

			// Fill in the letters typed so far
			for j := 0; j < 5; j++ {
				if j < currentGuessLen {
					row = append(row, s.activeStyle.Render(string(s.currentGuess[j])))
				} else {
					row = append(row, s.activeStyle.Render(" "))
				}
			}
		} else {
			// Empty future rows
			for j := 0; j < 5; j++ {
				row = append(row, s.emptyStyle.Render(" "))
			}
		}

		sb.WriteString("  " + lipgloss.JoinHorizontal(lipgloss.Center, row...) + "\n\n")
	}

	// Error message
	if s.errorMsg != "" {
		sb.WriteString("\n  ")
		sb.WriteString(s.errorMsg)
	}

	// Success message
	if s.IsCompleted() {
		sb.WriteString("\n  ")
		sb.WriteString("¡Felicitaciones! You've guessed the word!")
	}

	// Add current row indicator
	if !s.completed {
		sb.WriteString("\n\n  > Type a 5-letter word (submission is automatic)")
	}

	return sb.String()
}
