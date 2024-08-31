package main

import (
	"bufio"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"io"
	"os"
	"strings"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type model struct {
	lines    []string
	selected map[int]struct{} // which items are selected
	cursor   int
}

func main() {
	stat, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}

	if stat.Mode()&os.ModeNamedPipe == 0 && stat.Size() == 0 {
		fmt.Println("Try piping in some text.")
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)
	var b strings.Builder

	for {
		r, _, err := reader.ReadRune()
		if err != nil && err == io.EOF {
			break
		}
		_, err = b.WriteRune(r)
		if err != nil {
			fmt.Println("Error getting input:", err)
			os.Exit(1)
		}
	}

	model := newModel(strings.TrimSpace(b.String()))

	if _, err := tea.NewProgram(model).Run(); err != nil {
		fmt.Println("Couldn't start program:", err)
		os.Exit(1)
	}

}

func newModel(pipedInput string) (m model) {

	m.lines = strings.Split(pipedInput, "\n")
	m.selected = make(map[int]struct{})

	return
}

func (m model) Init() tea.Cmd {
	return nil
}

// Update called when something happens (input, etc)
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit
			// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.lines)-1 {
				m.cursor++
			}
		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}
	//m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {

	s := "Select the differences (-/+) that you want to carry over to the new file\n"
	s += "------------------------------------------------------------------------\n"
	for i, line := range m.lines {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		// Is this choice selected?
		checked := " " // not selected
		if _, ok := m.selected[i]; ok {
			checked = "x" // selected!
		}
		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, line)

	}
	return s
}
