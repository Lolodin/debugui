package debugui

import "strings"

type container struct {
	layout  Layout
	headIdx int
	tailIdx int
	zIndex  int
	open    bool
}

func (c *container) SetOpen(state bool) {
	c.open = state
}
func (c *container) IsOpen() bool {
	return c.open
}

func (c *Context) WindowContainer(title string) *container {
	title, idStr, _ := strings.Cut(title, idSeparator)
	var id controlID
	if len(idStr) > 0 {
		id = c.pushID([]byte(idStr))
		defer c.popID()
	} else if len(title) > 0 {
		id = c.pushID([]byte(title))
		defer c.popID()
	}

	cnt := c.container(id, 0)
	return cnt
}
