// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package debugui

import (
	"fmt"
	"image"
	"math"
	"os"
	"strconv"
	"unicode/utf8"
	"unsafe"

	"github.com/hajimehoshi/ebiten/v2/exp/textinput"
)

// inHoverRoot determines if the hover state is within the current root container by checking the container stack.
func (c *Context) inHoverRoot() bool {
	for i := len(c.containerStack) - 1; i >= 0; i-- {
		if c.containerStack[i] == c.hoverRoot {
			return true
		}
		// only root containers have their `head` field set; stop searching if we've
		// reached the current root container
		if c.containerStack[i].headIdx >= 0 {
			break
		}
	}
	return false
}

// drawControlFrame renders a frame around a control with styling based on focus, hover state, and provided options.
func (c *Context) drawControlFrame(id controlID, rect image.Rectangle, colorid int, opt option) {
	if (opt & optionNoFrame) != 0 {
		return
	}
	if c.focus == id {
		colorid += 2
	} else if c.hover == id {
		colorid++
	}
	c.drawFrame(rect, colorid)
}

// drawControlText renders a given string within a specified rectangle using the provided color and alignment options.
func (c *Context) drawControlText(str string, rect image.Rectangle, colorid int, opt option) {
	var pos image.Point
	tw := textWidth(str)
	c.pushClipRect(rect)
	pos.Y = rect.Min.Y + (rect.Dy()-lineHeight())/2
	if (opt & optionAlignCenter) != 0 {
		pos.X = rect.Min.X + (rect.Dx()-tw)/2
	} else if (opt & optionAlignRight) != 0 {
		pos.X = rect.Min.X + rect.Dx() - tw - c.style.padding
	} else {
		pos.X = rect.Min.X + c.style.padding
	}
	c.drawText(str, pos, c.style.colors[colorid])
	c.popClipRect()
}

// mouseOver checks if the mouse position is within the given rectangle, the clip rectangle, and the hover root.
func (c *Context) mouseOver(rect image.Rectangle) bool {
	return c.mousePos.In(rect) && c.mousePos.In(c.clipRect()) && c.inHoverRoot()
}

// updateControl updates the state of a UI control based on its ID, bounding rectangle, and interaction options.
func (c *Context) updateControl(id controlID, rect image.Rectangle, opt option) {
	if id == 0 {
		return
	}

	mouseover := c.mouseOver(rect)

	if c.focus == id {
		c.keepFocus = true
	}
	if (opt & optionNoInteract) != 0 {
		return
	}
	if mouseover && c.mouseDown == 0 {
		c.hover = id
	}

	if c.focus == id {
		if c.mousePressed != 0 && !mouseover {
			c.setFocus(0)
		}
		if c.mouseDown == 0 && (^opt&optionHoldFocus) != 0 {
			c.setFocus(0)
		}
	}

	if c.hover == id {
		if c.mousePressed != 0 {
			c.setFocus(id)
		} else if !mouseover {
			c.hover = 0
		}
	}
}

// Control executes a provided function within a specific context, managing identifier lifecycle for the control element.
func (c *Context) Control(idStr string, f func(r image.Rectangle) Response) Response {
	id := c.pushID([]byte(idStr))
	defer c.popID()
	return c.control(id, 0, f)
}

// control manages a UI control within a given layout and executes a callback function with the control's rectangle.
// It updates the control's state based on the provided id and options, returning the callback's response.
func (c *Context) control(id controlID, opt option, f func(r image.Rectangle) Response) Response {
	r := c.layoutNext()
	c.updateControl(id, r, opt)
	return f(r)
}

