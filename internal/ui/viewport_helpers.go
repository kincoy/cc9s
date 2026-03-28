package ui

import (
	"charm.land/bubbles/v2/viewport"
)

// NewViewportWithSize creates a viewport with given dimensions
func NewViewportWithSize(width, height int) viewport.Model {
	vp := viewport.New(
		viewport.WithWidth(width),
		viewport.WithHeight(height),
	)
	return vp
}

// EnsureLineVisible scrolls viewport to make lineNum visible
func EnsureLineVisible(vp *viewport.Model, lineNum int) {
	viewportTop := vp.YOffset()
	viewportBottom := viewportTop + vp.Height()

	if lineNum < viewportTop {
		vp.SetYOffset(lineNum)
	} else if lineNum >= viewportBottom {
		newOffset := lineNum - vp.Height() + 1
		if newOffset < 0 {
			newOffset = 0
		}
		vp.SetYOffset(newOffset)
	}
}
