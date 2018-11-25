package unicorn

// types pulled and made Go compatible from https://github.com/pimoroni/unicorn-hat/tree/master/library_c/unicornd

/*
UNICORND_CMD_SET_BRIGHTNESS = 0
UNICORND_CMD_SET_PIXEL      = 1
UNICORND_CMD_SET_ALL_PIXELS = 2
UNICORND_CMD_SHOW           = 3
*/

const (
	CMDSetBrightness uint = iota
	CMDSetPixel
	CMDSetAllPixels
	CMDShow
)

// Set pixel
type setPixel struct {
	Code uint `struc:"uint8"` //uint8_t code; // set to 1
	Pos  position             //pos_t pos;
	Col  Pixel                //col_t col;
}

//where pos_t is a struct like this:
type position struct {
	X uint `struc:"uint8"` //uint8_t x;
	Y uint `struc:"uint8"` //uint8_t y;
}

// brightness
type brightness struct {
	Code uint `struc:"uint8"` //uint8_t code; // set to 0
	Val  uint `struc:"uint8"` //double  val;
}

// Set all pixels
type setAll struct {
	Code   uint `struc:"uint8"` //uint8_t code; // set to 2
	Pixels [64]Pixel            //col_t pixels[64];
}

// Show
type show struct {
	Code uint `struc:"uint8"` //uint8_t code; // set to 3
}
