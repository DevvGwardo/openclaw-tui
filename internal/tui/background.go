package tui

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// BgMode identifies a background animation mode.
type BgMode string

const (
	BgOff       BgMode = "off"
	BgWave      BgMode = "wave"
	BgMatrix    BgMode = "matrix"
	BgAurora    BgMode = "aurora"
	BgRain      BgMode = "rain"
	BgParticles BgMode = "particles"
	BgPulse     BgMode = "pulse"
)

// BgModes lists all available background modes in cycle order.
var BgModes = []BgMode{BgOff, BgWave, BgMatrix, BgAurora, BgRain, BgParticles, BgPulse}

type matrixCol struct {
	y      float64
	speed  float64
	length int
	active bool
}

type raindrop struct {
	x     int
	y     float64
	speed float64
	char  rune
}

type particle struct {
	x, y   float64
	vx, vy float64
	char   rune
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

	// Matrix state
	matrixCols []matrixCol

	// Rain state
	raindrops []raindrop

	// Particles state
	particles []particle

	// Pre-built character grid for current frame (character effects only)
	charGrid map[int]map[int]bgCell

	rng *rand.Rand
}

// NewBackgroundModel creates a new background renderer.
func NewBackgroundModel(theme Theme) BackgroundModel {
	return BackgroundModel{
		mode:     BgOff,
		theme:    theme,
		charGrid: make(map[int]map[int]bgCell),
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
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

func (b *BackgroundModel) initMode() {
	switch b.mode {
	case BgMatrix:
		b.initMatrix()
	case BgRain:
		b.initRain()
	case BgParticles:
		b.initParticles()
	}
}

// --- Matrix ---

var matrixChars = []rune("アイウエオカキクケコサシスセソタチツテトナニヌネノハヒフヘホ0123456789ABCDEF")

func (b *BackgroundModel) initMatrix() {
	if b.width == 0 {
		return
	}
	// Only use every other column since CJK chars are double-width in some terminals
	cols := b.width
	b.matrixCols = make([]matrixCol, cols)
	for i := range b.matrixCols {
		b.matrixCols[i] = matrixCol{
			y:      float64(-b.rng.Intn(b.height+10)),
			speed:  0.2 + b.rng.Float64()*0.6,
			length: 4 + b.rng.Intn(12),
			active: b.rng.Float64() < 0.3,
		}
	}
}

// --- Rain ---

var rainChars = []rune{'│', '·', '|', ':', '.'}

func (b *BackgroundModel) initRain() {
	if b.width == 0 || b.height == 0 {
		return
	}
	count := b.width / 4
	if count < 10 {
		count = 10
	}
	b.raindrops = make([]raindrop, count)
	for i := range b.raindrops {
		b.raindrops[i] = raindrop{
			x:     b.rng.Intn(b.width),
			y:     float64(b.rng.Intn(b.height)),
			speed: 0.3 + b.rng.Float64()*1.2,
			char:  rainChars[b.rng.Intn(len(rainChars))],
		}
	}
}

// --- Particles ---

var particleChars = []rune{'·', '∘', '°', '⋅', '∙'}

func (b *BackgroundModel) initParticles() {
	if b.width == 0 || b.height == 0 {
		return
	}
	count := (b.width * b.height) / 100
	if count < 5 {
		count = 5
	}
	if count > 50 {
		count = 50
	}
	b.particles = make([]particle, count)
	for i := range b.particles {
		b.particles[i] = particle{
			x:    b.rng.Float64() * float64(b.width),
			y:    b.rng.Float64() * float64(b.height),
			vx:   (b.rng.Float64() - 0.5) * 0.3,
			vy:   -0.05 - b.rng.Float64()*0.2,
			char: particleChars[b.rng.Intn(len(particleChars))],
		}
	}
}

// --- Animation updates ---

func (b *BackgroundModel) updateAnimation() {
	switch b.mode {
	case BgMatrix:
		b.updateMatrix()
		b.buildMatrixGrid()
	case BgRain:
		b.updateRain()
		b.buildRainGrid()
	case BgParticles:
		b.updateParticles()
		b.buildParticleGrid()
	}
}

func (b *BackgroundModel) updateMatrix() {
	for i := range b.matrixCols {
		col := &b.matrixCols[i]
		if !col.active {
			if b.rng.Float64() < 0.008 {
				col.active = true
				col.y = 0
				col.speed = 0.2 + b.rng.Float64()*0.6
				col.length = 4 + b.rng.Intn(12)
			}
			continue
		}
		col.y += col.speed
		if int(col.y)-col.length > b.height {
			col.active = false
		}
	}
}

func (b *BackgroundModel) updateRain() {
	h := float64(b.height)
	for i := range b.raindrops {
		drop := &b.raindrops[i]
		drop.y += drop.speed
		if drop.y >= h {
			drop.y = 0
			drop.x = b.rng.Intn(b.width)
			drop.char = rainChars[b.rng.Intn(len(rainChars))]
		}
	}
}

func (b *BackgroundModel) updateParticles() {
	w := float64(b.width)
	h := float64(b.height)
	for i := range b.particles {
		p := &b.particles[i]
		p.x += p.vx
		p.y += p.vy
		if p.x < 0 {
			p.x += w
		} else if p.x >= w {
			p.x -= w
		}
		if p.y < 0 {
			p.y += h
		} else if p.y >= h {
			p.y -= h
		}
	}
}

// --- Grid builders for O(1) cell lookup ---

func (b *BackgroundModel) buildMatrixGrid() {
	b.charGrid = make(map[int]map[int]bgCell)
	pr, pg, pb := hexToRGB(string(b.theme.Palette.Primary))

	for x, col := range b.matrixCols {
		if !col.active {
			continue
		}
		headY := int(col.y)
		tailY := headY - col.length

		for y := max(0, tailY); y <= min(headY, b.height-1); y++ {
			dist := headY - y
			brightness := 1.0 - float64(dist)/float64(col.length)
			if brightness < 0 {
				brightness = 0
			}
			scale := brightness * 0.35

			charIdx := (x*7 + y*13 + b.frame) % len(matrixChars)

			if b.charGrid[y] == nil {
				b.charGrid[y] = make(map[int]bgCell)
			}
			b.charGrid[y][x] = bgCell{
				ch: matrixChars[charIdx],
				fg: [3]uint8{
					uint8(float64(pr) * scale),
					uint8(float64(pg) * scale),
					uint8(float64(pb) * scale),
				},
			}
		}
	}
}

func (b *BackgroundModel) buildRainGrid() {
	b.charGrid = make(map[int]map[int]bgCell)
	pr, pg, pb := hexToRGB(string(b.theme.Palette.Primary))

	for _, drop := range b.raindrops {
		y := int(drop.y)
		if y < 0 || y >= b.height || drop.x < 0 || drop.x >= b.width {
			continue
		}
		if b.charGrid[y] == nil {
			b.charGrid[y] = make(map[int]bgCell)
		}
		b.charGrid[y][drop.x] = bgCell{
			ch: drop.char,
			fg: [3]uint8{
				uint8(float64(pr) * 0.2),
				uint8(float64(pg) * 0.2),
				uint8(float64(pb) * 0.2),
			},
		}
	}
}

func (b *BackgroundModel) buildParticleGrid() {
	b.charGrid = make(map[int]map[int]bgCell)
	pr, pg, pb := hexToRGB(string(b.theme.Palette.Primary))

	for _, p := range b.particles {
		x, y := int(p.x), int(p.y)
		if y < 0 || y >= b.height || x < 0 || x >= b.width {
			continue
		}
		if b.charGrid[y] == nil {
			b.charGrid[y] = make(map[int]bgCell)
		}
		b.charGrid[y][x] = bgCell{
			ch: p.char,
			fg: [3]uint8{
				uint8(float64(pr) * 0.18),
				uint8(float64(pg) * 0.18),
				uint8(float64(pb) * 0.18),
			},
		}
	}
}

// --- Rendering ---

// renderBgCell returns the ANSI-formatted string for a single background cell.
// Uses direct ANSI codes for performance instead of lipgloss.
func (b *BackgroundModel) renderBgCell(x, y int) string {
	switch b.mode {
	case BgWave:
		return b.renderWaveCell(x, y)
	case BgMatrix, BgRain, BgParticles:
		return b.renderCharCell(x, y)
	case BgAurora:
		return b.renderAuroraCell(x, y)
	case BgPulse:
		return b.renderPulseCell(x, y)
	default:
		return " "
	}
}

func (b *BackgroundModel) renderWaveCell(x, y int) string {
	t := float64(b.frame) * 0.05
	fx := float64(x) * 0.08
	fy := float64(y) * 0.15

	val := (math.Sin(fx+t) + math.Sin(fy+t*0.7) + math.Sin((fx+fy)*0.5+t*0.5)) / 3.0
	val = (val + 1.0) / 2.0

	p := b.theme.Palette
	r, g, bv := interpolateColors3(p.Primary, p.Secondary, p.Accent, val)
	r, g, bv = r/5, g/5, bv/5

	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm \x1b[0m", r, g, bv)
}

func (b *BackgroundModel) renderCharCell(x, y int) string {
	if row, ok := b.charGrid[y]; ok {
		if cell, ok := row[x]; ok {
			return fmt.Sprintf("\x1b[38;2;%d;%d;%dm%c\x1b[0m", cell.fg[0], cell.fg[1], cell.fg[2], cell.ch)
		}
	}
	return " "
}

func (b *BackgroundModel) renderAuroraCell(x, y int) string {
	t := float64(b.frame) * 0.03
	fx := float64(x) * 0.04
	fy := float64(y) * 0.08

	v1 := math.Sin(fx + t)
	v2 := math.Sin(fy + t*0.7)
	v3 := math.Sin((fx+fy)*0.5 + t*0.5)
	v4 := math.Sin(math.Sqrt(fx*fx+fy*fy)*0.3 + t*0.3)

	r := uint8((math.Sin(v1*math.Pi+t)*0.5 + 0.5) * 35)
	g := uint8((math.Sin(v2*math.Pi+t*1.3)*0.5 + 0.5) * 45)
	bv := uint8((math.Sin((v3+v4)*math.Pi*0.5+t*0.7)*0.5 + 0.5) * 50)

	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm \x1b[0m", r, g, bv)
}

func (b *BackgroundModel) renderPulseCell(x, y int) string {
	t := float64(b.frame) * 0.04
	val := (math.Sin(t) + 1.0) / 2.0

	p := b.theme.Palette
	r1, g1, b1 := hexToRGB(string(p.Bg))
	r2, g2, b2 := hexToRGB(string(p.BgSubtle))

	r := lerpInt(r1, r2, val)
	g := lerpInt(g1, g2, val)
	bv := lerpInt(b1, b2, val)

	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm \x1b[0m", r, g, bv)
}

// RenderLine renders a full background line at row y.
func (b *BackgroundModel) RenderLine(y, width int) string {
	if b.mode == BgOff || width == 0 {
		return strings.Repeat(" ", width)
	}

	// For pulse mode, the entire line is the same color — batch it
	if b.mode == BgPulse {
		t := float64(b.frame) * 0.04
		val := (math.Sin(t) + 1.0) / 2.0
		p := b.theme.Palette
		r1, g1, b1 := hexToRGB(string(p.Bg))
		r2, g2, b2 := hexToRGB(string(p.BgSubtle))
		r := lerpInt(r1, r2, val)
		g := lerpInt(g1, g2, val)
		bv := lerpInt(b1, b2, val)
		return fmt.Sprintf("\x1b[48;2;%d;%d;%dm%s\x1b[0m", r, g, bv, strings.Repeat(" ", width))
	}

	// For wave/aurora, batch cells in groups for performance
	if b.mode == BgWave || b.mode == BgAurora {
		return b.renderColorLine(y, width)
	}

	// For character effects, render sparse characters
	return b.renderCharLine(y, width)
}

// renderColorLine batches color-based effects in groups of 4 columns.
func (b *BackgroundModel) renderColorLine(y, width int) string {
	var sb strings.Builder
	sb.Grow(width * 25)

	groupSize := 4
	for x := 0; x < width; x += groupSize {
		end := x + groupSize
		if end > width {
			end = width
		}
		count := end - x

		// Sample color at group midpoint
		mx := x + count/2
		cell := b.renderBgCell(mx, y)
		// Extract the ANSI bg color and apply to the whole group
		if b.mode == BgWave {
			t := float64(b.frame) * 0.05
			fx := float64(mx) * 0.08
			fy := float64(y) * 0.15
			val := (math.Sin(fx+t) + math.Sin(fy+t*0.7) + math.Sin((fx+fy)*0.5+t*0.5)) / 3.0
			val = (val + 1.0) / 2.0
			p := b.theme.Palette
			r, g, bv := interpolateColors3(p.Primary, p.Secondary, p.Accent, val)
			r, g, bv = r/5, g/5, bv/5
			fmt.Fprintf(&sb, "\x1b[48;2;%d;%d;%dm%s\x1b[0m", r, g, bv, strings.Repeat(" ", count))
		} else if b.mode == BgAurora {
			t := float64(b.frame) * 0.03
			fx := float64(mx) * 0.04
			fy := float64(y) * 0.08
			v1 := math.Sin(fx + t)
			v2 := math.Sin(fy + t*0.7)
			v3 := math.Sin((fx+fy)*0.5 + t*0.5)
			v4 := math.Sin(math.Sqrt(fx*fx+fy*fy)*0.3 + t*0.3)
			r := uint8((math.Sin(v1*math.Pi+t)*0.5 + 0.5) * 35)
			g := uint8((math.Sin(v2*math.Pi+t*1.3)*0.5 + 0.5) * 45)
			bv := uint8((math.Sin((v3+v4)*math.Pi*0.5+t*0.7)*0.5 + 0.5) * 50)
			fmt.Fprintf(&sb, "\x1b[48;2;%d;%d;%dm%s\x1b[0m", r, g, bv, strings.Repeat(" ", count))
		} else {
			_ = cell
			sb.WriteString(strings.Repeat(" ", count))
		}
	}

	return sb.String()
}

// renderCharLine renders a line for character-based effects (matrix, rain, particles).
func (b *BackgroundModel) renderCharLine(y, width int) string {
	row, hasRow := b.charGrid[y]
	if !hasRow {
		return strings.Repeat(" ", width)
	}

	var sb strings.Builder
	sb.Grow(width * 15)

	x := 0
	for x < width {
		if cell, ok := row[x]; ok {
			fmt.Fprintf(&sb, "\x1b[38;2;%d;%d;%dm%c\x1b[0m", cell.fg[0], cell.fg[1], cell.fg[2], cell.ch)
			x++
		} else {
			// Count consecutive empty cells
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

	var sb strings.Builder
	width := endX - startX

	switch b.mode {
	case BgPulse:
		t := float64(b.frame) * 0.04
		val := (math.Sin(t) + 1.0) / 2.0
		p := b.theme.Palette
		r1, g1, b1 := hexToRGB(string(p.Bg))
		r2, g2, b2 := hexToRGB(string(p.BgSubtle))
		r := lerpInt(r1, r2, val)
		g := lerpInt(g1, g2, val)
		bv := lerpInt(b1, b2, val)
		fmt.Fprintf(&sb, "\x1b[48;2;%d;%d;%dm%s\x1b[0m", r, g, bv, strings.Repeat(" ", width))

	case BgWave, BgAurora:
		for x := startX; x < endX; x++ {
			sb.WriteString(b.renderBgCell(x, y))
		}

	case BgMatrix, BgRain, BgParticles:
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

	default:
		sb.WriteString(strings.Repeat(" ", width))
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
			visWidth := lipgloss.Width(line)
			stripped := stripAnsi(line)

			if strings.TrimSpace(stripped) == "" {
				// Empty/whitespace line — replace with background
				result = append(result, b.RenderLine(y, width))
			} else if visWidth < width {
				// Pad right margin with background
				result = append(result, line+b.RenderSegment(y, visWidth, width))
			} else {
				result = append(result, line)
			}
		} else {
			// Below the view content — fill with background
			result = append(result, b.RenderLine(y, width))
		}
	}

	return strings.Join(result, "\n")
}

// --- Color helpers ---

func interpolateColors3(c1, c2, c3 lipgloss.Color, t float64) (uint8, uint8, uint8) {
	r1, g1, b1 := hexToRGB(string(c1))
	r2, g2, b2 := hexToRGB(string(c2))
	r3, g3, b3 := hexToRGB(string(c3))

	if t < 0.5 {
		t2 := t * 2.0
		return lerpInt(r1, r2, t2), lerpInt(g1, g2, t2), lerpInt(b1, b2, t2)
	}
	t2 := (t - 0.5) * 2.0
	return lerpInt(r2, r3, t2), lerpInt(g2, g3, t2), lerpInt(b2, b3, t2)
}

func lerpInt(a, b int, t float64) uint8 {
	return uint8(float64(a) + t*float64(b-a))
}

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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
