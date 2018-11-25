package unicorn

import (
	"fmt"
	"github.com/lunixbochs/struc"
	"net"
)

// Client is the unix socket client for `unicornd`.
type Client struct {
	Path    string
	sock    net.Conn
	verbose bool
}

// NewClient returns a new unicorn Client.
// If no socket is specified, it will use a default. Omit if unsure.
func NewClient(verbose bool, socketPath string) *Client {
	if socketPath == "" {
		socketPath = SocketPath
	}

	return &Client{
		Path:    socketPath,
		verbose: verbose,
	}
}

// Connect opens a connection to the Client.Path
func (c *Client) Connect() (err error) {
	if c.verbose {
		fmt.Printf("Connecting to (%v)...\n", SocketPath)
	}
	c.sock, err = net.Dial("unix", SocketPath)

	return err
}

// SetBrightness of the display, 0..255
func (c Client) SetBrightness(v uint) error {
	if c.verbose {
		fmt.Printf("Setting brightness to (%v)\n", v)
	}

	if v > 255 {
		return fmt.Errorf("brightness must be 0..255, passed: %v", v)
	}

	b := brightness{
		Code: CMDSetBrightness,
		Val:  v,
	}

	return struc.Pack(c.sock, &b)
}

// SetPixel sets the color of an individual pixel.
func (c Client) SetPixel(x, y, r, g, b uint) error {
	if c.verbose {
		fmt.Printf("Setting pixel `x:%v`, `y:%v` to rgb(%v, %v, %v)\n", x, y, r, g, b)
	}

	switch {
	case x > 7 || y > 7:
		return fmt.Errorf("`x`, `y` must be 0..7, passed `x: %v`, `y: %v`", x, y)
	case r > 255 || g > 255 || b > 255:
		return fmt.Errorf("`r`, `g`, `b` must be 0..255, passed `r:%v`, `g:%v`, `b:%v`", r, g, b)
	}

	sp := setPixel{
		Code: CMDSetPixel,
		Pos: position{
			X: x,
			Y: y,
		},
		Col: Pixel{
			R: g, // due to a bug in https://github.com/pimoroni/unicorn-hat/blob/master/library_c/unicornd/unicornd.c
			G: r,
			B: b,
		},
	}

	//buf := bytes.NewBuffer([]byte{})
	//buf.Write([]byte{byte(sp.Code)})
	//buf.Write([]byte{byte(sp.Pos.X), byte(sp.Pos.Y)})
	//buf.Write([]byte{byte(sp.Col.R), byte(sp.Col.G), byte(sp.Col.B)})
	//bw, err := c.sock.Write(buf.Bytes())
	//return err

	return struc.Pack(c.sock, &sp)
}

// SetAllPixels allows you to set the color of all pixels in a single call.
func (c Client) SetAllPixels(ps [64]Pixel) error {
	if c.verbose {
		fmt.Println("Setting all pixels...")
	}

	for i := range ps {
		if ps[i].R > 255 || ps[i].G > 255 || ps[i].B > 255 {
			return fmt.Errorf("pixel out of range; `r`, `g`, `b` must be 0..255, passed: `r:%v`, `g:%v`, `b:%v`",
				ps[i].R, ps[i].G, ps[i].B,
			)
		}
	}

	sp := setAll{
		Code:   CMDSetAllPixels,
		Pixels: ps,
	}
	return struc.Pack(c.sock, &sp)
}

// Show the pixels written to the buffer.
func (c Client) Show() error {
	if c.verbose {
		fmt.Println("Displaying pixels")
	}

	s := show{
		Code: CMDShow,
	}

	return struc.Pack(c.sock, &s)
}

// Clear sets all pixels to 0,0,0.
// wrapper for Client.SetAllPixels([64]Pixel{})
func (c Client) Clear() error {
	if c.verbose {
		fmt.Println("Display cleared")
	}
	c.SetAllPixels([64]Pixel{})
	c.Show()

	return nil
}

// Silent suppresses writing to stdout.
func (c *Client) Silent() {
	c.verbose = false
}

// Verbose shows the true therapeutic nature of this program.
func (c *Client) Verbose() {
	c.verbose = true
}
