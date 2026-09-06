package ui

const (
	MinWidth  = 80
	MinHeight = 24

	tabBarHeight  = 1
	helpBarHeight = 1
)

type layout struct {
	width, height int
	contentHeight int
	tooSmall      bool
}

func computeLayout(width, height int) layout {
	return layout{
		width:         width,
		height:        height,
		contentHeight: max(height-tabBarHeight-helpBarHeight, 1),
		tooSmall:      width < MinWidth || height < MinHeight,
	}
}
