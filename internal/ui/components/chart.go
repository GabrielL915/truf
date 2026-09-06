package components

import (
	"math"
	"strconv"
	"strings"

	"github.com/NimbleMarkets/ntcharts/canvas"
	"github.com/NimbleMarkets/ntcharts/canvas/graph"
	"github.com/charmbracelet/lipgloss"
	"github.com/gabriel-luiz/truf/internal/ledger"
	"github.com/gabriel-luiz/truf/internal/ui/styles"
	"github.com/gabriel-luiz/truf/pkg/utils"
)

const (
	chartHeaderLines = 2
	chartAxisLines   = 1
	chartGridlines   = 5
	minGutter        = 8
)

type Chart struct {
	Data        ledger.ChartData
	ActiveMonth string
	Width       int
	Height      int
}

func NewChart() *Chart {
	return &Chart{}
}

func (c *Chart) SetData(data ledger.ChartData) {
	c.Data = data
}

func (c *Chart) SetActiveMonth(label string) {
	c.ActiveMonth = label
}

func (c *Chart) SetSize(width, height int) {
	c.Width = width
	c.Height = height
}

func (c *Chart) View() string {
	if len(c.Data.Months) == 0 {
		return styles.MutedStyle.
			Width(c.Width).
			Height(c.Height).
			MaxHeight(c.Height).
			Render("No data available.")
	}
	return c.renderChart()
}

func (c *Chart) renderChart() string {
	var sb strings.Builder
	sb.WriteString(c.renderHeader())
	sb.WriteString("\n\n")

	minY, maxY, step := niceRange(c.dataRange())
	labels := gridLabels(minY, maxY, step)
	gutter := minGutter
	for _, l := range labels {
		gutter = max(gutter, len(l)+1)
	}

	chartW := max(c.Width-gutter-2, 10)
	chartH := max(c.Height-chartHeaderLines-chartAxisLines, 4)

	n := len(c.Data.Balance)
	minX, maxX := 0.0, float64(n-1)
	if maxX <= 0 {
		maxX = 1
	}

	cnv := canvas.New(chartW, chartH)
	grid := graph.NewBrailleGrid(chartW, chartH, minX, maxX, minY, maxY)
	drawSeries(&cnv, grid, toFloat(c.Data.Income), lipgloss.NewStyle().Foreground(styles.Success))
	drawSeries(&cnv, grid, toFloat(c.Data.Expenses), lipgloss.NewStyle().Foreground(styles.Danger))
	drawSeries(&cnv, grid, toFloat(c.Data.Balance), lipgloss.NewStyle().Foreground(styles.Net))

	rowLabels := map[int]string{}
	for i, label := range labels {
		v := minY + float64(i)*step
		row := grid.GridPoint(canvas.Float64Point{X: 0, Y: v}).Y / 4
		if row < 0 || row >= chartH {
			continue
		}
		if _, taken := rowLabels[row]; taken {
			continue
		}
		rowLabels[row] = label
		drawGridline(&cnv, row)
	}

	rows := strings.Split(strings.TrimRight(cnv.View(), "\n"), "\n")
	for row, line := range rows {
		sb.WriteString(styles.MutedStyle.Width(gutter).Align(lipgloss.Right).Render(rowLabels[row]))
		sb.WriteString(" ")
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	sb.WriteString(c.renderXLabels(gutter+1, chartW))

	return lipgloss.NewStyle().
		Width(c.Width).
		Height(c.Height).
		MaxHeight(c.Height).
		Render(sb.String())
}

func (c *Chart) renderHeader() string {
	title := styles.TitleStyle.Render("Balance · last " + strconv.Itoa(len(c.Data.Months)) + " months")
	legend := lipgloss.NewStyle().Foreground(styles.Success).Render("⣿ Income") + "  " +
		lipgloss.NewStyle().Foreground(styles.Danger).Render("⣿ Expenses") + "  " +
		lipgloss.NewStyle().Foreground(styles.Net).Render("⣿ Balance")
	gap := max(c.Width-2-lipgloss.Width(title)-lipgloss.Width(legend), 1)
	return " " + title + strings.Repeat(" ", gap) + legend
}

func drawSeries(cnv *canvas.Model, grid *graph.BrailleGrid, data []float64, s lipgloss.Style) {
	if len(data) == 0 || cnv.Width() <= 0 || cnv.Height() <= 0 {
		return
	}
	grid.Clear()
	for i, v := range data {
		p1 := grid.GridPoint(canvas.Float64Point{X: float64(i), Y: v})
		grid.Set(p1)
		if i > 0 {
			p0 := grid.GridPoint(canvas.Float64Point{X: float64(i - 1), Y: data[i-1]})
			interpolateBraille(grid, p0, p1)
		}
	}
	graph.DrawBraillePatterns(cnv, canvas.Point{X: 0, Y: 0}, grid.BraillePatterns(), s)
}

func drawGridline(cnv *canvas.Model, row int) {
	cell := canvas.NewCellWithStyle('┈', styles.ChromeStyle)
	for x := 0; x < cnv.Width(); x++ {
		p := canvas.Point{X: x, Y: row}
		if cnv.Cell(p).Rune == 0 {
			cnv.SetCell(p, cell)
		}
	}
}

func interpolateBraille(bg *graph.BrailleGrid, p0, p1 canvas.Point) {
	dx := p1.X - p0.X
	dy := p1.Y - p0.Y
	steps := max(abs(dx), abs(dy))
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

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (c *Chart) dataRange() (minY, maxY float64) {
	first := true
	for _, series := range [][]int64{c.Data.Balance, c.Data.Income, c.Data.Expenses} {
		for _, v := range series {
			f := float64(v)
			if first {
				minY, maxY, first = f, f, false
				continue
			}
			minY = math.Min(minY, f)
			maxY = math.Max(maxY, f)
		}
	}
	if first {
		return 0, 1
	}
	if minY == maxY {
		minY -= 1
		maxY += 1
	}
	return minY, maxY
}

func niceRange(minY, maxY float64) (lo, hi, step float64) {
	step = niceStep((maxY - minY) / float64(chartGridlines-1))
	lo = math.Floor(minY/step) * step
	hi = math.Ceil(maxY/step) * step
	if lo == hi {
		hi = lo + step
	}
	return lo, hi, step
}

func niceStep(raw float64) float64 {
	if raw <= 1 || math.IsNaN(raw) || math.IsInf(raw, 0) {
		return 1
	}
	mag := math.Pow(10, math.Floor(math.Log10(raw)))
	switch f := raw / mag; {
	case f <= 1:
		return mag
	case f <= 2:
		return 2 * mag
	case f <= 5:
		return 5 * mag
	}
	return 10 * mag
}

func gridLabels(lo, hi, step float64) []string {
	var labels []string
	for v := lo; v <= hi+step/2 && len(labels) < 2*chartGridlines; v += step {
		labels = append(labels, compactCurrency(v))
	}
	return labels
}

func compactCurrency(v float64) string {
	if v >= math.MaxInt64 || v <= math.MinInt64 {
		return ""
	}
	return strings.TrimSuffix(utils.FormatCurrency(int64(math.Round(v))), ",00")
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
	spacing := max(chartW/n, 1)

	var sb strings.Builder
	sb.WriteString(strings.Repeat(" ", offset))
	for _, month := range c.Data.Months {
		label := []rune(month)
		if len(label) > spacing {
			label = label[:spacing]
		}
		style := styles.MutedStyle
		if month == c.ActiveMonth {
			style = styles.TextStyle.Bold(true)
		}
		sb.WriteString(style.Width(spacing).Render(string(label)))
	}
	return sb.String()
}