// Text renders the provided text string within the context, wrapping it within the available width of the layout.
func (c *Context) Text(text string) {
	color := c.style.colors[ColorText]
	c.LayoutColumn(func() {
		var endIdx, p int
		c.SetLayoutRow([]int{-1}, lineHeight())
		for endIdx < len(text) {
			c.control(0, 0, func(r image.Rectangle) Response {
				w := 0
				endIdx = p
				startIdx := endIdx
				for endIdx < len(text) && text[endIdx] != '\n' {
					word := p
					for p < len(text) && text[p] != ' ' && text[p] != '\n' {
						p++
					}
					w += textWidth(text[word:p])
					if w > r.Dx() && endIdx != startIdx {
						break
					}
					if p < len(text) {
						w += textWidth(string(text[p]))
					}
					endIdx = p
					p++
				}
				c.drawText(text[startIdx:endIdx], r.Min, color)
				p = endIdx + 1
				return 0
			})
		}
	})
}

// Label renders a text label within the specified control area using the provided text string.
func (c *Context) Label(text string) {
	c.control(0, 0, func(r image.Rectangle) Response {
		c.drawControlText(text, r, ColorText, 0)
		return 0
	})
}

// button creates a button control with the given label and id, applies options, and returns the interaction response.
func (c *Context) button(label string, idStr string, opt option) Response {
	var id controlID
	if len(idStr) > 0 {
		id = c.pushID([]byte(idStr))
		defer c.popID()
	} else if len(label) > 0 {
		id = c.pushID([]byte(label))
		defer c.popID()
	}
	return c.control(id, opt, func(r image.Rectangle) Response {
		var res Response
		// handle click
		if c.mousePressed == mouseLeft && c.focus == id {
			res |= ResponseSubmit
		}
		// draw
		c.drawControlFrame(id, r, ColorButton, opt)
		if len(label) > 0 {
			c.drawControlText(label, r, ColorText, opt)
		}
		return res
	})
}

// Checkbox renders a checkbox with a label and manages its state based on user interaction.
// The label specifies the text displayed next to the checkbox.
// The state pointer determines the checkbox's current state and reflects any user updates.
// Returns a Response indicating the interactions or state changes of the checkbox.
func (c *Context) Checkbox(label string, state *bool) Response {
	id := c.pushID(ptrToBytes(unsafe.Pointer(state)))
	defer c.popID()

	return c.control(id, 0, func(r image.Rectangle) Response {
		var res Response
		box := image.Rect(r.Min.X, r.Min.Y, r.Min.X+r.Dy(), r.Max.Y)
		c.updateControl(id, r, 0)
		// handle click
		if c.mousePressed == mouseLeft && c.focus == id {
			res |= ResponseChange
			*state = !*state
		}
		// draw
		c.drawControlFrame(id, box, ColorBase, 0)
		if *state {
			c.drawIcon(iconCheck, box, c.style.colors[ColorText])
		}
		r = image.Rect(r.Min.X+box.Dx(), r.Min.Y, r.Max.X, r.Max.Y)
		c.drawControlText(label, r, ColorText, 0)
		return res
	})
}

// textField retrieves or initializes a text input field associated with the given controlID.
func (c *Context) textField(id controlID) *textinput.Field {
	if id == 0 {
		return nil
	}
	if _, ok := c.textFields[id]; !ok {
		if c.textFields == nil {
			c.textFields = make(map[controlID]*textinput.Field)
		}
		c.textFields[id] = &textinput.Field{}
	}
	return c.textFields[id]
}

