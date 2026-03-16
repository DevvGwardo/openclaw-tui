package tui

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"
)

// BgMode identifies a background animation mode.
type BgMode string

const (
	BgOff       BgMode = "off"
	BgStarfield BgMode = "starfield"
	BgTunnel    BgMode = "tunnel"
	BgPlasma    BgMode = "plasma"
	BgFire      BgMode = "fire"
	BgMatrix    BgMode = "matrix"
	BgOcean     BgMode = "ocean"
	BgCube      BgMode = "cube"
)

// BgModes lists all available background modes in cycle order.
var BgModes = []BgMode{BgOff, BgStarfield, BgTunnel, BgPlasma, BgFire, BgMatrix, BgOcean, BgCube}

// --- Pixel buffer for half-block rendering ---

type pixel struct {
	r, g, b uint8
}

type pixelBuffer struct {
	width, height int // terminal dimensions
	pixels        []pixel
}

func newPixelBuffer(w, h int) *pixelBuffer {
	if w <= 0 || h <= 0 {
		return &pixelBuffer{width: 0, height: 0}
	}
	return &pixelBuffer{
		width:  w,
		height: h,
		pixels: make([]pixel, w*h*2), // double vertical resolution
	}
}

func (pb *pixelBuffer) set(x, py int, r, g, b uint8) {
	if x < 0 || x >= pb.width || py < 0 || py >= pb.height*2 {
		return
	}
	idx := py*pb.width + x
	pb.pixels[idx] = pixel{r, g, b}
}

func (pb *pixelBuffer) get(x, py int) pixel {
	if x < 0 || x >= pb.width || py < 0 || py >= pb.height*2 {
		return pixel{}
	}
	return pb.pixels[py*pb.width+x]
}

func (pb *pixelBuffer) clear() {
	for i := range pb.pixels {
		pb.pixels[i] = pixel{}
	}
}

// renderRow renders terminal row y using half-block characters.
// Top pixel = pixels[y*2], bottom pixel = pixels[y*2+1], character = ▀
func (pb *pixelBuffer) renderRow(y int) string {
	if pb.width <= 0 || y < 0 || y >= pb.height {
		return ""
	}
	var sb strings.Builder
	sb.Grow(pb.width * 40)

	topBase := y * 2 * pb.width
	botBase := (y*2 + 1) * pb.width

	// Batch consecutive identical color pairs
	prevTop := pixel{255, 255, 255} // impossible start
	prevBot := pixel{255, 255, 255}
	run := 0

	flush := func() {
		if run <= 0 {
			return
		}
		if prevTop.r == 0 && prevTop.g == 0 && prevTop.b == 0 &&
			prevBot.r == 0 && prevBot.g == 0 && prevBot.b == 0 {
			sb.WriteString(strings.Repeat(" ", run))
		} else {
			fmt.Fprintf(&sb, "\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm",
				prevTop.r, prevTop.g, prevTop.b,
				prevBot.r, prevBot.g, prevBot.b)
			for k := 0; k < run; k++ {
				sb.WriteRune('▀')
			}
			sb.WriteString("\x1b[0m")
		}
		run = 0
	}

	for x := 0; x < pb.width; x++ {
		top := pb.pixels[topBase+x]
		bot := pb.pixels[botBase+x]
		if top == prevTop && bot == prevBot {
			run++
		} else {
			flush()
			prevTop = top
			prevBot = bot
			run = 1
		}
	}
	flush()
	return sb.String()
}

// bgColorAt returns the average color of the two pixel rows for a terminal cell.
func (pb *pixelBuffer) bgColorAt(row, col int) (int, int, int) {
	if col < 0 || col >= pb.width || row < 0 || row >= pb.height {
		return 6, 6, 10
	}
	top := pb.get(col, row*2)
	bot := pb.get(col, row*2+1)
	r := (int(top.r) + int(bot.r)) / 2
	g := (int(top.g) + int(bot.g)) / 2
	b := (int(top.b) + int(bot.b)) / 2
	if r == 0 && g == 0 && b == 0 {
		return 6, 6, 10
	}
	return r, g, b
}

// --- State types ---

type star3d struct {
	x, y, z    float64
	prevSX     float64
	prevSY     float64
	colorTint  int // 0=white, 1=blue, 2=yellow
}

type matrixCol struct {
	y         float64
	speed     float64
	length    int
	active    bool
	chars     []rune
	spotlight bool
	spotTimer int
}

type cubeState struct {
	angleX, angleY float64
	history        [][8][2]float64 // projected vertex history for trails
}

// bgCell represents a single background cell for character-based effects (matrix).
type bgCell struct {
	ch rune
	fg [3]uint8
}

// BackgroundModel manages animated background effects.
type BackgroundModel struct {
	mode   BgMode
	theme  Theme
	width  int
	height int
	frame  int

	// Pixel buffer for half-block modes
	pb *pixelBuffer

	// Starfield state
	stars []star3d

	// Fire state
	fireHeat []float64 // width * (height*2) heat values

	// Matrix state (character-based, not pixel)
	matrixCols []matrixCol

	// Cube state
	cube cubeState

	// Pre-built character grid for character-based effects (matrix)
	charGrid map[int]map[int]bgCell

	// Pre-computed sine table
	sinTable [1024]float64

	rng *rand.Rand
}

