package renderer

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type Table struct {
	Headers []string
	Rows    [][]string
	Style   *lipgloss.Style
}

func NewTable(headers []string) *Table {
	defaultStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	return &Table{
		Headers: headers,
		Rows:    [][]string{},
		Style:   &defaultStyle,
	}
}

func (t *Table) AddRow(row []string) *Table {
	t.Rows = append(t.Rows, row)
	return t
}

func (t *Table) AddRows(rows [][]string) *Table {
	t.Rows = append(t.Rows, rows...)
	return t
}

func (t *Table) SetStyle(style *lipgloss.Style) *Table {
	t.Style = style
	return t
}

func (t *Table) Render() string {
	if len(t.Headers) == 0 {
		return ""
	}

	colWidths := make([]int, len(t.Headers))
	for i, h := range t.Headers {
		colWidths[i] = len(h)
	}

	for _, row := range t.Rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	var b strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FFFF"))
	separatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#444444"))

	headerRow := make([]string, len(t.Headers))
	for i, h := range t.Headers {
		headerRow[i] = headerStyle.Render(fmt.Sprintf("%-*s", colWidths[i], h))
	}
	b.WriteString(strings.Join(headerRow, "  "))
	b.WriteString("\n")

	separator := make([]string, len(t.Headers))
	for i := range t.Headers {
		separator[i] = separatorStyle.Render(strings.Repeat("─", colWidths[i]))
	}
	b.WriteString(strings.Join(separator, "  "))
	b.WriteString("\n")

	for _, row := range t.Rows {
		rowStrs := make([]string, len(t.Headers))
		for i := range t.Headers {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			rowStrs[i] = t.Style.Render(fmt.Sprintf("%-*s", colWidths[i], cell))
		}
		b.WriteString(strings.Join(rowStrs, "  "))
		b.WriteString("\n")
	}

	return b.String()
}

func RenderPodStatusTable(podData interface{}) string {
	t := NewTable([]string{"CONTAINER", "READY", "STATUS", "RESTARTS", "EXIT CODE", "REASON"})
	return t.Render()
}

func RenderEventsTable(events interface{}) string {
	t := NewTable([]string{"TYPE", "REASON", "AGE", "FROM", "MESSAGE"})
	return t.Render()
}

func RenderResourceTable(resources interface{}) string {
	t := NewTable([]string{"NAME", "TYPE", "STATUS", "AGE"})
	return t.Render()
}