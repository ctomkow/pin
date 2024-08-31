package main

import (
	"bufio"
	"fmt"
	//"github.com/evertras/bubble-table/table"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"io"
	"os"
	"strings"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

//const (
//	columnKeyID     = "id"
//	columnKeyScore  = "score"
//	columnKeyStatus = "status"
//)
//
//var (
//	styleCritical = lipgloss.NewStyle().Foreground(lipgloss.Color("#f00"))
//	styleStable   = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0"))
//	styleGood     = lipgloss.NewStyle().Foreground(lipgloss.Color("#0f0"))
//)

type model struct {
	lines   []string
	columns []table.Column
	rows    []table.Row
	table   table.Model
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

	//tb := table.New(generateColumns(0)).WithRowStyleFunc(rowStyleFunc)
	//tb.NewColumn("status", "selkd", 3)
	//tb.NewColumn("diff_data", "your diff. mark the rows you want to keep", 80)
	m.columns = []table.Column{
		{Title: "", Width: 1},
		{Title: "your diff. mark the rows you want to keep", Width: 80},
	}

	// generate the rows
	for _, line := range m.lines {
		row := table.Row{" ", line}
		m.rows = append(m.rows, row)
	}

	// Initialize the table model
	m.table = table.New(
		table.WithColumns(m.columns),
		table.WithRows(m.rows),
		table.WithFocused(true),
		table.WithHeight(20),
	)

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
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			selectedRow := m.table.Cursor()
			m.markTableRowAsSelected(selectedRow)
			//return m, nil
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	// Send the UI for rendering
	// Highlight the selected row
	return baseStyle.Render(m.table.View()) + "\n  " + m.table.HelpView() + "\n"
}

func (m model) markTableRowAsSelected(rowId int) {

	// make new table with the * beside it
	currentRows := m.table.Rows()
	m.rows = []table.Row{}
	m.table = table.Model{}

	for i, row := range currentRows {
		if i == rowId {
			row = table.Row{"*", row[1]}
		}
		m.rows = append(m.rows, row)
	}

	// Initialize the table model
	m.table = table.New(
		table.WithColumns(m.columns),
		table.WithRows(m.rows),
		table.WithFocused(true),
		table.WithHeight(20),
	)
}

//func rowStyleFunc(input table.RowStyleFuncInput) lipgloss.Style {
//	calculatedStyle := lipgloss.NewStyle()
//
//	switch input.Row.Data[columnKeyStatus] {
//	case "Critical":
//		calculatedStyle = styleCritical.Copy()
//	case "Stable":
//		calculatedStyle = styleStable.Copy()
//	case "Good":
//		calculatedStyle = styleGood.Copy()
//	}
//
//	if input.Index%2 == 0 {
//		calculatedStyle = calculatedStyle.Background(lipgloss.Color("#222"))
//	} else {
//		calculatedStyle = calculatedStyle.Background(lipgloss.Color("#444"))
//	}
//
//	return calculatedStyle
//}
//
//func generateColumns(numCritical int) []table.Column {
//	// Show how many critical there are
//	statusStr := fmt.Sprintf("Score (%d)", numCritical)
//	statusCol := table.NewColumn(columnKeyStatus, statusStr, 10)
//
//	if numCritical > 3 {
//		// This normally applies the critical style to everything in the column,
//		// but in this case we apply a row style which overrides it anyway.
//		statusCol = statusCol.WithStyle(styleCritical)
//	}
//
//	return []table.Column{
//		table.NewColumn(columnKeyID, "ID", 10),
//		table.NewColumn(columnKeyScore, "Score", 8),
//		statusCol,
//	}
//}
