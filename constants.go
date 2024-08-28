// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package microui

import (
	"image"
	"image/color"
)

const (
	commandListSize    = 256 * 1024
	rootListSize       = 32
	containerStackSize = 32
	clipStackSize      = 32
	idStackSize        = 32
	layoutStackSize    = 16
	containerPoolSize  = 48
	treeNodePoolSize   = 48
	maxWidths          = 16
)

const (
	realFmt   = "%.3g"
	sliderFmt = "%.2f"
)

var defaultStyle Style = Style{
	Font:          nil,
	Size:          image.Pt(68, 10),
	Padding:       5,
	Spacing:       4,
	Indent:        24,
	TitleHeight:   24,
	ScrollbarSize: 12,
	ThumbSize:     8,
	Colors: [...]color.RGBA{
		{230, 230, 230, 255}, // MU_COLOR_TEXT
		{25, 25, 25, 255},    // MU_COLOR_BORDER
		{50, 50, 50, 255},    // MU_COLOR_WINDOWBG
		{25, 25, 25, 255},    // MU_COLOR_TITLEBG
		{240, 240, 240, 255}, // MU_COLOR_TITLETEXT
		{0, 0, 0, 0},         // MU_COLOR_PANELBG
		{75, 75, 75, 255},    // MU_COLOR_BUTTON
		{95, 95, 95, 255},    // MU_COLOR_BUTTONHOVER
		{115, 115, 115, 255}, // MU_COLOR_BUTTONFOCUS
		{30, 30, 30, 255},    // MU_COLOR_BASE
		{35, 35, 35, 255},    // MU_COLOR_BASEHOVER
		{40, 40, 40, 255},    // MU_COLOR_BASEFOCUS
		{43, 43, 43, 255},    // MU_COLOR_SCROLLBASE
		{30, 30, 30, 255},    // MU_COLOR_SCROLLTHUMB
	},
}

var (
	unclippedRect = image.Rect(0, 0, 0x1000000, 0x1000000)
)