// textBoxRaw manages low-level text box rendering and behavior, handling user interaction, keyboard input, and focus state.
// Handles text input, editing, focus, key presses (backspace/return), and updates the text buffer if changes occur.
// Draws the text box and its contents based on the current state and specified options.
// Returns a Response indicating whether the text has changed or the textbox was submitted.
func (c *Context) textBoxRaw(buf *string, id controlID, opt option) Response {
	return c.control(id, opt|optionHoldFocus, func(r image.Rectangle) Response {
		var res Response

		if c.focus == id {
			// handle text input
			f := c.textField(id)
			f.Focus()
			x := r.Min.X + c.style.padding + textWidth(*buf)
			y := r.Min.Y + lineHeight()
			handled, err := f.HandleInput(x, y)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return 0
			}
			if *buf != f.TextForRendering() {
				*buf = f.TextForRendering()
				res |= ResponseChange
			}

			if !handled {
				// handle backspace
				if (c.keyPressed&keyBackspace) != 0 && len(*buf) > 0 {
					_, size := utf8.DecodeLastRuneInString(*buf)
					*buf = (*buf)[:len(*buf)-size]
					f.SetTextAndSelection(*buf, len(*buf), len(*buf))
					res |= ResponseChange
				}

				// handle return
				if (c.keyPressed & keyReturn) != 0 {
					c.setFocus(0)
					res |= ResponseSubmit
					f.SetTextAndSelection("", 0, 0)
				}
			}
		} else {
			f := c.textField(id)
			if *buf != f.TextForRendering() {
				f.SetTextAndSelection(*buf, len(*buf), len(*buf))
			}
		}

		// draw
		c.drawControlFrame(id, r, ColorBase, opt)
		if c.focus == id {
			color := c.style.colors[ColorText]
			textw := textWidth(*buf)
			texth := lineHeight()
			ofx := r.Dx() - c.style.padding - textw - 1
			textx := r.Min.X + min(ofx, c.style.padding)
			texty := r.Min.Y + (r.Dy()-texth)/2
			c.pushClipRect(r)
			c.drawText(*buf, image.Pt(textx, texty), color)
			c.drawRect(image.Rect(textx+textw, texty, textx+textw+1, texty+texth), color)
			c.popClipRect()
		} else {
			c.drawControlText(*buf, r, ColorText, opt)
		}
		return res
	})
}

// numberTextBox renders an editable numeric text box tied to a float64 value and handles input and focus behavior.
func (c *Context) numberTextBox(value *float64, id controlID) bool {
	if c.mousePressed == mouseLeft && (c.keyDown&keyShift) != 0 &&
		c.hover == id {
		c.numberEdit = id
		c.numberEditBuf = fmt.Sprintf(realFmt, *value)
	}
	if c.numberEdit == id {
		res := c.textBoxRaw(&c.numberEditBuf, id, 0)
		if (res&ResponseSubmit) != 0 || c.focus != id {
			nval, err := strconv.ParseFloat(c.numberEditBuf, 32)
			if err != nil {
				nval = 0
			}
			*value = float64(nval)
			c.numberEdit = 0
		}
		return true
	}
	return false
}

// textBox updates the text input box based on a provided string buffer and options, returning a Response status.
// It uniquely identifies the text box by generating an ID from the buffer's memory address.
// The method interacts with textBoxRaw to handle the input box rendering and behavior using the computed ID and options.
func (c *Context) textBox(buf *string, opt option) Response {
	id := c.pushID(ptrToBytes(unsafe.Pointer(buf)))
	defer c.popID()

	return c.textBoxRaw(buf, id, opt)
}

// formatNumber formats a floating-point number `v` to a string with a specified number of decimal places `digits`.
func formatNumber(v float64, digits int) string {
	return fmt.Sprintf("%."+strconv.Itoa(digits)+"f", v)
}

// slider is a method for rendering and handling a slider control for inputting float values within a specified range.
// The slider supports optional configurations such as step size for increments and the number of digits to display.
// It updates the passed value pointer, clamping it within the provided low and high bounds during interaction.
// Returns a Response indicating changes or interactions with the slider.
func (c *Context) slider(value *float64, low, high, step float64, digits int, opt option) Response {
	last := *value
	v := last
	id := c.pushID(ptrToBytes(unsafe.Pointer(value)))
	defer c.popID()

	// handle text input mode
	if c.numberTextBox(&v, id) {
		return 0
	}

	// handle normal mode
	return c.control(id, opt, func(r image.Rectangle) Response {
		var res Response
		// handle input
		if c.focus == id && (c.mouseDown|c.mousePressed) == mouseLeft {
			v = low + float64(c.mousePos.X-r.Min.X)*(high-low)/float64(r.Dx())
			if step != 0 {
				v = math.Round(v/step) * step
			}
		}
		// clamp and store value, update res
		*value = clampF(v, low, high)
		v = *value
		if last != v {
			res |= ResponseChange
		}

		// draw base
		c.drawControlFrame(id, r, ColorBase, opt)
		// draw thumb
		w := c.style.thumbSize
		x := int((v - low) * float64(r.Dx()-w) / (high - low))
		thumb := image.Rect(r.Min.X+x, r.Min.Y, r.Min.X+x+w, r.Max.Y)
		c.drawControlFrame(id, thumb, ColorButton, opt)
		// draw text
		text := formatNumber(v, digits)
		c.drawControlText(text, r, ColorText, opt)

		return res
	})
}

