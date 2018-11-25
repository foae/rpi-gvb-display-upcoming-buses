package unicorn

const (
	SocketPath = "/var/run/unicornd.socket"
)

var (
	Red    = Pixel{230, 0, 0}
	Orange = Pixel{230, 150, 0}
	Green  = Pixel{0, 255, 0}
	Blue   = Pixel{0, 0, 255}
	Cyan   = Pixel{0, 255, 255}
	Black  = Pixel{0, 0, 0}
	White  = Pixel{255, 230, 230}
)

// Pixel definition.
type Pixel struct {
	R uint `struc:"uint8"` //uint8_t r;
	G uint `struc:"uint8"` //uint8_t g;
	B uint `struc:"uint8"` //uint8_t b;
}
