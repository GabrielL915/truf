package components

import (
	"fmt"
	"math"
	"strings"

	"github.com/NimbleMarkets/ntcharts/canvas"
	"github.com/NimbleMarkets/ntcharts/canvas/graph"
	"github.com/charmbracelet/lipgloss"
	"github.com/gabriel-luiz/truf/internal/ledger"
	"github.com/gabriel-luiz/truf/internal/ui/styles"
	"github.com/gabriel-luiz/truf/pkg/utils"
)

type Chart struct {
	Data    ledger.ChartData
	Focused bool
	Width   int
	Height  int
}

func NewChart() *Chart {
	return &Chart{}
}

func (c *Chart) SetData(data ledger.ChartData) {
	c.Data = data
}

func (c *Chart) SetSize(width, height int) {
	c.Width = width
	c.Height = height
}

func (c *Chart) View() string {
	if len(c.Data.Months) == 0 {
		return c.renderEmpty()
	}
	return c.renderChart()
}

func (c *Chart) renderEmpty() string {
	content := lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Render("No data available.\nAdd income or expenses to see your balance chart.")

	return c.panelStyle().
		Width(c.Width - 2).
		Height(c.Height - 2).
		Render(content)
}

func (c *Chart) panelStyle() lipgloss.Style {
	if c.Focused {
		return styles.FocusedPanelStyle
	}
	return styles.PanelStyle
}

func (c *Chart) renderChart() string {
	var sb strings.Builder

	title := lipgloss.NewStyle().Foreground(styles.Primary).Bold(true).Render("Balance Over Time")
	summary := ""
	if n := len(c.Data.Balance); n > 0 {
		last := c.Data.Balance[n-1]
		col := styles.Success
		if last < 0 {
			col = styles.Danger
		}
		summary = "  " + lipgloss.NewStyle().Foreground(col).Render(utils.FormatCurrency(last))
	}
	sb.WriteString(title + summary + "\n\n")

	yLabelWidth := 9
	chartW := c.Width - yLabelWidth - 6
	chartH := c.Height - 9
	if chartW < 10 {
		chartW = 10
	}
	if chartH < 4 {
		chartH = 4
	}

	minY, maxY := c.dataRange()
	n := len(c.Data.Balance)
	minX := 0.0
	maxX := float64(n - 1)
	if maxX == 0 {
		maxX = 1
	}

	cnv := canvas.New(chartW, chartH)
	c.drawSeries(&cnv, toFloat(c.Data.Income), minX, maxX, minY, maxY,
		lipgloss.NewStyle().Foreground(styles.Success))
	c.drawSeries(&cnv, toFloat(c.Data.Expenses), minX, maxX, minY, maxY,
		lipgloss.NewStyle().Foreground(styles.Danger))
	c.drawSeries(&cnv, toFloat(c.Data.Balance), minX, maxX, minY, maxY,
		lipgloss.NewStyle().Foreground(styles.Accent))

	canvasStr := cnv.View()
	rows := strings.Split(strings.TrimRight(canvasStr, "\n"), "\n")
	for i, row := range rows {
		var yLabel string
		switch i {
		case 0:
			yLabel = fmt.Sprintf("%*s", yLabelWidth, utils.FormatCurrency(int64(math.Round(maxY))))
		case len(rows) - 1:
			yLabel = fmt.Sprintf("%*s", yLabelWidth, utils.FormatCurrency(int64(math.Round(minY))))
		default:
			yLabel = strings.Repeat(" ", yLabelWidth)
		}
		sb.WriteString(lipgloss.NewStyle().Foreground(styles.TextMuted).Render(yLabel))
		sb.WriteString(" ")
		sb.WriteString(row)
		sb.WriteString("\n")
	}

	sb.WriteString(c.renderXLabels(yLabelWidth+1, chartW))
	sb.WriteString("\n")

	sb.WriteString(c.renderLegend())

	return c.panelStyle().
		Width(c.Width - 2).
		Height(c.Height - 2).
		Render(sb.String())
}

func (c *Chart) drawSeries(cnv *canvas.Model, data []float64, minX, maxX, minY, maxY float64, s lipgloss.Style) {
	if len(data) == 0 {
		return
	}

	w := cnv.Width()
	h := cnv.Height()
	if w <= 0 || h <= 0 {
		return
	}

	bg := graph.NewBrailleGrid(w, h, minX, maxX, minY, maxY)

	for i, v := range data {
		p1 := bg.GridPoint(canvas.Float64Point{X: float64(i), Y: v})
		bg.Set(p1)

		if i > 0 {
			prev := data[i-1]
			p0 := bg.GridPoint(canvas.Float64Point{X: float64(i - 1), Y: prev})
			interpolateBraille(bg, p0, p1)
		}
	}

	graph.DrawBraillePatterns(cnv, canvas.Point{X: 0, Y: 0}, bg.BraillePatterns(), s)
}

func interpolateBraille(bg *graph.BrailleGrid, p0, p1 canvas.Point) {
	dx := p1.X - p0.X
	dy := p1.Y - p0.Y
	steps := abs2(dx)
	if abs2(dy) > steps {
		steps = abs2(dy)
	}
	if steps == 0 {
		return
	}
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		x := int(math.Round(float64(p0.X) + t*float64(dx)))
		y := int(math.Round(float64(p0.Y) + t*float64(dy)))
		bg.Set(canvas.Point{X: x, Y: y})
	}
}

func abs2(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (c *Chart) dataRange() (minY, maxY float64) {
	var all []int64
	all = append(all, c.Data.Balance...)
	all = append(all, c.Data.Income...)
	all = append(all, c.Data.Expenses...)
	if len(all) == 0 {
		return 0, 1
	}
	minY, maxY = float64(all[0]), float64(all[0])
	for _, v := range all[1:] {
		minY = math.Min(minY, float64(v))
		maxY = math.Max(maxY, float64(v))
	}
	if minY == maxY {
		minY -= 1
		maxY += 1
	}
	return minY, maxY
}

func toFloat(cents []int64) []float64 {
	out := make([]float64, len(cents))
	for i, v := range cents {
		out[i] = float64(v)
	}
	return out
}

func (c *Chart) renderXLabels(offset, chartW int) string {
	n := len(c.Data.Months)
	if n == 0 {
		return ""
	}
	spacing := chartW / n
	if spacing < 1 {
		spacing = 1
	}

	var sb strings.Builder
	sb.WriteString(strings.Repeat(" ", offset))
	for i, month := range c.Data.Months {
		label := month
		if len(label) > spacing {
			label = label[:spacing]
		}
		pos := i * spacing
		_ = pos
		sb.WriteString(fmt.Sprintf("%-*s", spacing, label))
	}
	return sb.String()
}

func (c *Chart) renderLegend() string {
	income := lipgloss.NewStyle().Foreground(styles.Success).Render("⣿ Income")
	expense := lipgloss.NewStyle().Foreground(styles.Danger).Render("⣿ Expenses")
	balance := lipgloss.NewStyle().Foreground(styles.Accent).Render("⣿ Balance")
	return fmt.Sprintf("\n%s  %s  %s", income, expense, balance)
}
