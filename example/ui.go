// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package main

import (
	"fmt"
	"image"
	"image/color"

	"github.com/ebitengine/microui"
)

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (g *Game) writeLog(text string) {
	if len(g.logBuf) > 0 {
		g.logBuf += "\n"
	}
	g.logBuf += text
	g.logUpdated = true
}

func (g *Game) testWindow() {
	g.ctx.Window("Demo Window", image.Rect(40, 40, 340, 490), func(res microui.Res) {
		win := g.ctx.CurrentContainer()
		win.Rect.Max.X = win.Rect.Min.X + max(win.Rect.Dx(), 240)
		win.Rect.Max.Y = win.Rect.Min.Y + max(win.Rect.Dy(), 300)

		// window info
		if g.ctx.Header("Window Info") != 0 {
			win := g.ctx.CurrentContainer()
			g.ctx.LayoutRow(2, []int{54, -1}, 0)
			g.ctx.Label("Position:")
			g.ctx.Label(fmt.Sprintf("%d, %d", win.Rect.Min.X, win.Rect.Min.Y))
			g.ctx.Label("Size:")
			g.ctx.Label(fmt.Sprintf("%d, %d", win.Rect.Dx(), win.Rect.Dy()))
		}

		// labels + buttons
		if g.ctx.HeaderEx("Test Buttons", microui.OptExpanded) != 0 {
			g.ctx.LayoutRow(3, []int{100, -110, -1}, 0)
			g.ctx.Label("Test buttons 1:")
			if g.ctx.Button("Button 1") != 0 {
				g.writeLog("Pressed button 1")
			}
			if g.ctx.Button("Button 2") != 0 {
				g.writeLog("Pressed button 2")
			}
			g.ctx.Label("Test buttons 2:")
			if g.ctx.Button("Button 3") != 0 {
				g.writeLog("Pressed button 3")
			}
			if g.ctx.Button("Popup") != 0 {
				g.ctx.OpenPopup("Test Popup")
			}
			g.ctx.Popup("Test Popup", func(res microui.Res) {
				g.ctx.Button("Hello")
				g.ctx.Button("World")
			})
		}

		// tree
		if g.ctx.HeaderEx("Tree and Text", microui.OptExpanded) != 0 {
			g.ctx.LayoutRow(2, []int{140, -1}, 0)
			g.ctx.LayoutColumn(func() {
				g.ctx.TreeNode("Test 1", func(res microui.Res) {
					g.ctx.TreeNode("Test 1a", func(res microui.Res) {
						g.ctx.Label("Hello")
						g.ctx.Label("World")
					})
					g.ctx.TreeNode("Test 1b", func(res microui.Res) {
						if g.ctx.Button("Button 1") != 0 {
							g.writeLog("Pressed button 1")
						}
						if g.ctx.Button("Button 2") != 0 {
							g.writeLog("Pressed button 2")
						}
					})
				})
				g.ctx.TreeNode("Test 2", func(res microui.Res) {
					g.ctx.LayoutRow(2, []int{54, 54}, 0)
					if g.ctx.Button("Button 3") != 0 {
						g.writeLog("Pressed button 3")
					}
					if g.ctx.Button("Button 4") != 0 {
						g.writeLog("Pressed button 4")
					}
					if g.ctx.Button("Button 5") != 0 {
						g.writeLog("Pressed button 5")
					}
					if g.ctx.Button("Button 6") != 0 {
						g.writeLog("Pressed button 6")
					}
				})
				g.ctx.TreeNode("Test 3", func(res microui.Res) {
					g.ctx.Checkbox("Checkbox 1", &g.checks[0])
					g.ctx.Checkbox("Checkbox 2", &g.checks[1])
					g.ctx.Checkbox("Checkbox 3", &g.checks[2])
				})
			})

			g.ctx.Text("Lorem ipsum dolor sit amet, consectetur adipiscing " +
				"elit. Maecenas lacinia, sem eu lacinia molestie, mi risus faucibus " +
				"ipsum, eu varius magna felis a nulla.")
		}

		// background color sliders
		if g.ctx.HeaderEx("Background Color", microui.OptExpanded) != 0 {
			g.ctx.LayoutRow(2, []int{-78, -1}, 74)
			// sliders
			g.ctx.LayoutColumn(func() {
				g.ctx.LayoutRow(2, []int{46, -1}, 0)
				g.ctx.Label("Red:")
				g.ctx.Slider(&g.bg[0], 0, 255)
				g.ctx.Label("Green:")
				g.ctx.Slider(&g.bg[1], 0, 255)
				g.ctx.Label("Blue:")
				g.ctx.Slider(&g.bg[2], 0, 255)
			})
			// color preview
			r := g.ctx.LayoutNext()
			g.ctx.DrawRect(r, color.RGBA{byte(g.bg[0]), byte(g.bg[1]), byte(g.bg[2]), 255})
			clr := fmt.Sprintf("#%02X%02X%02X", int(g.bg[0]), int(g.bg[1]), int(g.bg[2]))
			g.ctx.DrawControlText(clr, r, microui.ColorText, microui.OptAlignCenter)
		}
	})
}

