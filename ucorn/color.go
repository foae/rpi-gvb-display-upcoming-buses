package ucorn

import (
	"fmt"
)

type Color struct {
	Red   byte
	Green byte
	Blue  byte
}

func ColorNew(red byte, green byte, blue byte) Color {
	return Color{Red: red, Green: green, Blue: blue}
}

func (c *Color) Bytes(mode byte) ([]byte) {
	if mode == MODE_RGB {
		return []byte{c.Red, c.Green, c.Blue}
	}

	return []byte{c.Green, c.Red, c.Blue}
}

func (c *Color) String() string {
	return fmt.Sprintf("#%x%x%x", c.Red, c.Green, c.Blue)
}
