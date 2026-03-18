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
	BgSkibidi   BgMode = "skibidi"
	BgSigma     BgMode = "sigma"
	BgNpc       BgMode = "npc"
	BgOhio      BgMode = "ohio"
	BgRizz      BgMode = "rizz"
	BgGyatt     BgMode = "gyatt"
	BgAmogus    BgMode = "amogus"
	BgBussin    BgMode = "bussin"
	BgAquarium  BgMode = "aquarium"
)

// BgModes lists all available background modes in cycle order.
var BgModes = []BgMode{BgOff, BgStarfield, BgTunnel, BgPlasma, BgFire, BgMatrix, BgOcean, BgCube,
	BgSkibidi, BgSigma, BgNpc, BgOhio, BgRizz, BgGyatt, BgAmogus, BgBussin, BgAquarium}

// --- Pixel buffer for half-block rendering (color-intensive modes) ---

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

// --- Braille buffer for high-resolution shape rendering ---
// Unicode Braille (U+2800..U+28FF) encodes a 2×4 dot grid per character cell.
// Each cell = 2 pixels wide × 4 pixels tall = 8 subpixels.

// brailleDotBit maps (dx, dy) within a 2×4 cell to the braille bit flag.
var brailleDotBit = [2][4]rune{
	{0x01, 0x02, 0x04, 0x40}, // left column  (dx=0): rows 0-3
	{0x08, 0x10, 0x20, 0x80}, // right column (dx=1): rows 0-3
}

type pixColor struct {
	r, g, b uint8
}

type brailleBuffer struct {
	termW, termH int          // terminal dimensions
	pixW, pixH   int          // pixel dimensions (termW*2, termH*4)
	dots         []bool       // [pixH * pixW] flat array
	colors       []pixColor   // [pixH * pixW] color per dot
}

func newBrailleBuffer(termW, termH int) *brailleBuffer {
	if termW <= 0 || termH <= 0 {
		return &brailleBuffer{}
	}
	pixW := termW * 2
	pixH := termH * 4
	return &brailleBuffer{
		termW:  termW,
		termH:  termH,
		pixW:   pixW,
		pixH:   pixH,
		dots:   make([]bool, pixW*pixH),
		colors: make([]pixColor, pixW*pixH),
	}
}

func (bb *brailleBuffer) clear() {
	for i := range bb.dots {
		bb.dots[i] = false
		bb.colors[i] = pixColor{}
	}
}

func (bb *brailleBuffer) set(px, py int, r, g, b uint8) {
	if px < 0 || px >= bb.pixW || py < 0 || py >= bb.pixH {
		return
	}
	idx := py*bb.pixW + px
	bb.dots[idx] = true
	bb.colors[idx] = pixColor{r, g, b}
}

func (bb *brailleBuffer) get(px, py int) (bool, pixColor) {
	if px < 0 || px >= bb.pixW || py < 0 || py >= bb.pixH {
		return false, pixColor{}
	}
	idx := py*bb.pixW + px
	return bb.dots[idx], bb.colors[idx]
}

// renderRow renders a terminal row using braille characters with averaged fg colors.
func (bb *brailleBuffer) renderRow(termY int) string {
	if bb.termW <= 0 || termY < 0 || termY >= bb.termH {
		return ""
	}
	var sb strings.Builder
	sb.Grow(bb.termW * 30)

	basePixY := termY * 4

	// Track runs of identical char+color for batching
	type cellInfo struct {
		ch   rune
		r, g, b uint8
	}
	var prevCell cellInfo
	prevCell.ch = 0xFFFF // impossible sentinel
	run := 0

	flush := func() {
		if run <= 0 {
			return
		}
		if prevCell.ch == 0x2800 {
			// Empty braille = space
			sb.WriteString(strings.Repeat(" ", run))
		} else {
			fmt.Fprintf(&sb, "\x1b[38;2;%d;%d;%dm", prevCell.r, prevCell.g, prevCell.b)
			for k := 0; k < run; k++ {
				sb.WriteRune(prevCell.ch)
			}
			sb.WriteString("\x1b[0m")
		}
		run = 0
	}

	for tx := 0; tx < bb.termW; tx++ {
		basePixX := tx * 2

		var pattern rune
		var totalR, totalG, totalB int
		var dotCount int

		for dx := 0; dx < 2; dx++ {
			for dy := 0; dy < 4; dy++ {
				px := basePixX + dx
				py := basePixY + dy
				if px < bb.pixW && py < bb.pixH {
					lit, c := bb.get(px, py)
					if lit {
						pattern |= brailleDotBit[dx][dy]
						totalR += int(c.r)
						totalG += int(c.g)
						totalB += int(c.b)
						dotCount++
					}
				}
			}
		}

		ch := rune(0x2800 + pattern)
		var cr, cg, cb uint8
		if dotCount > 0 {
			cr = uint8(totalR / dotCount)
			cg = uint8(totalG / dotCount)
			cb = uint8(totalB / dotCount)
		}

		cur := cellInfo{ch, cr, cg, cb}
		if cur == prevCell {
			run++
		} else {
			flush()
			prevCell = cur
			run = 1
		}
	}
	flush()
	return sb.String()
}