func (g *Game) logWindow() {
	g.ctx.Window("Log Window", image.Rect(350, 40, 650, 240), func(res microui.Res) {
		// output text panel
		g.ctx.LayoutRow(1, []int{-1}, -25)
		var panel *microui.Container
		g.ctx.Panel("Log Output", func() {
			panel = g.ctx.CurrentContainer()
			g.ctx.LayoutRow(1, []int{-1}, -1)
			g.ctx.Text(g.logBuf)
		})
		if g.logUpdated {
			panel.Scroll.Y = panel.ContentSize.Y
			g.logUpdated = false
		}

		// input textbox + submit button
		var submitted bool
		g.ctx.LayoutRow(2, []int{-70, -1}, 0)
		if g.ctx.TextBox(&g.logSubmitBuf)&microui.ResSubmit != 0 {
			g.ctx.SetFocus(g.ctx.LastID)
			submitted = true
		}
		if g.ctx.Button("Submit") != 0 {
			submitted = true
		}
		if submitted {
			g.writeLog(g.logSubmitBuf)
			g.logSubmitBuf = ""
		}
	})
}

func (g *Game) byteSlider(fvalue *float64, value *byte, low, high byte) microui.Res {
	*fvalue = float64(*value)
	res := g.ctx.SliderEx(fvalue, float64(low), float64(high), 0, "%.0f", microui.OptAlignCenter)
	*value = byte(*fvalue)
	return res
}

var (
	fcolors = [14]struct {
		R, G, B, A float64
	}{}
	colors = []struct {
		Label   string
		ColorID int
	}{
		{"text:", microui.ColorText},
		{"border:", microui.ColorBorder},
		{"windowbg:", microui.ColorWindowBG},
		{"titlebg:", microui.ColorTitleBG},
		{"titletext:", microui.ColorTitleText},
		{"panelbg:", microui.ColorPanelBG},
		{"button:", microui.ColorButton},
		{"buttonhover:", microui.ColorButtonHover},
		{"buttonfocus:", microui.ColorButtonFocus},
		{"base:", microui.ColorBase},
		{"basehover:", microui.ColorBaseHover},
		{"basefocus:", microui.ColorBaseFocus},
		{"scrollbase:", microui.ColorScrollBase},
		{"scrollthumb:", microui.ColorScrollThumb},
	}
)

func (g *Game) styleWindow() {
	g.ctx.Window("Style Editor", image.Rect(350, 250, 650, 490), func(res microui.Res) {
		sw := int(float64(g.ctx.CurrentContainer().Body.Dx()) * 0.14)
		g.ctx.LayoutRow(6, []int{80, sw, sw, sw, sw, -1}, 0)
		for _, c := range colors {
			g.ctx.Label(c.Label)
			g.byteSlider(&fcolors[c.ColorID].R, &g.ctx.Style.Colors[c.ColorID].R, 0, 255)
			g.byteSlider(&fcolors[c.ColorID].G, &g.ctx.Style.Colors[c.ColorID].G, 0, 255)
			g.byteSlider(&fcolors[c.ColorID].B, &g.ctx.Style.Colors[c.ColorID].B, 0, 255)
			g.byteSlider(&fcolors[c.ColorID].A, &g.ctx.Style.Colors[c.ColorID].A, 0, 255)
			g.ctx.DrawRect(g.ctx.LayoutNext(), g.ctx.Style.Colors[c.ColorID])
		}
	})
}

func (g *Game) ProcessFrame() {
	g.ctx.Update(func() {
		g.testWindow()
		g.logWindow()
		g.styleWindow()
	})
}