// number creates and handles a numeric input control with specified `step` and `digits`, updating the `value`.
// It uses `opt` for additional configuration and returns a `Response` indicating the control state.
func (c *Context) number(value *float64, step float64, digits int, opt option) Response {
	id := c.pushID(ptrToBytes(unsafe.Pointer(value)))
	defer c.popID()
	last := *value

	// handle text input mode
	if c.numberTextBox(value, id) {
		return 0
	}

	// handle normal mode
	return c.control(id, opt, func(r image.Rectangle) Response {
		var res Response
		// handle input
		if c.focus == id && c.mouseDown == mouseLeft {
			*value += float64(c.mouseDelta.X) * step
		}
		// set flag if value changed
		if *value != last {
			res |= ResponseChange
		}

		// draw base
		c.drawControlFrame(id, r, ColorBase, opt)
		// draw text
		text := formatNumber(*value, digits)
		c.drawControlText(text, r, ColorText, opt)

		return res
	})
}

// header creates and manages a header control with an optional tree node state, label, and ID. Returns a Response.
func (c *Context) header(label string, idStr string, istreenode bool, opt option) Response {
	var id controlID
	if len(idStr) > 0 {
		id = c.pushID([]byte(idStr))
		defer c.popID()
	} else if len(label) > 0 {
		id = c.pushID([]byte(label))
		defer c.popID()
	}

	idx := c.poolGet(c.treeNodePool[:], id)
	c.SetLayoutRow([]int{-1}, 0)

	active := idx >= 0
	var expanded bool
	if (opt & optionExpanded) != 0 {
		expanded = !active
	} else {
		expanded = active
	}

	return c.control(id, 0, func(r image.Rectangle) Response {
		// handle click (TODO (port): check if this is correct)
		clicked := c.mousePressed == mouseLeft && c.focus == id
		v1, v2 := 0, 0
		if active {
			v1 = 1
		}
		if clicked {
			v2 = 1
		}
		active = (v1 ^ v2) == 1

		// update pool ref
		if idx >= 0 {
			if active {
				c.poolUpdate(c.treeNodePool[:], idx)
			} else {
				c.treeNodePool[idx] = poolItem{}
			}
		} else if active {
			c.poolInit(c.treeNodePool[:], id)
		}

		// draw
		if istreenode {
			if c.hover == id {
				c.drawFrame(r, ColorButtonHover)
			}
		} else {
			c.drawControlFrame(id, r, ColorButton, 0)
		}
		var icon icon
		if expanded {
			icon = iconExpanded
		} else {
			icon = iconCollapsed
		}
		c.drawIcon(
			icon,
			image.Rect(r.Min.X, r.Min.Y, r.Min.X+r.Dy(), r.Max.Y),
			c.style.colors[ColorText],
		)
		r.Min.X += r.Dy() - c.style.padding
		c.drawControlText(label, r, ColorText, 0)

		if expanded {
			return ResponseActive
		}
		return 0
	})
}

// treeNode is a helper method to handle tree node operations with given label, id, options, and response function.
func (c *Context) treeNode(label string, idStr string, opt option, f func(res Response)) {
	res := c.header(label, idStr, true, opt)
	if res&ResponseActive == 0 {
		return
	}
	c.layout().indent += c.style.indent
	defer func() {
		c.layout().indent -= c.style.indent
	}()
	f(res)
}