// bgColorAt returns an averaged background color for compositing text over braille.
func (bb *brailleBuffer) bgColorAt(row, col int) (int, int, int) {
	if col < 0 || col >= bb.termW || row < 0 || row >= bb.termH {
		return 6, 6, 10
	}
	baseX := col * 2
	baseY := row * 4
	var totalR, totalG, totalB, count int
	for dy := 0; dy < 4; dy++ {
		for dx := 0; dx < 2; dx++ {
			px := baseX + dx
			py := baseY + dy
			if px < bb.pixW && py < bb.pixH {
				lit, c := bb.get(px, py)
				if lit {
					totalR += int(c.r)
					totalG += int(c.g)
					totalB += int(c.b)
					count++
				}
			}
		}
	}
	if count == 0 {
		return 6, 6, 10
	}
	return totalR / count, totalG / count, totalB / count
}

// drawLine draws a line in the braille buffer using Bresenham's algorithm.
func (bb *brailleBuffer) drawLine(x0, y0, x1, y1 int, r, g, b uint8) {
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
	if maxSteps > 4000 {
		maxSteps = 4000
	}

	for steps < maxSteps {
		bb.set(x0, y0, r, g, b)
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

// bgCell represents a single background cell for character-based effects.
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

	// Pixel buffer for half-block modes (color-intensive)
	pb *pixelBuffer

	// Braille buffer for shape-intensive modes (starfield, cube)
	bb *brailleBuffer

	// Starfield state
	stars []star3d

	// Fire state
	fireHeat []float64 // width * (height*2) heat values

	// Matrix state
	matrixCols []matrixCol

	// Cube state
	cube cubeState

	// Pre-built character grid for character-based effects (matrix)
	charGrid map[int]map[int]bgCell

	// Pre-computed sine table
	sinTable [1024]float64

	rng *rand.Rand

	// Brainrot mode state
	skibidiObjs  []skibidiObj
	sigmaTexts   []sigmaText
	npcs         []npcChar
	ohioGlitches []ohioGlitch
	ohioFlash    int
	rizzSparkles []rizzSparkle
	gyattTexts   []gyattText
	amogusCrews  []amogusCrew
	amogusSus    int
	bussinDrops  []bussinDrop

	// Aquarium state
	aquariumFish    []aquariumFish
	aquariumBubbles []aquariumBubble
	aquariumWeeds   []aquariumWeed
	aquariumCrabs   []aquariumCrab
	aquariumTasks   []string       // current task labels for crabs to display
	aquariumFood    []aquariumFood // dropped food particles
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
// Mode returns the current background mode.
func (b *BackgroundModel) Mode() BgMode {
	return b.mode
}

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
// For aquarium mode, we preserve the existing entities and just let them adapt to new bounds.
func (b *BackgroundModel) SetSize(w, h int) {
	if w == b.width && h == b.height {
		return
	}
	b.width = w
	b.height = h
	// Only re-init if not in aquarium mode - aquarium adapts dynamically
	if b.mode != BgAquarium {
		b.initMode()
	}
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

func (b *BackgroundModel) ensureBrailleBuffer() {
	if b.bb == nil || b.bb.termW != b.width || b.bb.termH != b.height {
		b.bb = newBrailleBuffer(b.width, b.height)
	}
}

// isBrailleMode returns true if the mode uses braille rendering.
func (b *BackgroundModel) isBrailleMode() bool {
	switch b.mode {
	case BgStarfield, BgCube, BgAmogus:
		return true
	}
	return false
}

// isPixelMode returns true if the mode uses the pixel buffer (half-block rendering).
func (b *BackgroundModel) isPixelMode() bool {
	switch b.mode {
	case BgTunnel, BgPlasma, BgFire, BgOcean, BgSigma, BgOhio, BgRizz, BgBussin, BgAquarium:
		return true
	}
	return false
}

// isCharMode returns true if the mode uses character grid rendering.
func (b *BackgroundModel) isCharMode() bool {
	switch b.mode {
	case BgMatrix, BgSkibidi, BgNpc, BgGyatt:
		return true
	}
	return false
}

func (b *BackgroundModel) initMode() {
	if b.isPixelMode() {
		b.ensurePixelBuffer()
	}
	if b.isBrailleMode() {
		b.ensureBrailleBuffer()
	}
	switch b.mode {
	case BgStarfield:
		b.initStarfield()
	case BgFire:
		b.initFire()
	case BgMatrix:
		b.initMatrix()
	case BgCube:
		b.initCube()
	case BgSkibidi:
		b.initSkibidi()
	case BgSigma:
		b.initSigma()
	case BgNpc:
		b.initNpc()
	case BgOhio:
		b.initOhio()
	case BgRizz:
		b.initRizz()
	case BgGyatt:
		b.initGyatt()
	case BgAmogus:
		b.initAmogus()
	case BgBussin:
		b.initBussin()
	case BgAquarium:
		b.initAquarium()
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

// --- Starfield (3D warp speed — braille rendering) ---

func (b *BackgroundModel) initStarfield() {
	if b.width == 0 || b.height == 0 {
		return
	}
	count := 350 // more stars — braille can show them all as fine dots
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
	if r < 0.10 {
		tint = 1 // blue
	} else if r < 0.18 {
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
	b.ensureBrailleBuffer()
	b.bb.clear()

	if b.width == 0 || b.height == 0 {
		return
	}

	// Braille pixel space: pixW = width*2, pixH = height*4
	pixW := float64(b.bb.pixW)
	pixH := float64(b.bb.pixH)
	cx := pixW / 2.0
	cy := pixH / 2.0

	for i := range b.stars {
		s := &b.stars[i]

		// Project to braille pixel space
		sx := cx + (s.x/s.z)*cx
		sy := cy + (s.y/s.z)*cy

		prevSX := s.prevSX
		prevSY := s.prevSY
		s.prevSX = sx
		s.prevSY = sy

		// Move star toward viewer
		s.z -= 0.016
		if s.z <= 0.005 {
			b.stars[i] = b.newStar3d(false)
			continue
		}

		// Check bounds (with margin)
		if sx < -4 || sx >= pixW+4 || sy < -4 || sy >= pixH+4 {
			b.stars[i] = b.newStar3d(false)
			continue
		}

		// Brightness: exponential as z→0
		brightness := math.Pow(1.0-s.z, 2.5)
		if brightness > 1.0 {
			brightness = 1.0
		}

		// Base color
		var cr, cg, cb float64
		switch s.colorTint {
		case 1: // blue tint
			cr, cg, cb = 0.55, 0.65, 1.0
		case 2: // yellow tint
			cr, cg, cb = 1.0, 0.92, 0.55
		default: // white
			cr, cg, cb = 1.0, 1.0, 1.0
		}

		maxBright := 0.6
		br := brightness * maxBright
		r := uint8(cr * br * 255)
		g := uint8(cg * br * 255)
		bv := uint8(cb * br * 255)

		ix := int(sx)
		iy := int(sy)

		// Draw star as a single braille dot — fine point of light
		b.bb.set(ix, iy, r, g, bv)

		// Close stars get a small cross pattern for brightness
		if s.z < 0.15 {
			b.bb.set(ix+1, iy, r, g, bv)
			b.bb.set(ix-1, iy, r, g, bv)
			b.bb.set(ix, iy+1, r, g, bv)
			b.bb.set(ix, iy-1, r, g, bv)
		} else if s.z < 0.3 {
			// Medium stars: 2-dot cluster
			b.bb.set(ix+1, iy, r, g, bv)
		}

		// Motion trail — thin line from previous to current
		if prevSX >= 0 && prevSY >= 0 && s.z < 0.55 {
			trailSteps := 6
			if s.z < 0.12 {
				trailSteps = 12
			} else if s.z < 0.3 {
				trailSteps = 8
			}
			for t := 1; t <= trailSteps; t++ {
				frac := float64(t) / float64(trailSteps+1)
				tx := int(sx + (prevSX-sx)*frac)
				ty := int(sy + (prevSY-sy)*frac)
				fade := (1.0 - frac) * br * 0.45
				tr := uint8(cr * fade * 255)
				tg := uint8(cg * fade * 255)
				tb := uint8(cb * fade * 255)
				b.bb.set(tx, ty, tr, tg, tb)
			}
		}
	}

	// Speed lines at edges — thin braille dot streaks
	if b.frame%2 == 0 {
		edgeW := b.bb.pixW / 6
		for k := 0; k < 6; k++ {
			px := b.rng.Intn(edgeW)
			if b.rng.Intn(2) == 0 {
				px = b.bb.pixW - 1 - px
			}
			py := b.rng.Intn(b.bb.pixH)
			lineLen := 4 + b.rng.Intn(8)
			for j := 0; j < lineLen; j++ {
				fade := 0.18 * (1.0 - float64(j)/float64(lineLen))
				v := uint8(fade * 255)
				b.bb.set(px, py+j, v, v, v)
			}
		}
	}
}

// --- Tunnel (3D wormhole — half-block, color-intensive) ---

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

			dist := math.Sqrt(dx*dx + dy*dy)
			if dist < 0.5 {
				dist = 0.5
			}

			angle := math.Atan2(dy, dx)
			depth := 80.0 / dist

			u := angle/(2.0*math.Pi) + 0.5
			v := depth + t*2.0
			undulate := 1.0 + 0.15*b.fastSin(angle*3.0+t*1.5)
			v *= undulate

			ring := b.fastSin(v * 4.0 * math.Pi)
			depthBright := 1.0 / (1.0 + depth*0.15)

			depthFrac := depth / 5.0
			if depthFrac > 1 {
				depthFrac = 1
			}

			var cr, cg, cb uint8
			if ring > 0 {
				intensity := ring * depthBright * 0.45
				cr, cg, cb = lerpColor(pr, pg, ppb, sr, sg, sb, depthFrac)
				cr, cg, cb = dimColor(cr, cg, cb, intensity)
			} else {
				intensity := (1.0 + ring*0.5) * depthBright * 0.12
				cr = uint8(intensity * 40)
				cg = uint8(intensity * 20)
				cb = uint8(intensity * 60)
			}

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

// --- Plasma (classic demoscene — half-block, color-intensive) ---

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

			v1 := b.fastSin(fx*6.0*math.Pi + t)
			v2 := b.fastSin(fy*8.0*math.Pi + t*1.3)
			v3 := b.fastSin((fx+fy)*5.0*math.Pi + t*0.7)
			v4 := b.fastSin(math.Sqrt((fx-0.5)*(fx-0.5)*16+(fy-0.5)*(fy-0.5)*16)*4.0*math.Pi + t*1.1)
			v5 := b.fastSin(fx*3.0*math.Pi+b.fastSin(fy*4.0*math.Pi+t)*2.0+t*0.5)

			val := (v1 + v2 + v3 + v4 + v5) / 5.0
			val = (val + 1.0) / 2.0

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

// --- Fire (Doom-style — half-block, color-intensive) ---

func (b *BackgroundModel) initFire() {
	if b.width == 0 || b.height == 0 {
		return
	}
	pH := b.height * 2
	b.fireHeat = make([]float64, b.width*pH)
	for x := 0; x < b.width; x++ {
		b.fireHeat[(pH-1)*b.width+x] = 1.0
	}
}

func fireColor(heat float64) (uint8, uint8, uint8) {
	if heat < 0 {
		heat = 0
	}
	if heat > 1 {
		heat = 1
	}
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

	if len(b.fireHeat) != w*pH {
		b.initFire()
	}

	for x := 0; x < w; x++ {
		b.fireHeat[(pH-1)*w+x] = 0.7 + b.rng.Float64()*0.3
	}

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

			coolRate := 0.012 + 0.006*b.rng.Float64()
			avg -= coolRate
			if avg < 0 {
				avg = 0
			}
			b.fireHeat[py*w+x] = avg
		}
	}

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

// --- Matrix (character-based with braille glyphs) ---

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
				r, g, bv = 200, 255, 200
			} else if dist <= 2 {
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

			// Use braille characters for detail in the tail
			if dist > 3 && brightness < 0.5 {
				// Map the character to a braille glyph for organic look
				brailleBase := rune(0x2800)
				pattern := rune(ch) % 255
				if pattern == 0 {
					pattern = 1
				}
				ch = brailleBase + pattern
			}

			if b.charGrid[y] == nil {
				b.charGrid[y] = make(map[int]bgCell)
			}
			b.charGrid[y][i] = bgCell{ch: ch, fg: [3]uint8{r, g, bv}}
		}
	}
}

// --- Ocean (3D water surface — half-block, color-intensive) ---

func (b *BackgroundModel) updateOcean() {
	b.ensurePixelBuffer()
	b.pb.clear()

	if b.width == 0 || b.height == 0 {
		return
	}

	t := float64(b.frame) * 0.04

	for py := 0; py < b.height*2; py++ {
		fy := float64(py) / float64(b.height*2)
		perspective := 0.3 + fy*0.7
		waveScale := 1.0 + fy*2.0

		for x := 0; x < b.width; x++ {
			fx := float64(x) / float64(b.width)

			h1 := b.fastSin(fx*4.0*math.Pi*waveScale + t*1.2)
			h2 := b.fastSin(fx*7.0*math.Pi*waveScale + fy*3.0*math.Pi + t*0.8)
			h3 := b.fastSin((fx+fy)*5.0*math.Pi + t*0.6)
			h4 := b.fastSin(fx*11.0*math.Pi*waveScale*0.5 + t*1.8)

			waveH := (h1*0.4 + h2*0.3 + h3*0.2 + h4*0.1)
			waveH = (waveH + 1.0) / 2.0

			var cr, cg, cb uint8
			maxBright := 0.45 * perspective

			if waveH < 0.3 {
				intensity := maxBright * 0.4
				cr = uint8(5 * intensity * 255 / 100)
				cg = uint8(15 * intensity * 255 / 100)
				cb = uint8(50 * intensity * 255 / 100)
			} else if waveH < 0.6 {
				frac := (waveH - 0.3) / 0.3
				intensity := maxBright * (0.5 + frac*0.3)
				cr = uint8(10 * intensity * 255 / 100)
				cg = uint8((30 + 40*frac) * intensity * 255 / 100)
				cb = uint8((60 + 30*frac) * intensity * 255 / 100)
			} else if waveH < 0.85 {
				frac := (waveH - 0.6) / 0.25
				intensity := maxBright * (0.7 + frac*0.3)
				cr = uint8((20 + 40*frac) * intensity * 255 / 100)
				cg = uint8((70 + 50*frac) * intensity * 255 / 100)
				cb = uint8((90 + 10*frac) * intensity * 255 / 100)
			} else {
				intensity := maxBright
				cr = uint8(80 * intensity * 255 / 100)
				cg = uint8(95 * intensity * 255 / 100)
				cb = uint8(100 * intensity * 255 / 100)
			}

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

// --- Cube (rotating 3D wireframe — braille rendering) ---

var cubeVertices = [8][3]float64{
	{-1, -1, -1}, {1, -1, -1}, {1, 1, -1}, {-1, 1, -1},
	{-1, -1, 1}, {1, -1, 1}, {1, 1, 1}, {-1, 1, 1},
}

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
	b.ensureBrailleBuffer()
	b.bb.clear()

	if b.width == 0 || b.height == 0 {
		return
	}

	b.cube.angleX += 0.025
	b.cube.angleY += 0.018

	pr, pg, ppb := b.themeRGB("primary")
	sr, sg, sb := b.themeRGB("secondary")

	// Braille pixel space center
	pixW := float64(b.bb.pixW)
	pixH := float64(b.bb.pixH)
	cx := pixW / 2.0
	cy := pixH / 2.0

	// Scale to fill ~60% of braille pixel space
	scale := math.Min(pixW, pixH) * 0.28

	sinX := b.fastSin(b.cube.angleX)
	cosX := b.fastCos(b.cube.angleX)
	sinY := b.fastSin(b.cube.angleY)
	cosY := b.fastCos(b.cube.angleY)

	// Project vertices into braille pixel space
	var projected [8][2]float64
	var zVals [8]float64
	for i, v := range cubeVertices {
		rx := v[0]*cosY - v[2]*sinY
		rz := v[0]*sinY + v[2]*cosY
		ry := v[1]

		ry2 := ry*cosX - rz*sinX
		rz2 := ry*sinX + rz*cosX

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
	if len(b.cube.history) > 6 {
		b.cube.history = b.cube.history[len(b.cube.history)-6:]
	}

	// Draw trail frames (dimmer older frames)
	for hi, hist := range b.cube.history {
		fade := float64(hi+1) / float64(len(b.cube.history)+1) * 0.12
		tr, tg, tb := dimColor(sr, sg, sb, fade)
		for _, edge := range cubeEdges {
			v0 := hist[edge[0]]
			v1 := hist[edge[1]]
			b.bb.drawLine(int(v0[0]), int(v0[1]), int(v1[0]), int(v1[1]), tr, tg, tb)
		}
	}

	// Draw current frame edges — high resolution braille lines
	for _, edge := range cubeEdges {
		v0 := projected[edge[0]]
		v1 := projected[edge[1]]

		avgZ := (zVals[edge[0]] + zVals[edge[1]]) / 2.0
		depthFrac := (avgZ + 1.5) / 3.0
		if depthFrac < 0 {
			depthFrac = 0
		}
		if depthFrac > 1 {
			depthFrac = 1
		}

		brightness := 0.25 + depthFrac*0.40
		cr, cg, cb := lerpColor(sr, sg, sb, pr, pg, ppb, depthFrac)
		cr, cg, cb = dimColor(cr, cg, cb, brightness)

		b.bb.drawLine(int(v0[0]), int(v0[1]), int(v1[0]), int(v1[1]), cr, cg, cb)
	}

	// Draw vertices as small dot clusters (3x3 for braille sub-pixel precision)
	for i, p := range projected {
		depthFrac := (zVals[i] + 1.5) / 3.0
		if depthFrac < 0 {
			depthFrac = 0
		}
		if depthFrac > 1 {
			depthFrac = 1
		}
		brightness := 0.40 + depthFrac*0.35
		cr, cg, cb := dimColor(pr, pg, ppb, brightness)
		ix, iy := int(p[0]), int(p[1])
		// 3x3 dot cluster for visible vertices
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				b.bb.set(ix+dx, iy+dy, cr, cg, cb)
			}
		}
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
	case BgSkibidi:
		b.updateSkibidi()
	case BgSigma:
		b.updateSigma()
	case BgNpc:
		b.updateNpc()
	case BgOhio:
		b.updateOhio()
	case BgRizz:
		b.updateRizz()
	case BgGyatt:
		b.updateGyatt()
	case BgAmogus:
		b.updateAmogus()
	case BgBussin:
		b.updateBussin()
	case BgAquarium:
		b.updateAquarium()
	}
}

// --- Rendering ---

// RenderLine renders a full background line at row y.
func (b *BackgroundModel) RenderLine(y, width int) string {
	if b.mode == BgOff || width == 0 {
		return strings.Repeat(" ", width)
	}
	if b.isBrailleMode() && b.bb != nil {
		rendered := b.bb.renderRow(y)
		renderedLen := b.bb.termW
		if renderedLen < width {
			rendered += strings.Repeat(" ", width-renderedLen)
		}
		return rendered
	}
	if b.isPixelMode() && b.pb != nil {
		rendered := b.pb.renderRow(y)
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

	if b.isBrailleMode() && b.bb != nil {
		// For braille modes, render segment as individual braille cells
		var sb strings.Builder
		sb.Grow((endX - startX) * 30)
		basePixY := y * 4
		for tx := startX; tx < endX; tx++ {
			basePixX := tx * 2
			var pattern rune
			var totalR, totalG, totalB int
			var dotCount int

			for dx := 0; dx < 2; dx++ {
				for dy := 0; dy < 4; dy++ {
					px := basePixX + dx
					py := basePixY + dy
					if px < b.bb.pixW && py < b.bb.pixH {
						lit, c := b.bb.get(px, py)
						if lit {
							pattern |= brailleDotBit[dx][dy]
							totalR += int(c.r)
							totalG += int(c.g)
							totalB += int(c.b)
							dotCount++
						}
					}
				}
			}

			ch := rune(0x2800 + pattern)
			if dotCount == 0 {
				sb.WriteByte(' ')
			} else {
				cr := uint8(totalR / dotCount)
				cg := uint8(totalG / dotCount)
				cb := uint8(totalB / dotCount)
				fmt.Fprintf(&sb, "\x1b[38;2;%d;%d;%dm%c\x1b[0m", cr, cg, cb, ch)
			}
		}
		return sb.String()
	}

	if b.isPixelMode() && b.pb != nil {
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
		var rendered string
		if y < len(lines) {
			line := lines[y]
			stripped := stripAnsi(line)

			if strings.TrimSpace(stripped) == "" {
				rendered = b.RenderLine(y, width)
			} else {
				rendered = b.compositeLineWithBg(line, y, width)
			}
		} else {
			rendered = b.RenderLine(y, width)
		}

		result = append(result, rendered)
	}

	return strings.Join(result, "\n")
}

// overlayCrabLabels composites crab task label text onto a rendered background line.
func (b *BackgroundModel) overlayCrabLabels(line string, labels []crabLabel, row, totalWidth int) string {
	// Build a character map of what to overlay
	type overlayChar struct {
		ch   rune
		r, g, bv int
	}
	overlays := make(map[int]overlayChar)
	for _, cl := range labels {
		runes := []rune(cl.text)
		// Center the label on the crab position
		startCol := cl.col - len(runes)/2
		for i, ch := range runes {
			col := startCol + i
			if col >= 0 && col < totalWidth {
				overlays[col] = overlayChar{ch: ch, r: 255, g: 230, bv: 180} // warm white text
			}
		}
	}

	if len(overlays) == 0 {
		return line
	}

	// Walk the rendered line and replace characters at overlay positions
	runes := []rune(line)
	var result strings.Builder
	col := 0
	i := 0

	for i < len(runes) {
		// Skip ANSI sequences
		if runes[i] == '\x1b' && i+1 < len(runes) && runes[i+1] == '[' {
			j := i + 2
			for j < len(runes) && !((runes[j] >= 'a' && runes[j] <= 'z') || (runes[j] >= 'A' && runes[j] <= 'Z')) {
				j++
			}
			if j < len(runes) {
				j++
			}
			result.WriteString(string(runes[i:j]))
			i = j
			continue
		}

		if ov, ok := overlays[col]; ok {
			// Get background color from pixel buffer for this cell
			bgR, bgG, bgB := b.cellBgColor(row, col)
			fmt.Fprintf(&result, "\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm%c",
				ov.r, ov.g, ov.bv, bgR, bgG, bgB, ov.ch)
			result.WriteString("\x1b[0m")
			col++
			i++
			continue
		}

		result.WriteRune(runes[i])
		col++
		i++
	}

	return result.String()
}

// OverlayCrabLabelsOnView overlays crab task labels on top of an already-rendered view.
// This is called after all UI elements are rendered so labels appear above everything.
func (b *BackgroundModel) OverlayCrabLabelsOnView(view string, width, height int) string {
	if b.mode != BgAquarium {
		return view
	}

	crabLabels := b.CrabLabels()
	if len(crabLabels) == 0 {
		return view
	}

	labelsByRow := make(map[int][]crabLabel)
	for _, cl := range crabLabels {
		labelsByRow[cl.row] = append(labelsByRow[cl.row], cl)
	}

	lines := strings.Split(view, "\n")
	result := make([]string, 0, height)

	for y := 0; y < height; y++ {
		var line string
		if y < len(lines) {
			line = lines[y]
		} else {
			line = strings.Repeat(" ", width)
		}

		// Overlay crab task labels on this row
		if labels, ok := labelsByRow[y]; ok {
			line = b.overlayCrabLabels(line, labels, y, width)
		}

		result = append(result, line)
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
	if b.isBrailleMode() && b.bb != nil {
		// Return averaged color from braille cell
		baseX := col * 2
		baseY := row * 4
		var totalR, totalG, totalB, count int
		for dy := 0; dy < 4; dy++ {
			for dx := 0; dx < 2; dx++ {
				lit, c := b.bb.get(baseX+dx, baseY+dy)
				if lit {
					totalR += int(c.r)
					totalG += int(c.g)
					totalB += int(c.b)
					count++
				}
			}
		}
		if count > 0 {
			return '⠿', uint8(totalR / count), uint8(totalG / count), uint8(totalB / count)
		}
		return ' ', 0, 0, 0
	}
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
	if b.isBrailleMode() && b.bb != nil {
		return b.bb.bgColorAt(row, col)
	}
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
