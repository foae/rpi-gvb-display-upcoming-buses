package unicorn

import "math"

// Matrix is an 8x8 matrix of unicorn.Pixels
type Matrix [8][8]Pixel

// Supersample is a 128x128 matrix of Pixels, used for smoother shapes and antialiasing
type Supersample [128][128]Pixel

// Circle creates a circle of a given radius, offset, and color
func Circle(r int, o [2]int, c Pixel) Supersample {
	s := Supersample{}
	for j := r; j > 0; j-- {
		for i := 0; i < 360; i++ {
			xr := 64 + float64(o[0]) + float64(j)*math.Cos(float64(i)*math.Pi/180)
			yr := 64 + float64(o[1]) + float64(j)*math.Sin(float64(i)*math.Pi/180)
			if xr > 127 {
				xr = 127
			}
			if xr < 0 {
				xr = 0
			}
			if yr > 127 {
				yr = 127
			}
			if yr < 0 {
				yr = 0
			}
			s[int(xr)][int(yr)] = c
		}
	}
	return s
}

// DeMatrix converts an 8x8 Pixel grid into a [64]Pixel
// for use by Client.SetAllPixels.
func DeMatrix(m [8][8]Pixel) [64]Pixel {
	ps := [64]Pixel{}
	n := 0
	for i := range m {
		for j := range m[i] {
			ps[n] = m[i][j]
			n++
		}
	}
	return ps

}

// MapSupersample draws a supersample to the matrix,
// setting pixels to the absolute values in the Supersample.
func (m *Matrix) MapSupersample(s Supersample) {
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			var r, g, b uint
			for k := 0; k < 16; k++ {
				for l := 0; l < 16; l++ {
					r = r + s[i*16+k][j*16+l].R
					g = g + s[i*16+k][j*16+l].G
					b = b + s[i*16+k][j*16+l].B
				}
			}
			r = r / (16 * 16)
			g = g / (16 * 16)
			b = b / (16 * 16)
			if r > 255 {
				r = 255
			}
			if g > 255 {
				g = 255
			}
			if b > 255 {
				b = 255
			}
			m[i][j].R = r
			m[i][j].G = g
			m[i][j].B = b
		}
	}
}

// AddSupersample draws a supersample to the matrix,
// adding the values of the supersample to existing Pixels.
func (m *Matrix) AddSupersample(s Supersample) {
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			var r, g, b uint
			for k := 0; k < 16; k++ {
				for l := 0; l < 16; l++ {
					r = r + s[i*16+k][j*16+l].R
					g = g + s[i*16+k][j*16+l].G
					b = b + s[i*16+k][j*16+l].B
				}
			}
			r = r / (16 * 16)
			g = g / (16 * 16)
			b = b / (16 * 16)
			m[i][j].R += r
			m[i][j].G += g
			m[i][j].B += b
			if m[i][j].R > 255 {
				m[i][j].R = 255
			}
			if m[i][j].G > 255 {
				m[i][j].G = 255
			}
			if m[i][j].B > 255 {
				m[i][j].B = 255
			}
		}
	}
}