// scrollbarVertical handles rendering and interaction for a vertical scrollbar within a specified container.
func (c *Context) scrollbarVertical(cnt *container, b image.Rectangle, cs image.Point) {
	maxscroll := cs.Y - b.Dy()
	if maxscroll > 0 && b.Dy() > 0 {
		// get sizing / positioning
		base := b
		base.Min.X = b.Max.X
		base.Max.X = base.Min.X + c.style.scrollbarSize

		// handle input
		id := c.idFromBytes([]byte("!scrollbar" + "y"))
		c.updateControl(id, base, 0)
		if c.focus == id && c.mouseDown == mouseLeft {
			cnt.layout.Scroll.Y += c.mouseDelta.Y * cs.Y / base.Dy()
		}
		// clamp scroll to limits
		cnt.layout.Scroll.Y = clamp(cnt.layout.Scroll.Y, 0, maxscroll)

		// draw base and thumb
		c.drawFrame(base, ColorScrollBase)
		thumb := base
		thumb.Max.Y = thumb.Min.Y + max(c.style.thumbSize, base.Dy()*b.Dy()/cs.Y)
		thumb = thumb.Add(image.Pt(0, cnt.layout.Scroll.Y*(base.Dy()-thumb.Dy())/maxscroll))
		c.drawFrame(thumb, ColorScrollThumb)

		// set this as the scroll_target (will get scrolled on mousewheel)
		// if the mouse is over it
		if c.mouseOver(b) {
			c.scrollTarget = cnt
		}
	} else {
		cnt.layout.Scroll.Y = 0
	}
}

// scrollbarHorizontal draws and handles the horizontal scrollbar for a given container's layout.
// It calculates the scrollbar's position, dimensions, and movement based on container dimensions and input events.
// It updates the container's horizontal scroll offset and clamps it within valid limits.
// The scrollbar is visually represented with a base track and a movable thumb.
// If the mouse is over the container, it sets the container as the current scroll target.
func (c *Context) scrollbarHorizontal(cnt *container, b image.Rectangle, cs image.Point) {
	maxscroll := cs.X - b.Dx()
	if maxscroll > 0 && b.Dx() > 0 {
		// get sizing / positioning
		base := b
		base.Min.Y = b.Max.Y
		base.Max.Y = base.Min.Y + c.style.scrollbarSize

		// handle input
		id := c.idFromBytes([]byte("!scrollbar" + "x"))
		c.updateControl(id, base, 0)
		if c.focus == id && c.mouseDown == mouseLeft {
			cnt.layout.Scroll.X += c.mouseDelta.X * cs.X / base.Dx()
		}
		// clamp scroll to limits
		cnt.layout.Scroll.X = clamp(cnt.layout.Scroll.X, 0, maxscroll)

		// draw base and thumb
		c.drawFrame(base, ColorScrollBase)
		thumb := base
		thumb.Max.X = thumb.Min.X + max(c.style.thumbSize, base.Dx()*b.Dx()/cs.X)
		thumb = thumb.Add(image.Pt(cnt.layout.Scroll.X*(base.Dx()-thumb.Dx())/maxscroll, 0))
		c.drawFrame(thumb, ColorScrollThumb)

		// set this as the scroll_target (will get scrolled on mousewheel)
		// if the mouse is over it
		if c.mouseOver(b) {
			c.scrollTarget = cnt
		}
	} else {
		cnt.layout.Scroll.X = 0
	}
}

// scrollbar renders a horizontal or vertical scrollbar based on the swap parameter and container's layout dimensions.
func (c *Context) scrollbar(cnt *container, b image.Rectangle, cs image.Point, swap bool) {
	if swap {
		c.scrollbarHorizontal(cnt, b, cs)
	} else {
		c.scrollbarVertical(cnt, b, cs)
	}
}

