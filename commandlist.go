// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package microui

import (
	"image"
	"image/color"
)

// pushCommand adds a new command with type cmd_type to command_list
func (c *Context) pushCommand(cmd_type int) *command {
	cmd := command{
		typ: cmd_type,
	}
	//expect(uintptr(len(ctx.CommandList))*size+size < MU_COMMANDLIST_SIZE)
	cmd.base.typ = cmd_type
	cmd.idx = len(c.commandList)
	c.commandList = append(c.commandList, &cmd)
	return &cmd
}

func (c *Context) nextCommand(cmd **command) bool {
	if len(c.commandList) == 0 {
		return false
	}
	if *cmd == nil {
		*cmd = c.commandList[0]
	} else {
		*cmd = c.commandList[(*cmd).idx+1]
	}

	for (*cmd).idx < len(c.commandList) {
		if (*cmd).typ != commandJump {
			return true
		}
		idx := (*cmd).jump.dstIdx
		if idx > len(c.commandList)-1 {
			break
		}
		*cmd = c.commandList[idx]
	}
	return false
}

// pushJump pushes a new jump command to command_list
func (c *Context) pushJump(dstIdx int) int {
	cmd := c.pushCommand(commandJump)
	cmd.jump.dstIdx = dstIdx
	return len(c.commandList) - 1
}

// SetClip pushes a new clip command
func (c *Context) SetClip(rect image.Rectangle) {
	cmd := c.pushCommand(commandClip)
	cmd.clip.rect = rect
}

// DrawRect pushes a new rect command
func (c *Context) DrawRect(rect image.Rectangle, color color.Color) {
	rect2 := rect.Intersect(c.clipRect())
	if rect2.Dx() > 0 && rect2.Dy() > 0 {
		cmd := c.pushCommand(commandRect)
		cmd.rect.rect = rect2
		cmd.rect.color = color
	}
}

func (c *Context) DrawBox(rect image.Rectangle, color color.Color) {
	c.DrawRect(image.Rect(rect.Min.X+1, rect.Min.Y, rect.Max.X-1, rect.Min.Y+1), color)
	c.DrawRect(image.Rect(rect.Min.X+1, rect.Max.Y-1, rect.Max.X-1, rect.Max.Y), color)
	c.DrawRect(image.Rect(rect.Min.X, rect.Min.Y, rect.Min.X+1, rect.Max.Y), color)
	c.DrawRect(image.Rect(rect.Max.X-1, rect.Min.Y, rect.Max.X, rect.Max.Y), color)
}

func (c *Context) DrawText(str string, pos image.Point, color color.Color) {
	rect := image.Rect(pos.X, pos.Y, pos.X+textWidth(str), pos.Y+textHeight())
	clipped := c.checkClip(rect)
	if clipped == clipAll {
		return
	}
	if clipped == clipPart {
		c.SetClip(c.clipRect())
	}
	// add command
	cmd := c.pushCommand(commandText)
	cmd.text.str = str
	cmd.text.pos = pos
	cmd.text.color = color
	// reset clipping if it was set
	if clipped != 0 {
		c.SetClip(unclippedRect)
	}
}

func (c *Context) DrawIcon(icon Icon, rect image.Rectangle, color color.Color) {
	// do clip command if the rect isn't fully contained within the cliprect
	clipped := c.checkClip(rect)
	if clipped == clipAll {
		return
	}
	if clipped == clipPart {
		c.SetClip(c.clipRect())
	}
	// do icon command
	cmd := c.pushCommand(commandIcon)
	cmd.icon.icon = icon
	cmd.icon.rect = rect
	cmd.icon.color = color
	// reset clipping if it was set
	if clipped != 0 {
		c.SetClip(unclippedRect)
	}
}