// NewBackgroundModel creates a new background renderer.
func NewBackgroundModel(theme Theme) BackgroundModel {
	m := BackgroundModel{
		mode:     BgOff,
		theme:    theme,
		charGrid: make(map[int]map[int]bgCell),
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	for i := range m.sinTable {
		m.sinTable[i] = math.Sin(float64(i) * 2.0 * math.Pi / 1024.0)
	}
	return m
}

// fastSin returns sin using the pre-computed table.
func (b *BackgroundModel) fastSin(v float64) float64 {
	// Normalize to 0..1024
	idx := int(v*1024.0/(2.0*math.Pi)) % 1024
	if idx < 0 {
		idx += 1024
	}
	return b.sinTable[idx]
}

// fastCos returns cos using the pre-computed table.
func (b *BackgroundModel) fastCos(v float64) float64 {
	return b.fastSin(v + math.Pi/2.0)
}

// SetMode sets the background animation mode.
func (b *BackgroundModel) SetMode(mode BgMode) {
	b.mode = mode
	b.frame = 0
	b.charGrid = make(map[int]map[int]bgCell)
	b.initMode()
}

// CycleMode advances to the next background mode and returns it.
func (b *BackgroundModel) CycleMode() BgMode {
	current := 0
	for i, m := range BgModes {
		if m == b.mode {
			current = i
			break
		}
	}
	next := (current + 1) % len(BgModes)
	b.SetMode(BgModes[next])
	return b.mode
}

// SetTheme updates the theme used for background colors.
func (b *BackgroundModel) SetTheme(theme Theme) {
	b.theme = theme
}

// SetSize updates the terminal dimensions.
func (b *BackgroundModel) SetSize(w, h int) {
	if w == b.width && h == b.height {
		return
	}
	b.width = w
	b.height = h
	b.initMode()
}

// IsActive returns true if a background animation is running.
func (b *BackgroundModel) IsActive() bool {
	return b.mode != BgOff
}

// Tick advances the animation by one frame.
func (b *BackgroundModel) Tick() {
	if b.mode == BgOff {
		return
	}
	b.frame++
	b.updateAnimation()
}

func (b *BackgroundModel) ensurePixelBuffer() {
	if b.pb == nil || b.pb.width != b.width || b.pb.height != b.height {
		b.pb = newPixelBuffer(b.width, b.height)
	}
}

func (b *BackgroundModel) initMode() {
	b.ensurePixelBuffer()
	switch b.mode {
	case BgStarfield:
		b.initStarfield()
	case BgTunnel:
		// no extra init needed
	case BgPlasma:
		// no extra init needed
	case BgFire:
		b.initFire()
	case BgMatrix:
		b.initMatrix()
	case BgOcean:
		// no extra init needed
	case BgCube:
		b.initCube()
	}
}

// --- Theme color helpers ---

func (b *BackgroundModel) themeRGB(which string) (uint8, uint8, uint8) {
	p := b.theme.Palette
	var hex string
	switch which {
	case "primary":
		hex = string(p.Primary)
	case "secondary":
		hex = string(p.Secondary)
	case "accent":
		hex = string(p.Accent)
	case "success":
		hex = string(p.Success)
	case "warning":
		hex = string(p.Warning)
	default:
		hex = string(p.Primary)
	}
	r, g, bv := hexToRGB(hex)
	return uint8(r), uint8(g), uint8(bv)
}

// lerpColor linearly interpolates between two colors.
func lerpColor(r1, g1, b1, r2, g2, b2 uint8, t float64) (uint8, uint8, uint8) {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return uint8(float64(r1)*(1-t) + float64(r2)*t),
		uint8(float64(g1)*(1-t) + float64(g2)*t),
		uint8(float64(b1)*(1-t) + float64(b2)*t)
}

// dimColor scales a color to a fraction.
func dimColor(r, g, b uint8, scale float64) (uint8, uint8, uint8) {
	return uint8(float64(r) * scale), uint8(float64(g) * scale), uint8(float64(b) * scale)
}

// --- Starfield (3D warp speed) ---

func (b *BackgroundModel) initStarfield() {
	if b.width == 0 || b.height == 0 {
		return
	}
	count := 250
	b.stars = make([]star3d, count)
	for i := range b.stars {
		b.stars[i] = b.newStar3d(true)
	}
}

func (b *BackgroundModel) newStar3d(randomZ bool) star3d {
	z := 1.0
	if randomZ {
		z = b.rng.Float64()*0.95 + 0.05
	}
	tint := 0
	r := b.rng.Float64()
	if r < 0.08 {
		tint = 1 // blue
	} else if r < 0.14 {
		tint = 2 // yellow
	}
	return star3d{
		x:         (b.rng.Float64() - 0.5) * 2.5,
		y:         (b.rng.Float64() - 0.5) * 2.5,
		z:         z,
		prevSX:    -1,
		prevSY:    -1,
		colorTint: tint,
	}
}

func (b *BackgroundModel) updateStarfield() {
	b.ensurePixelBuffer()
	b.pb.clear()

	cx := float64(b.width) / 2.0
	cy := float64(b.height) // pixel-y center (double res)

	for i := range b.stars {
		s := &b.stars[i]

		// Project current position before moving
		sx := cx + (s.x/s.z)*cx
		sy := cy + (s.y/s.z)*cy

		// Store previous screen position for trails
		prevSX := s.prevSX
		prevSY := s.prevSY
		s.prevSX = sx
		s.prevSY = sy

		// Move star toward viewer
		s.z -= 0.018
		if s.z <= 0.005 {
			b.stars[i] = b.newStar3d(false)
			continue
		}

		// Check bounds
		if sx < -2 || sx >= float64(b.width)+2 || sy < -2 || sy >= float64(b.height*2)+2 {
			b.stars[i] = b.newStar3d(false)
			continue
		}

		// Brightness increases exponentially as z approaches 0
		brightness := math.Pow(1.0-s.z, 3.0)
		if brightness > 1.0 {
			brightness = 1.0
		}

		// Base color
		var cr, cg, cb float64
		switch s.colorTint {
		case 1: // blue tint
			cr, cg, cb = 0.6, 0.7, 1.0
		case 2: // yellow tint
			cr, cg, cb = 1.0, 0.95, 0.6
		default: // white
			cr, cg, cb = 1.0, 1.0, 1.0
		}

		// Dim to keep text readable
		maxBright := 0.55
		br := brightness * maxBright
		r := uint8(cr * br * 255)
		g := uint8(cg * br * 255)
		bv := uint8(cb * br * 255)

		// Draw star
		ix := int(sx)
		iy := int(sy)
		b.pb.set(ix, iy, r, g, bv)

		// Close stars are bigger (2x2 pixel block)
		if s.z < 0.2 {
			b.pb.set(ix+1, iy, r, g, bv)
			b.pb.set(ix, iy+1, r, g, bv)
			b.pb.set(ix+1, iy+1, r, g, bv)
		}

		// Motion trail (draw line from previous to current position)
		if prevSX >= 0 && prevSY >= 0 && s.z < 0.6 {
			trailLen := 3
			if s.z < 0.15 {
				trailLen = 5
			}
			for t := 1; t <= trailLen; t++ {
				frac := float64(t) / float64(trailLen+1)
				tx := int(sx + (prevSX-sx)*frac)
				ty := int(sy + (prevSY-sy)*frac)
				fade := (1.0 - frac) * br * 0.5
				tr := uint8(cr * fade * 255)
				tg := uint8(cg * fade * 255)
				tb := uint8(cb * fade * 255)
				b.pb.set(tx, ty, tr, tg, tb)
			}
		}
	}

	// Speed lines at edges
	if b.frame%3 == 0 {
		edgeW := b.width / 8
		for k := 0; k < 4; k++ {
			x := b.rng.Intn(edgeW)
			if b.rng.Intn(2) == 0 {
				x = b.width - 1 - x
			}
			py := b.rng.Intn(b.height * 2)
			for j := 0; j < 3+b.rng.Intn(4); j++ {
				fade := 0.15 * (1.0 - float64(j)/7.0)
				v := uint8(fade * 255)
				b.pb.set(x, py+j, v, v, v)
			}
		}
	}
}

// --- Tunnel (3D wormhole) ---

func (b *BackgroundModel) updateTunnel() {
	b.ensurePixelBuffer()
	b.pb.clear()

	if b.width == 0 || b.height == 0 {
		return
	}

	pr, pg, ppb := b.themeRGB("primary")
	sr, sg, sb := b.themeRGB("secondary")

	t := float64(b.frame) * 0.04
	cx := float64(b.width) / 2.0
	cy := float64(b.height) // pixel-y center

	// Wobble the center
	wobX := b.fastSin(t*0.7) * float64(b.width) * 0.05
	wobY := b.fastSin(t*0.5) * float64(b.height) * 0.08

	for py := 0; py < b.height*2; py++ {
		for x := 0; x < b.width; x++ {
			dx := float64(x) - cx - wobX
			dy := float64(py) - cy - wobY

			// Distance from center
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist < 0.5 {
				dist = 0.5
			}

			// Angle
			angle := math.Atan2(dy, dx)

			// Tunnel mapping: depth from distance
			depth := 80.0 / dist

			// Tunnel texture coordinates
			u := angle/(2.0*math.Pi) + 0.5
			v := depth + t*2.0

			// Undulating radius
			undulate := 1.0 + 0.15*b.fastSin(angle*3.0+t*1.5)
			v *= undulate

			// Ring pattern: alternating bands
			ring := b.fastSin(v * 4.0 * math.Pi)

			// Brightness based on depth (closer = brighter)
			depthBright := 1.0 / (1.0 + depth*0.15)

			// Color: interpolate primary (near) to secondary (far) based on depth
			depthFrac := depth / 5.0
			if depthFrac > 1 {
				depthFrac = 1
			}

			var cr, cg, cb uint8
			if ring > 0 {
				// Light ring - theme colored
				intensity := ring * depthBright * 0.45
				cr, cg, cb = lerpColor(pr, pg, ppb, sr, sg, sb, depthFrac)
				cr, cg, cb = dimColor(cr, cg, cb, intensity)
			} else {
				// Dark ring
				intensity := (1.0 + ring*0.5) * depthBright * 0.12
				cr = uint8(intensity * 40)
				cg = uint8(intensity * 20)
				cb = uint8(intensity * 60)
			}

			// Checkerboard overlay
			checker := b.fastSin(u*16.0*math.Pi) * b.fastSin(v*4.0*math.Pi)
			if checker > 0 {
				cr = uint8(minInt(255, int(cr)+8))
				cg = uint8(minInt(255, int(cg)+4))
				cb = uint8(minInt(255, int(cb)+8))
			}

			b.pb.set(x, py, cr, cg, cb)
		}
	}
}

// --- Plasma (classic demoscene) ---

func (b *BackgroundModel) updatePlasma() {
	b.ensurePixelBuffer()
	b.pb.clear()

	if b.width == 0 || b.height == 0 {
		return
	}

	pr, pg, ppb := b.themeRGB("primary")
	sr, sg, sb := b.themeRGB("secondary")
	ar, ag, ab := b.themeRGB("accent")

	t := float64(b.frame) * 0.03

	for py := 0; py < b.height*2; py++ {
		fy := float64(py) / float64(b.height*2)
		for x := 0; x < b.width; x++ {
			fx := float64(x) / float64(b.width)

			// Multiple overlapping sine waves at different frequencies
			v1 := b.fastSin(fx*6.0*math.Pi + t)
			v2 := b.fastSin(fy*8.0*math.Pi + t*1.3)
			v3 := b.fastSin((fx+fy)*5.0*math.Pi + t*0.7)
			v4 := b.fastSin(math.Sqrt((fx-0.5)*(fx-0.5)*16+(fy-0.5)*(fy-0.5)*16)*4.0*math.Pi + t*1.1)
			v5 := b.fastSin(fx*3.0*math.Pi+b.fastSin(fy*4.0*math.Pi+t)*2.0+t*0.5)

			val := (v1 + v2 + v3 + v4 + v5) / 5.0 // -1..1
			val = (val + 1.0) / 2.0                  // 0..1

			// Map to theme palette: primary → secondary → accent → primary
			var cr, cg, cb uint8
			maxBright := 0.40
			if val < 0.33 {
				frac := val / 0.33
				cr, cg, cb = lerpColor(pr, pg, ppb, sr, sg, sb, frac)
			} else if val < 0.66 {
				frac := (val - 0.33) / 0.33
				cr, cg, cb = lerpColor(sr, sg, sb, ar, ag, ab, frac)
			} else {
				frac := (val - 0.66) / 0.34
				cr, cg, cb = lerpColor(ar, ag, ab, pr, pg, ppb, frac)
			}
			cr, cg, cb = dimColor(cr, cg, cb, maxBright)

			b.pb.set(x, py, cr, cg, cb)
		}
	}
}

// --- Fire (Doom-style) ---

func (b *BackgroundModel) initFire() {
	if b.width == 0 || b.height == 0 {
		return
	}
	pH := b.height * 2
	b.fireHeat = make([]float64, b.width*pH)
	// Ignite bottom row
	for x := 0; x < b.width; x++ {
		b.fireHeat[(pH-1)*b.width+x] = 1.0
	}
}

// fireColor maps a heat value (0..1) to RGB.
func fireColor(heat float64) (uint8, uint8, uint8) {
	if heat < 0 {
		heat = 0
	}
	if heat > 1 {
		heat = 1
	}
	// black → dark red → red → orange → yellow → white
	// Dimmed overall to keep text readable
	dim := 0.50
	if heat < 0.15 {
		t := heat / 0.15
		return uint8(40 * t * dim), 0, 0
	} else if heat < 0.35 {
		t := (heat - 0.15) / 0.20
		return uint8((40 + 140*t) * dim), uint8(20 * t * dim), 0
	} else if heat < 0.55 {
		t := (heat - 0.35) / 0.20
		return uint8((180 + 50*t) * dim), uint8((20 + 80*t) * dim), 0
	} else if heat < 0.75 {
		t := (heat - 0.55) / 0.20
		return uint8((230 + 25*t) * dim), uint8((100 + 100*t) * dim), uint8(20 * t * dim)
	} else {
		t := (heat - 0.75) / 0.25
		return uint8((255) * dim), uint8((200 + 55*t) * dim), uint8((20 + 180*t) * dim)
	}
}

func (b *BackgroundModel) updateFire() {
	b.ensurePixelBuffer()

	if b.width == 0 || b.height == 0 {
		return
	}

	pH := b.height * 2
	w := b.width

	// Ensure heat buffer size matches
	if len(b.fireHeat) != w*pH {
		b.initFire()
	}

	// Set bottom row to hot with random variation
	for x := 0; x < w; x++ {
		b.fireHeat[(pH-1)*w+x] = 0.7 + b.rng.Float64()*0.3
	}

	// Random hotspots at bottom
	for k := 0; k < w/8; k++ {
		x := b.rng.Intn(w)
		b.fireHeat[(pH-1)*w+x] = 1.0
		if x > 0 {
			b.fireHeat[(pH-1)*w+x-1] = 0.95
		}
		if x < w-1 {
			b.fireHeat[(pH-1)*w+x+1] = 0.95
		}
	}

	// Propagate fire upward: each pixel = avg(below-left, below, below-right, 2-below) - random cooling
	for py := 0; py < pH-1; py++ {
		for x := 0; x < w; x++ {
			below := py + 1
			below2 := py + 2
			if below2 >= pH {
				below2 = pH - 1
			}

			left := x - 1
			if left < 0 {
				left = 0
			}
			right := x + 1
			if right >= w {
				right = w - 1
			}

			avg := (b.fireHeat[below*w+left] +
				b.fireHeat[below*w+x] +
				b.fireHeat[below*w+right] +
				b.fireHeat[below2*w+x]) / 4.0

			// Cooling factor increases toward the top
			coolRate := 0.012 + 0.006*b.rng.Float64()
			avg -= coolRate
			if avg < 0 {
				avg = 0
			}
			b.fireHeat[py*w+x] = avg
		}
	}

	// Render to pixel buffer
	b.pb.clear()
	for py := 0; py < pH; py++ {
		for x := 0; x < w; x++ {
			heat := b.fireHeat[py*w+x]
			if heat > 0.01 {
				r, g, bv := fireColor(heat)
				b.pb.set(x, py, r, g, bv)
			}
		}
	}
}

// --- Matrix (enhanced, character-based with half-block awareness) ---

var matrixChars = []rune("ｦｧｨｩｪｫｬｭｮｯｱｲｳｴｵｶｷｸｹｺABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func (b *BackgroundModel) initMatrix() {
	if b.width == 0 {
		return
	}
	b.matrixCols = make([]matrixCol, b.width)
	for i := range b.matrixCols {
		b.matrixCols[i] = matrixCol{
			y:      float64(-b.rng.Intn(b.height + 10)),
			speed:  0.2 + b.rng.Float64()*1.0,
			length: 5 + b.rng.Intn(20),
			active: b.rng.Float64() < 0.4,
			chars:  b.randomMatrixChars(25),
		}
	}
}

func (b *BackgroundModel) randomMatrixChars(n int) []rune {
	chars := make([]rune, n)
	for i := range chars {
		chars[i] = matrixChars[b.rng.Intn(len(matrixChars))]
	}
	return chars
}

func (b *BackgroundModel) updateMatrix() {
	b.charGrid = make(map[int]map[int]bgCell)

	for i := range b.matrixCols {
		col := &b.matrixCols[i]
		if !col.active {
			if b.rng.Float64() < 0.015 {
				col.active = true
				col.y = 0
				col.speed = 0.2 + b.rng.Float64()*1.0
				col.length = 5 + b.rng.Intn(20)
				col.chars = b.randomMatrixChars(25)
				col.spotlight = false
			}
			continue
		}

		col.y += col.speed

		if int(col.y)-col.length > b.height {
			col.active = false
			col.spotlight = false
			continue
		}

		// Random spotlight (briefly go bright white)
		if !col.spotlight && b.rng.Float64() < 0.003 {
			col.spotlight = true
			col.spotTimer = 5 + b.rng.Intn(8)
		}
		if col.spotlight {
			col.spotTimer--
			if col.spotTimer <= 0 {
				col.spotlight = false
			}
		}

		// Glitch characters
		if b.rng.Float64() < 0.08 && len(col.chars) > 0 {
			idx := b.rng.Intn(len(col.chars))
			col.chars[idx] = matrixChars[b.rng.Intn(len(matrixChars))]
		}

		headY := int(col.y)
		tailY := headY - col.length

		for y := maxInt(0, tailY); y <= minInt(headY, b.height-1); y++ {
			dist := headY - y
			brightness := 1.0 - float64(dist)/float64(col.length)
			if brightness < 0 {
				brightness = 0
			}

			var r, g, bv uint8
			if dist == 0 {
				// Head: bright white
				r, g, bv = 200, 255, 200
			} else if dist <= 2 {
				// Near head: bright green
				scale := brightness * 0.6
				if col.spotlight {
					r = uint8(180 * scale)
					g = uint8(255 * scale)
					bv = uint8(180 * scale)
				} else {
					r = uint8(30 * scale)
					g = uint8(255 * scale)
					bv = uint8(30 * scale)
				}
			} else {
				// Tail: fading green
				scale := brightness * 0.40
				if col.spotlight {
					r = uint8(100 * scale)
					g = uint8(255 * scale)
					bv = uint8(100 * scale)
				} else {
					r = 0
					g = uint8(255.0 * scale)
					bv = 0
				}
			}

			charIdx := y % len(col.chars)
			if charIdx < 0 {
				charIdx += len(col.chars)
			}
			ch := col.chars[charIdx]

			if b.charGrid[y] == nil {
				b.charGrid[y] = make(map[int]bgCell)
			}
			b.charGrid[y][i] = bgCell{ch: ch, fg: [3]uint8{r, g, bv}}
		}
	}
}

// --- Ocean (3D water surface) ---

func (b *BackgroundModel) updateOcean() {
	b.ensurePixelBuffer()
	b.pb.clear()

	if b.width == 0 || b.height == 0 {
		return
	}

	t := float64(b.frame) * 0.04

	for py := 0; py < b.height*2; py++ {
		fy := float64(py) / float64(b.height*2)
		// Perspective: bottom rows are "closer" — waves are larger
		perspective := 0.3 + fy*0.7
		waveScale := 1.0 + fy*2.0

		for x := 0; x < b.width; x++ {
			fx := float64(x) / float64(b.width)

			// Overlapping sine waves for wave height
			h1 := b.fastSin(fx*4.0*math.Pi*waveScale + t*1.2)
			h2 := b.fastSin(fx*7.0*math.Pi*waveScale + fy*3.0*math.Pi + t*0.8)
			h3 := b.fastSin((fx+fy)*5.0*math.Pi + t*0.6)
			h4 := b.fastSin(fx*11.0*math.Pi*waveScale*0.5 + t*1.8)

			waveH := (h1*0.4 + h2*0.3 + h3*0.2 + h4*0.1) // -1..1
			waveH = (waveH + 1.0) / 2.0                     // 0..1

			// Color based on wave height
			var cr, cg, cb uint8
			maxBright := 0.45 * perspective

			if waveH < 0.3 {
				// Deep trough: dark blue
				intensity := maxBright * 0.4
				cr = uint8(5 * intensity * 255 / 100)
				cg = uint8(15 * intensity * 255 / 100)
				cb = uint8(50 * intensity * 255 / 100)
			} else if waveH < 0.6 {
				// Mid: medium blue
				frac := (waveH - 0.3) / 0.3
				intensity := maxBright * (0.5 + frac*0.3)
				cr = uint8(10 * intensity * 255 / 100)
				cg = uint8((30 + 40*frac) * intensity * 255 / 100)
				cb = uint8((60 + 30*frac) * intensity * 255 / 100)
			} else if waveH < 0.85 {
				// Crest: cyan
				frac := (waveH - 0.6) / 0.25
				intensity := maxBright * (0.7 + frac*0.3)
				cr = uint8((20 + 40*frac) * intensity * 255 / 100)
				cg = uint8((70 + 50*frac) * intensity * 255 / 100)
				cb = uint8((90 + 10*frac) * intensity * 255 / 100)
			} else {
				// Foam/crest: bright white-cyan
				intensity := maxBright
				cr = uint8(80 * intensity * 255 / 100)
				cg = uint8(95 * intensity * 255 / 100)
				cb = uint8(100 * intensity * 255 / 100)
			}

			// Sparkle on crests
			if waveH > 0.88 && b.rng.Float64() < 0.15 {
				sparkle := uint8(maxBright * 255)
				cr = sparkle
				cg = sparkle
				cb = sparkle
			}

			b.pb.set(x, py, cr, cg, cb)
		}
	}
}

// --- Cube (rotating 3D wireframe) ---

// Unit cube vertices
var cubeVertices = [8][3]float64{
	{-1, -1, -1}, {1, -1, -1}, {1, 1, -1}, {-1, 1, -1},
	{-1, -1, 1}, {1, -1, 1}, {1, 1, 1}, {-1, 1, 1},
}

// Cube edges (pairs of vertex indices)
var cubeEdges = [12][2]int{
	{0, 1}, {1, 2}, {2, 3}, {3, 0}, // back face
	{4, 5}, {5, 6}, {6, 7}, {7, 4}, // front face
	{0, 4}, {1, 5}, {2, 6}, {3, 7}, // connecting edges
}

func (b *BackgroundModel) initCube() {
	b.cube = cubeState{
		angleX: 0,
		angleY: 0,
	}
}

func (b *BackgroundModel) updateCube() {
	b.ensurePixelBuffer()
	b.pb.clear()

	if b.width == 0 || b.height == 0 {
		return
	}

	b.cube.angleX += 0.025
	b.cube.angleY += 0.018

	pr, pg, ppb := b.themeRGB("primary")
	sr, sg, sb := b.themeRGB("secondary")

	cx := float64(b.width) / 2.0
	cy := float64(b.height) // pixel center y (double res)

	// Scale cube to fill ~60% of screen
	scale := math.Min(float64(b.width), float64(b.height)) * 0.55

	sinX := b.fastSin(b.cube.angleX)
	cosX := b.fastCos(b.cube.angleX)
	sinY := b.fastSin(b.cube.angleY)
	cosY := b.fastCos(b.cube.angleY)

	// Project vertices
	var projected [8][2]float64
	var zVals [8]float64
	for i, v := range cubeVertices {
		// Rotate around Y
		rx := v[0]*cosY - v[2]*sinY
		rz := v[0]*sinY + v[2]*cosY
		ry := v[1]

		// Rotate around X
		ry2 := ry*cosX - rz*sinX
		rz2 := ry*sinX + rz*cosX

		// Perspective projection
		dist := 4.0 + rz2
		if dist < 0.5 {
			dist = 0.5
		}
		px := cx + rx*scale/dist
		py := cy + ry2*scale/dist

		projected[i] = [2]float64{px, py}
		zVals[i] = rz2
	}

	// Save to history for trail effect
	b.cube.history = append(b.cube.history, projected)
	if len(b.cube.history) > 5 {
		b.cube.history = b.cube.history[len(b.cube.history)-5:]
	}

	// Draw trail frames (dimmer)
	for hi, hist := range b.cube.history {
		fade := float64(hi+1) / float64(len(b.cube.history)+1) * 0.15
		tr, tg, tb := dimColor(sr, sg, sb, fade)
		for _, edge := range cubeEdges {
			v0 := hist[edge[0]]
			v1 := hist[edge[1]]
			b.drawPixelLine(int(v0[0]), int(v0[1]), int(v1[0]), int(v1[1]), tr, tg, tb)
		}
	}

	// Draw current frame edges
	for _, edge := range cubeEdges {
		v0 := projected[edge[0]]
		v1 := projected[edge[1]]

		// Color based on average Z of the edge vertices (front = bright, back = dim)
		avgZ := (zVals[edge[0]] + zVals[edge[1]]) / 2.0
		depthFrac := (avgZ + 1.5) / 3.0 // normalize roughly to 0..1
		if depthFrac < 0 {
			depthFrac = 0
		}
		if depthFrac > 1 {
			depthFrac = 1
		}

		brightness := 0.20 + depthFrac*0.35
		cr, cg, cb := lerpColor(sr, sg, sb, pr, pg, ppb, depthFrac)
		cr, cg, cb = dimColor(cr, cg, cb, brightness)

		b.drawPixelLine(int(v0[0]), int(v0[1]), int(v1[0]), int(v1[1]), cr, cg, cb)
	}

	// Draw vertices as bright dots
	for i, p := range projected {
		depthFrac := (zVals[i] + 1.5) / 3.0
		if depthFrac < 0 {
			depthFrac = 0
		}
		if depthFrac > 1 {
			depthFrac = 1
		}
		brightness := 0.35 + depthFrac*0.35
		cr, cg, cb := dimColor(pr, pg, ppb, brightness)
		ix, iy := int(p[0]), int(p[1])
		// Draw 2x2 dot
		b.pb.set(ix, iy, cr, cg, cb)
		b.pb.set(ix+1, iy, cr, cg, cb)
		b.pb.set(ix, iy+1, cr, cg, cb)
		b.pb.set(ix+1, iy+1, cr, cg, cb)
	}
}

// drawPixelLine draws a line in the pixel buffer using Bresenham's algorithm.
func (b *BackgroundModel) drawPixelLine(x0, y0, x1, y1 int, r, g, bv uint8) {
	dx := x1 - x0
	dy := y1 - y0
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	sx := 1
	if x0 > x1 {
		sx = -1
	}
	sy := 1
	if y0 > y1 {
		sy = -1
	}

	err := dx - dy
	steps := 0
	maxSteps := dx + dy + 1
	if maxSteps > 1000 {
		maxSteps = 1000
	}

	for steps < maxSteps {
		b.pb.set(x0, y0, r, g, bv)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
		steps++
	}
}

// --- Animation dispatch ---

func (b *BackgroundModel) updateAnimation() {
	switch b.mode {
	case BgStarfield:
		b.updateStarfield()
	case BgTunnel:
		b.updateTunnel()
	case BgPlasma:
		b.updatePlasma()
	case BgFire:
		b.updateFire()
	case BgMatrix:
		b.updateMatrix()
	case BgOcean:
		b.updateOcean()
	case BgCube:
		b.updateCube()
	}
}

// --- Rendering ---

// isPixelMode returns true if the mode uses the pixel buffer (half-block rendering).
func (b *BackgroundModel) isPixelMode() bool {
	switch b.mode {
	case BgStarfield, BgTunnel, BgPlasma, BgFire, BgOcean, BgCube:
		return true
	}
	return false
}

// RenderLine renders a full background line at row y.
func (b *BackgroundModel) RenderLine(y, width int) string {
	if b.mode == BgOff || width == 0 {
		return strings.Repeat(" ", width)
	}
	if b.isPixelMode() && b.pb != nil {
		rendered := b.pb.renderRow(y)
		// Pad if needed
		renderedLen := b.pb.width
		if renderedLen < width {
			rendered += strings.Repeat(" ", width-renderedLen)
		}
		return rendered
	}
	return b.renderCharLine(y, width)
}

// renderCharLine renders a line for character-based effects (matrix).
func (b *BackgroundModel) renderCharLine(y, width int) string {
	row, hasRow := b.charGrid[y]
	if !hasRow {
		return strings.Repeat(" ", width)
	}

	var sb strings.Builder
	sb.Grow(width * 20)

	x := 0
	for x < width {
		if cell, ok := row[x]; ok {
			fmt.Fprintf(&sb, "\x1b[38;2;%d;%d;%dm%c\x1b[0m", cell.fg[0], cell.fg[1], cell.fg[2], cell.ch)
			x++
		} else {
			start := x
			for x < width {
				if _, ok := row[x]; ok {
					break
				}
				x++
			}
			sb.WriteString(strings.Repeat(" ", x-start))
		}
	}

	return sb.String()
}

// RenderSegment renders background from startX to endX at row y.
func (b *BackgroundModel) RenderSegment(y, startX, endX int) string {
	if b.mode == BgOff || startX >= endX {
		return ""
	}

	if b.isPixelMode() && b.pb != nil {
		// For pixel modes, render the segment using half-blocks
		var sb strings.Builder
		sb.Grow((endX - startX) * 40)
		for x := startX; x < endX; x++ {
			top := b.pb.get(x, y*2)
			bot := b.pb.get(x, y*2+1)
			if top.r == 0 && top.g == 0 && top.b == 0 &&
				bot.r == 0 && bot.g == 0 && bot.b == 0 {
				sb.WriteByte(' ')
			} else {
				fmt.Fprintf(&sb, "\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm▀\x1b[0m",
					top.r, top.g, top.b, bot.r, bot.g, bot.b)
			}
		}
		return sb.String()
	}

	var sb strings.Builder
	row := b.charGrid[y]
	for x := startX; x < endX; x++ {
		if row != nil {
			if cell, ok := row[x]; ok {
				fmt.Fprintf(&sb, "\x1b[38;2;%d;%d;%dm%c\x1b[0m", cell.fg[0], cell.fg[1], cell.fg[2], cell.ch)
				continue
			}
		}
		sb.WriteByte(' ')
	}

	return sb.String()
}

// ApplyToView composites the background behind the main view content.
func (b *BackgroundModel) ApplyToView(view string, width, height int) string {
	if b.mode == BgOff || width == 0 || height == 0 {
		return view
	}

	lines := strings.Split(view, "\n")
	result := make([]string, 0, height)

	for y := 0; y < height; y++ {
		if y < len(lines) {
			line := lines[y]
			stripped := stripAnsi(line)

			if strings.TrimSpace(stripped) == "" {
				result = append(result, b.RenderLine(y, width))
			} else {
				composed := b.compositeLineWithBg(line, y, width)
				result = append(result, composed)
			}
		} else {
			result = append(result, b.RenderLine(y, width))
		}
	}

	return strings.Join(result, "\n")
}

// compositeLineWithBg walks through a line's ANSI sequences and injects
// background colors from the animation behind characters that don't already
// have an explicit background set.
func (b *BackgroundModel) compositeLineWithBg(line string, row, totalWidth int) string {
	var result strings.Builder
	col := 0
	hasBg := false
	i := 0
	runes := []rune(line)

	for i < len(runes) {
		if runes[i] == '\x1b' && i+1 < len(runes) && runes[i+1] == '[' {
			j := i + 2
			for j < len(runes) && !((runes[j] >= 'a' && runes[j] <= 'z') || (runes[j] >= 'A' && runes[j] <= 'Z')) {
				j++
			}
			if j < len(runes) {
				j++
			}
			seq := string(runes[i:j])
			if strings.Contains(seq, "48;") || strings.Contains(seq, "\x1b[4") {
				hasBg = true
			}
			if seq == "\x1b[m" || seq == "\x1b[0m" {
				hasBg = false
			}
			result.WriteString(seq)
			i = j
			continue
		}

		if !hasBg && col < totalWidth {
			bgR, bgG, bgB := b.cellBgColor(row, col)
			fmt.Fprintf(&result, "\x1b[48;2;%d;%d;%dm", bgR, bgG, bgB)
		}
		result.WriteRune(runes[i])
		col++
		i++
	}

	if col < totalWidth {
		result.WriteString(b.RenderSegment(row, col, totalWidth))
	}

	result.WriteString("\x1b[m")
	return result.String()
}

// cellAt returns the animation character and color at a grid position.
func (b *BackgroundModel) cellAt(row, col int) (ch rune, r, g, bv uint8) {
	if row < 0 || row >= b.height || col < 0 || col >= b.width {
		return ' ', 0, 0, 0
	}
	// For pixel modes, return a space with the pixel color as "foreground"
	if b.isPixelMode() && b.pb != nil {
		top := b.pb.get(col, row*2)
		bot := b.pb.get(col, row*2+1)
		r := uint8((int(top.r) + int(bot.r)) / 2)
		g := uint8((int(top.g) + int(bot.g)) / 2)
		bv := uint8((int(top.b) + int(bot.b)) / 2)
		return '▀', r, g, bv
	}
	if b.charGrid[row] != nil {
		if cell, ok := b.charGrid[row][col]; ok {
			return cell.ch, cell.fg[0], cell.fg[1], cell.fg[2]
		}
	}
	return ' ', 0, 0, 0
}

// cellBgColor returns the background RGB for a specific cell position.
func (b *BackgroundModel) cellBgColor(row, col int) (r, g, bVal int) {
	if b.isPixelMode() && b.pb != nil {
		return b.pb.bgColorAt(row, col)
	}
	_, cr, cg, cb := b.cellAt(row, col)
	if cr == 0 && cg == 0 && cb == 0 {
		return 6, 6, 10
	}
	return int(cr) * 25 / 100, int(cg) * 25 / 100, int(cb) * 25 / 100
}

// --- Color helpers ---

func stripAnsi(s string) string {
	var result strings.Builder
	inEsc := false
	for _, r := range s {
		if r == '\x1b' {
			inEsc = true
			continue
		}
		if inEsc {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEsc = false
			}
			continue
		}
		result.WriteRune(r)
	}
	return result.String()
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