// scrollbars adjusts the body rectangle dimensions to account for scrollbars and handles their creation for the container.
func (c *Context) scrollbars(cnt *container, body image.Rectangle) image.Rectangle {
	sz := c.style.scrollbarSize
	cs := cnt.layout.ContentSize
	cs.X += c.style.padding * 2
	cs.Y += c.style.padding * 2
	c.pushClipRect(body)
	// resize body to make room for scrollbars
	if cs.Y > cnt.layout.Body.Dy() {
		body.Max.X -= sz
	}
	if cs.X > cnt.layout.Body.Dx() {
		body.Max.Y -= sz
	}
	// to create a horizontal or vertical scrollbar almost-identical code is
	// used; only the references to `x|y` `w|h` need to be switched
	c.scrollbar(cnt, body, cs, false)
	c.scrollbar(cnt, body, cs, true)
	c.popClipRect()
	return body
}

// pushContainerBody adjusts the layout and scroll behavior of a container based on the given body rectangle and options.
func (c *Context) pushContainerBody(cnt *container, body image.Rectangle, opt option) {
	if (^opt & optionNoScroll) != 0 {
		body = c.scrollbars(cnt, body)
	}
	c.pushLayout(body.Inset(c.style.padding), cnt.layout.Scroll)
	cnt.layout.Body = body
}

// window creates and manages a container window with a title, optional close and resize functionality, and customizable layout.
// If idStr is provided, it uniquely identifies the window. Otherwise, the title is used for identification.
// The rect parameter defines the initial size and position of the window.
// opt specifies additional options such as disabling the frame, resize, or title, and enabling popup or auto-sizing functionality.
// f is a callback function that provides the response and layout details of the window to the caller.
func (c *Context) window(title string, idStr string, rect image.Rectangle, opt option, f func(res Response, layout Layout)) {
	var id controlID
	if len(idStr) > 0 {
		id = c.pushID([]byte(idStr))
		defer c.popID()
	} else if len(title) > 0 {
		id = c.pushID([]byte(title))
		defer c.popID()
	}

	cnt := c.container(id, opt)
	if cnt == nil || !cnt.open {
		return
	}
	// This is popped at endRootContainer.
	// TODO: This is tricky. Refactor this.

	if cnt.layout.Rect.Dx() == 0 {
		cnt.layout.Rect = rect
	}

	c.containerStack = append(c.containerStack, cnt)
	defer c.popContainer()

	// push container to roots list and push head command
	c.rootList = append(c.rootList, cnt)
	cnt.headIdx = c.pushJump(-1)
	defer func() {
		// push tail 'goto' jump command and set head 'skip' command. the final steps
		// on initing these are done in End
		cnt := c.currentContainer()
		cnt.tailIdx = c.pushJump(-1)
		c.commandList[cnt.headIdx].jump.dstIdx = len(c.commandList) //- 1
	}()

	// set as hover root if the mouse is overlapping this container and it has a
	// higher zindex than the current hover root
	if c.mousePos.In(cnt.layout.Rect) && (c.nextHoverRoot == nil || cnt.zIndex > c.nextHoverRoot.zIndex) {
		c.nextHoverRoot = cnt
	}

	// clipping is reset here in case a root-container is made within
	// another root-containers's begin/end block; this prevents the inner
	// root-container being clipped to the outer
	c.clipStack = append(c.clipStack, unclippedRect)
	defer c.popClipRect()

	body := cnt.layout.Rect
	rect = body

	// draw frame
	if (^opt & optionNoFrame) != 0 {
		c.drawFrame(rect, ColorWindowBG)
	}

	// do title bar
	if (^opt & optionNoTitle) != 0 {
		tr := rect
		tr.Max.Y = tr.Min.Y + c.style.titleHeight
		c.drawFrame(tr, ColorTitleBG)

		// do title text
		if (^opt & optionNoTitle) != 0 {
			id := c.idFromBytes([]byte("!title"))
			c.updateControl(id, tr, opt)
			c.drawControlText(title, tr, ColorTitleText, opt)
			if id == c.focus && c.mouseDown == mouseLeft {
				cnt.layout.Rect = cnt.layout.Rect.Add(c.mouseDelta)
			}
			body.Min.Y += tr.Dy()
		}

		// do `close` button
		if (^opt & optionNoClose) != 0 {
			id := c.idFromBytes([]byte("!close"))
			r := image.Rect(tr.Max.X-tr.Dy(), tr.Min.Y, tr.Max.X, tr.Max.Y)
			tr.Max.X -= r.Dx()
			c.drawIcon(iconClose, r, c.style.colors[ColorTitleText])
			c.updateControl(id, r, opt)
			if c.mousePressed == mouseLeft && id == c.focus {
				cnt.open = false
			}
		}
	}

	c.pushContainerBody(cnt, body, opt)

	// do `resize` handle
	if (^opt & optionNoResize) != 0 {
		sz := c.style.titleHeight
		id := c.idFromBytes([]byte("!resize"))
		r := image.Rect(rect.Max.X-sz, rect.Max.Y-sz, rect.Max.X, rect.Max.Y)
		c.updateControl(id, r, opt)
		if id == c.focus && c.mouseDown == mouseLeft {
			cnt.layout.Rect.Max.X = cnt.layout.Rect.Min.X + max(96, cnt.layout.Rect.Dx()+c.mouseDelta.X)
			cnt.layout.Rect.Max.Y = cnt.layout.Rect.Min.Y + max(64, cnt.layout.Rect.Dy()+c.mouseDelta.Y)
		}
	}

	// resize to content size
	if (opt & optionAutoSize) != 0 {
		r := c.layout().body
		cnt.layout.Rect.Max.X = cnt.layout.Rect.Min.X + cnt.layout.ContentSize.X + (cnt.layout.Rect.Dx() - r.Dx())
		cnt.layout.Rect.Max.Y = cnt.layout.Rect.Min.Y + cnt.layout.ContentSize.Y + (cnt.layout.Rect.Dy() - r.Dy())
	}

	// close if this is a popup window and elsewhere was clicked
	if (opt&optionPopup) != 0 && c.mousePressed != 0 && c.hoverRoot != cnt {
		cnt.open = false
	}

	c.pushClipRect(cnt.layout.Body)
	defer c.popClipRect()

	f(ResponseActive, c.currentContainer().layout)
}

