package main

import (
	"bufio"
	"flag"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"io"
	"io/ioutil"
	"log"
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
	output   string
}

func main() {

	devMode := flag.Bool("dev", false, "Development test mode")
	flag.Parse()

	var inputString string
	if *devMode {
		content, err := ioutil.ReadFile("data/test.diff")
		if err != nil {
			log.Fatalf("Failed to read test file: data/test.diff")
		}
		inputString = string(content)
	} else {
		// read from pipe and do stuff. Normal operation
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
		inputString = strings.TrimSpace(b.String())
	}

	model := newModel(inputString)

	if _, err := tea.NewProgram(model).Run(); err != nil {
		fmt.Println("Couldn't start program:", err)
		os.Exit(1)
	}

}

func newModel(pipedInput string) *model {

	m := &model{
		lines:    strings.Split(pipedInput, "\n"),
		selected: make(map[int]struct{}),
	}

	return m
}

func (m *model) Init() tea.Cmd {
	return nil
}

// Update called when something happens (input, etc)
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		// a case to handle when someone wants to save and quit like in vim
		case "ctrl+s":
			m.transformDiff()
			err := ioutil.WriteFile("diff.pin", []byte(m.output), 0644)
			if err != nil {
				log.Fatalf("Failed to write output file: diff.pin")
			}
			return m, tea.Quit
		}

	}
	//m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *model) View() string {

	s := "------------------------------------------------------------------------\n"
	s += "Select the differences (-/+) that you want to carry over to the new file\n"
	s += "  Navigate with arrows or 'j' and 'k'\n"
	s += "  Press 'space' to select/deselect\n"
	s += "  Press 'ctrl+s' to save and quit\n"
	s += "  Apply diff.pin using patch. ie. `patch new_file < diff.pin`\n"
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

func (m *model) transformDiff() {
	var b strings.Builder

	var inHunk bool

	// Process the incoming diff
	for i, line := range m.lines {
		// Handle file headers (e.g., --- and +++)
		if strings.HasPrefix(line, "--- ") || strings.HasPrefix(line, "+++ ") {
			// Write the header lines as they are
			b.WriteString(line + "\n")
			continue
		}

		// Handle hunk headers (e.g., @@ -1,9 +1,7 @@)
		if strings.HasPrefix(line, "@@") {
			// Swap the ranges (keep `-` and `+` but swap the values)
			headerParts := strings.Fields(line)
			if len(headerParts) >= 3 {
				oldRange := headerParts[1] // -1,9
				newRange := headerParts[2] // +1,7

				// Swap the ranges: `-` first and `+` second, but swap their counts
				b.WriteString(fmt.Sprintf("@@ -%s +%s @@\n", newRange[1:], oldRange[1:]))
			}
			inHunk = true
			continue
		}

		// Process `-` lines (old file deletions)
		if strings.HasPrefix(line, "-") && inHunk {
			if _, ok := m.selected[i]; ok {
				// If the line is selected (pinned), convert `-` into `+`
				b.WriteString("+" + line[1:] + "\n") // Convert `-` to `+` for pinned lines
			}
			// If the `-` line is not selected, discard it (skip it)
			continue
		}

		// Process `+` lines (new file additions)
		if strings.HasPrefix(line, "+") {
			// Replace `+` with a space and keep the rest of the line unchanged for proper alignment
			b.WriteString(" " + line[1:] + "\n")
			continue
		}

		// Handle context lines (unchanged lines)
		// Write the context lines exactly as they are
		b.WriteString(line + "\n")
	}

	// Store the generated diff
	m.output = b.String()
}
