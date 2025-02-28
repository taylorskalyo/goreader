package render

import "github.com/jedib0t/go-pretty/v6/table"

// columnConfigs configures columns to fit within the viewport. It sets each
// column's max width and handles text wrapping.
func (r *Renderer) columnConfigs(style table.Style) (configs []table.ColumnConfig) {
	columnCount := 0
	for _, row := range r.parser.rowStack {
		if len(row) > columnCount {
			columnCount = len(row)
		}
	}

	if columnCount < 1 {
		return configs
	}

	// Determine width availble for text.
	availableWidth := r.width - tableDecorationWidth(style, columnCount)

	// Initialize column configs.
	configs = make([]table.ColumnConfig, columnCount)
	for i := range configs {
		configs[i] = table.ColumnConfig{
			Number:           i + 1,
			WidthMaxEnforcer: tviewWidthEnforcer,
		}
	}

	// Dynamically choose column width.
	strategies := []func(int, *[]table.ColumnConfig) bool{
		r.parser.tryFitColumn,
		r.parser.tryFairColumn,
	}
	for _, fn := range strategies {
		if ok := fn(availableWidth, &configs); ok {
			return configs
		}
	}

	return []table.ColumnConfig{}
}

// tryFitColumn fits each column width to its cell content. If the overall
// width is greater than the available width, it will abort.
func (p *parser) tryFitColumn(availableWidth int, configs *[]table.ColumnConfig) bool {
	maxWidths := make([]int, len(*configs))
	for _, row := range p.rowStack {
		for col, cell := range row {
			if len(cell) > maxWidths[col] {
				maxWidths[col] = len(cell)
			}
		}
	}

	totalMaxWidth := 0
	for _, width := range maxWidths {
		totalMaxWidth += width
	}

	if totalMaxWidth > availableWidth {
		return false
	}

	for i := range *configs {
		(*configs)[i].WidthMax = maxWidths[i]
	}

	return true
}

// tryFairColumn gives each column a proportion of the available width with a
// bias towards equal widths.
func (p *parser) tryFairColumn(availableWidth int, configs *[]table.ColumnConfig) bool {
	equalWidth := availableWidth / len(*configs)
	fairWidths := make([]int, len(*configs))
	for _, row := range p.rowStack {
		for col, cell := range row {
			fairWidth := (len(cell) + equalWidth) / 2
			if fairWidth > fairWidths[col] {
				fairWidths[col] = fairWidth
			}
		}
	}

	totalFairWidth := 0
	for _, width := range fairWidths {
		totalFairWidth += width
	}

	for i := range *configs {
		ratio := float64(fairWidths[i]) / float64(totalFairWidth)
		width := int(float64(availableWidth) * ratio)
		(*configs)[i].WidthMax = width
	}

	return true
}

// tableDecorationWidth determines how much width is needed to display table
// decorations using the given table style.
func tableDecorationWidth(style table.Style, columnCount int) int {
	// Handle padding.
	width := 2 * columnCount

	// Handle borders.
	if style.Options.DrawBorder {
		width += 2
	}

	// Handle separators.
	if style.Options.SeparateColumns {
		width += columnCount - 1
	}

	return width
}
