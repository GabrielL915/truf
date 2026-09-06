package ui

const (
	MinWidth  = 80
	MinHeight = 24

	tabBarHeight  = 1
	helpBarHeight = 1
	menuWidth     = 16
)

type layout struct {
	width, height int
	contentHeight int
	menuWidth     int
	mainWidth     int
	tooSmall      bool
}

func computeLayout(width, height int) layout {
	l := layout{width: width, height: height}
	l.tooSmall = width < MinWidth || height < MinHeight
	l.contentHeight = max(height-tabBarHeight-helpBarHeight, 1)
	l.menuWidth = min(menuWidth, width)
	l.mainWidth = max(width-l.menuWidth, 1)
	return l
}