// OpenPopup opens a popup with the specified name and positions it at the current mouse cursor location.
func (c *Context) OpenPopup(name string) {
	cnt := c.Container(name)
	// set as hover root so popup isn't closed in begin_window_ex()
	c.nextHoverRoot = cnt
	c.hoverRoot = c.nextHoverRoot
	// position at mouse cursor, open and bring-to-front
	cnt.layout.Rect = image.Rect(c.mousePos.X, c.mousePos.Y, c.mousePos.X+1, c.mousePos.Y+1)
	cnt.open = true
	c.bringToFront(cnt)
}

// Popup creates a modal popup window with the given name and callback function for rendering the content.
func (c *Context) Popup(name string, f func(res Response, layout Layout)) {
	opt := optionPopup | optionAutoSize | optionNoResize | optionNoScroll | optionNoTitle | optionClosed
	c.window(name, "", image.Rectangle{}, opt, f)
}

// panel sets up and manages a UI panel with the given layout, applying options and rendering content through a callback.
func (c *Context) panel(name string, opt option, f func(layout Layout)) {
	id := c.pushID([]byte(name))
	defer c.popID()

	cnt := c.container(id, opt)
	cnt.layout.Rect = c.layoutNext()
	if (^opt & optionNoFrame) != 0 {
		c.drawFrame(cnt.layout.Rect, ColorPanelBG)
	}

	c.containerStack = append(c.containerStack, cnt)
	c.pushContainerBody(cnt, cnt.layout.Rect, opt)
	defer c.popContainer()

	c.pushClipRect(cnt.layout.Body)
	defer c.popClipRect()

	f(c.currentContainer().layout)
}

// placeholder выделяет пустое пространство в layout без отрисовки.
func (c *Context) Placeholder() {
	c.control(0, 0, func(r image.Rectangle) Response {
		return 0
	})
}
