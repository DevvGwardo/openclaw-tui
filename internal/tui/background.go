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
	BgOff      BgMode = "off"
	BgStarfield BgMode = "starfield"
	BgPipes    BgMode = "pipes"
	BgDVD      BgMode = "dvd"
	BgMystify  BgMode = "mystify"
	BgMaze     BgMode = "maze"
	BgToasters BgMode = "toasters"
	BgMatrix   BgMode = "matrix"
)

// BgModes lists all available background modes in cycle order.
var BgModes = []BgMode{BgOff, BgStarfield, BgPipes, BgDVD, BgMystify, BgMaze, BgToasters, BgMatrix}

// --- State types ---

type star struct {
	x, y, z float64 // 3D position, z is depth (1.0 = far, 0.01 = close)
}

type pipeSegment struct {
	x, y  int
	ch    rune
	color int // 0=primary, 1=secondary, 2=accent
	age   int
}

type pipeHead struct {
	x, y      int
	dx, dy    int
	color     int
	segCount  int
	turnIn    int // cells until next turn
}

type dvdState struct {
	x, y     float64
	dx, dy   float64
	colorIdx int
}

type vertex struct {
	x, y   float64
	vx, vy float64
}

type polygon struct {
	verts    [4]vertex
	color    int
	history  [][4][2]int // past positions for trail
}

type mazeCell struct {
	visited bool
	walls   [4]bool // top, right, bottom, left
}

type mazeState struct {
	cells   [][]mazeCell
	stack   [][2]int
	curX    int
	curY    int
	done    bool
	fadeOut float64 // 0 = visible, 1 = fully faded
	mw, mh int     // maze dimensions
}

type toaster struct {
	x, y   float64
	speed  float64
	kind   int // 0=toaster, 1=toast
}

type matrixCol struct {
	y      float64
	speed  float64
	length int
	active bool
	chars  []rune // cached characters for this column
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

	// Starfield state
	stars []star

	// Pipes state
	pipeHeads    []pipeHead
	pipeSegments []pipeSegment

	// DVD state
	dvd dvdState

	// Mystify state
	polygons []polygon

	// Maze state
	maze mazeState

	// Toasters state
	toasters []toaster

	// Matrix state
	matrixCols []matrixCol

	// Pre-built character grid for current frame
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
	case BgStarfield:
		b.initStarfield()
	case BgPipes:
		b.initPipes()
	case BgDVD:
		b.initDVD()
	case BgMystify:
		b.initMystify()
	case BgMaze:
		b.initMaze()
	case BgToasters:
		b.initToasters()
	case BgMatrix:
		b.initMatrix()
	}
}

// --- Starfield ---

func (b *BackgroundModel) initStarfield() {
	if b.width == 0 || b.height == 0 {
		return
	}
	count := 80 + b.rng.Intn(41) // 80-120 stars
	b.stars = make([]star, count)
	for i := range b.stars {
		b.stars[i] = b.newStar()
	}
}

func (b *BackgroundModel) newStar() star {
	return star{
		x: (b.rng.Float64() - 0.5) * 2.0, // -1 to 1
		y: (b.rng.Float64() - 0.5) * 2.0,
		z: b.rng.Float64()*0.9 + 0.1, // 0.1 to 1.0
	}
}

func (b *BackgroundModel) updateStarfield() {
	for i := range b.stars {
		b.stars[i].z -= 0.015
		if b.stars[i].z <= 0.01 {
			b.stars[i] = b.newStar()
			b.stars[i].z = 1.0
		}
		// Check if projected position is off-screen
		cx := float64(b.width) / 2.0
		cy := float64(b.height) / 2.0
		sx := cx + (b.stars[i].x/b.stars[i].z)*cx
		sy := cy + (b.stars[i].y/b.stars[i].z)*cy
		if sx < 0 || sx >= float64(b.width) || sy < 0 || sy >= float64(b.height) {
			b.stars[i] = b.newStar()
			b.stars[i].z = 1.0
		}
	}
}

func (b *BackgroundModel) buildStarfieldGrid() {
	b.charGrid = make(map[int]map[int]bgCell)
	cx := float64(b.width) / 2.0
	cy := float64(b.height) / 2.0

	for _, s := range b.stars {
		// Perspective projection
		sx := int(cx + (s.x/s.z)*cx)
		sy := int(cy + (s.y/s.z)*cy)

		if sx < 0 || sx >= b.width || sy < 0 || sy >= b.height {
			continue
		}

		// Pick character and brightness based on depth
		var ch rune
		var brightness float64
		switch {
		case s.z > 0.7:
			ch = '·'
			brightness = 0.2
		case s.z > 0.4:
			ch = '∙'
			brightness = 0.4
		case s.z > 0.15:
			ch = '•'
			brightness = 0.7
		default:
			ch = '★'
			brightness = 1.0
		}

		gray := uint8(brightness * 180)

		if b.charGrid[sy] == nil {
			b.charGrid[sy] = make(map[int]bgCell)
		}
		b.charGrid[sy][sx] = bgCell{
			ch: ch,
			fg: [3]uint8{gray, gray, gray},
		}
	}
}

// --- Pipes ---

var pipeCorners = map[[2]int]map[[2]int]rune{
	// from direction -> to direction -> corner character
	{0, -1}: {{-1, 0}: '┐', {1, 0}: '┘'},  // was going up, turning
	{0, 1}:  {{-1, 0}: '┌', {1, 0}: '└'},  // was going down, turning (note: this is visual; directions are dx,dy)
	{-1, 0}: {{0, -1}: '└', {0, 1}: '┌'},  // was going left
	{1, 0}:  {{0, -1}: '┘', {0, 1}: '┐'},  // was going right
}

func (b *BackgroundModel) initPipes() {
	if b.width == 0 || b.height == 0 {
		return
	}
	b.pipeSegments = nil
	b.pipeHeads = nil
	for i := 0; i < 3; i++ {
		b.pipeHeads = append(b.pipeHeads, b.newPipeHead(i))
	}
}

func (b *BackgroundModel) newPipeHead(color int) pipeHead {
	dirs := [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	d := dirs[b.rng.Intn(len(dirs))]
	return pipeHead{
		x:        b.rng.Intn(b.width),
		y:        b.rng.Intn(b.height),
		dx:       d[0],
		dy:       d[1],
		color:    color,
		segCount: 0,
		turnIn:   3 + b.rng.Intn(10),
	}
}

func (b *BackgroundModel) updatePipes() {
	for i := range b.pipeHeads {
		ph := &b.pipeHeads[i]

		// Determine character for current segment
		var ch rune
		if ph.dx != 0 {
			ch = '─'
		} else {
			ch = '│'
		}

		ph.turnIn--
		if ph.turnIn <= 0 {
			// Turn! Pick a perpendicular direction
			oldDx, oldDy := ph.dx, ph.dy
			if ph.dx != 0 {
				// Was horizontal, go vertical
				if b.rng.Intn(2) == 0 {
					ph.dy = -1
				} else {
					ph.dy = 1
				}
				ph.dx = 0
			} else {
				// Was vertical, go horizontal
				if b.rng.Intn(2) == 0 {
					ph.dx = -1
				} else {
					ph.dx = 1
				}
				ph.dy = 0
			}
			ph.turnIn = 3 + b.rng.Intn(10)

			// Use corner character
			from := [2]int{oldDx, oldDy}
			to := [2]int{ph.dx, ph.dy}
			if corner, ok := pipeCorners[from][to]; ok {
				ch = corner
			} else {
				ch = '┼'
			}
		}

		b.pipeSegments = append(b.pipeSegments, pipeSegment{
			x:     ph.x,
			y:     ph.y,
			ch:    ch,
			color: ph.color,
			age:   0,
		})
		ph.segCount++

		// Move head
		ph.x += ph.dx
		ph.y += ph.dy

		// Check bounds or too long
		if ph.x < 0 || ph.x >= b.width || ph.y < 0 || ph.y >= b.height || ph.segCount > 80 {
			b.pipeHeads[i] = b.newPipeHead(ph.color)
		}
	}

	// Age all segments
	for i := range b.pipeSegments {
		b.pipeSegments[i].age++
	}

	// Remove segments that are too old
	if len(b.pipeSegments) > 200 {
		b.pipeSegments = b.pipeSegments[len(b.pipeSegments)-200:]
	}
}

func (b *BackgroundModel) buildPipesGrid() {
	b.charGrid = make(map[int]map[int]bgCell)

	colors := [3][3]uint8{}
	r, g, bv := hexToRGB(string(b.theme.Palette.Primary))
	colors[0] = [3]uint8{uint8(r), uint8(g), uint8(bv)}
	r, g, bv = hexToRGB(string(b.theme.Palette.Secondary))
	colors[1] = [3]uint8{uint8(r), uint8(g), uint8(bv)}
	r, g, bv = hexToRGB(string(b.theme.Palette.Accent))
	colors[2] = [3]uint8{uint8(r), uint8(g), uint8(bv)}

	for _, seg := range b.pipeSegments {
		if seg.x < 0 || seg.x >= b.width || seg.y < 0 || seg.y >= b.height {
			continue
		}

		c := colors[seg.color%3]
		// Fade based on age
		fade := 1.0 - float64(seg.age)/250.0
		if fade < 0.1 {
			fade = 0.1
		}
		// Keep dim
		scale := fade * 0.35

		if b.charGrid[seg.y] == nil {
			b.charGrid[seg.y] = make(map[int]bgCell)
		}
		b.charGrid[seg.y][seg.x] = bgCell{
			ch: seg.ch,
			fg: [3]uint8{
				uint8(float64(c[0]) * scale),
				uint8(float64(c[1]) * scale),
				uint8(float64(c[2]) * scale),
			},
		}
	}
}

// --- DVD ---

var dvdLogo = [3]string{
	"╭───╮",
	"│🦞 │",
	"╰───╯",
}
var dvdLogoWidth = 5
var dvdLogoHeight = 3

func (b *BackgroundModel) initDVD() {
	if b.width == 0 || b.height == 0 {
		return
	}
	b.dvd = dvdState{
		x:        float64(b.rng.Intn(maxInt(1, b.width-dvdLogoWidth))),
		y:        float64(b.rng.Intn(maxInt(1, b.height-dvdLogoHeight))),
		dx:       1.0,
		dy:       0.5,
		colorIdx: 0,
	}
	// Randomize initial direction
	if b.rng.Intn(2) == 0 {
		b.dvd.dx = -1.0
	}
	if b.rng.Intn(2) == 0 {
		b.dvd.dy = -0.5
	}
}

var dvdColors = []string{"Primary", "Secondary", "Accent", "Success", "Warning"}

func (b *BackgroundModel) getDVDColor() [3]uint8 {
	p := b.theme.Palette
	var hex string
	switch dvdColors[b.dvd.colorIdx%len(dvdColors)] {
	case "Primary":
		hex = string(p.Primary)
	case "Secondary":
		hex = string(p.Secondary)
	case "Accent":
		hex = string(p.Accent)
	case "Success":
		hex = string(p.Success)
	case "Warning":
		hex = string(p.Warning)
	default:
		hex = string(p.Primary)
	}
	r, g, bv := hexToRGB(hex)
	return [3]uint8{uint8(r), uint8(g), uint8(bv)}
}

func (b *BackgroundModel) updateDVD() {
	b.dvd.x += b.dvd.dx
	b.dvd.y += b.dvd.dy

	bounced := false
	if b.dvd.x <= 0 {
		b.dvd.x = 0
		b.dvd.dx = math.Abs(b.dvd.dx)
		bounced = true
	}
	if b.dvd.x >= float64(b.width-dvdLogoWidth) {
		b.dvd.x = float64(b.width - dvdLogoWidth)
		b.dvd.dx = -math.Abs(b.dvd.dx)
		bounced = true
	}
	if b.dvd.y <= 0 {
		b.dvd.y = 0
		b.dvd.dy = math.Abs(b.dvd.dy)
		bounced = true
	}
	if b.dvd.y >= float64(b.height-dvdLogoHeight) {
		b.dvd.y = float64(b.height - dvdLogoHeight)
		b.dvd.dy = -math.Abs(b.dvd.dy)
		bounced = true
	}
	if bounced {
		b.dvd.colorIdx++
	}
}

func (b *BackgroundModel) buildDVDGrid() {
	b.charGrid = make(map[int]map[int]bgCell)
	ix := int(b.dvd.x)
	iy := int(b.dvd.y)

	c := b.getDVDColor()
	// Dim the color
	dimC := [3]uint8{c[0] / 2, c[1] / 2, c[2] / 2}

	for row, line := range dvdLogo {
		y := iy + row
		if y < 0 || y >= b.height {
			continue
		}
		if b.charGrid[y] == nil {
			b.charGrid[y] = make(map[int]bgCell)
		}
		col := 0
		for _, ch := range line {
			x := ix + col
			if x >= 0 && x < b.width {
				b.charGrid[y][x] = bgCell{ch: ch, fg: dimC}
			}
			col++
		}
	}
}

// --- Mystify ---

func (b *BackgroundModel) initMystify() {
	if b.width == 0 || b.height == 0 {
		return
	}
	b.polygons = make([]polygon, 3)
	for i := range b.polygons {
		b.polygons[i] = polygon{color: i}
		for j := range b.polygons[i].verts {
			b.polygons[i].verts[j] = vertex{
				x:  b.rng.Float64() * float64(b.width),
				y:  b.rng.Float64() * float64(b.height),
				vx: (b.rng.Float64() - 0.5) * 2.0,
				vy: (b.rng.Float64() - 0.5) * 1.5,
			}
		}
	}
}

func (b *BackgroundModel) updateMystify() {
	w := float64(b.width)
	h := float64(b.height)
	for pi := range b.polygons {
		// Save current positions to history
		var pos [4][2]int
		for vi := range b.polygons[pi].verts {
			pos[vi] = [2]int{int(b.polygons[pi].verts[vi].x), int(b.polygons[pi].verts[vi].y)}
		}
		b.polygons[pi].history = append(b.polygons[pi].history, pos)
		if len(b.polygons[pi].history) > 8 {
			b.polygons[pi].history = b.polygons[pi].history[len(b.polygons[pi].history)-8:]
		}

		// Update vertex positions
		for vi := range b.polygons[pi].verts {
			v := &b.polygons[pi].verts[vi]
			v.x += v.vx
			v.y += v.vy
			if v.x <= 0 {
				v.x = 0
				v.vx = math.Abs(v.vx)
			}
			if v.x >= w-1 {
				v.x = w - 1
				v.vx = -math.Abs(v.vx)
			}
			if v.y <= 0 {
				v.y = 0
				v.vy = math.Abs(v.vy)
			}
			if v.y >= h-1 {
				v.y = h - 1
				v.vy = -math.Abs(v.vy)
			}
		}
	}
}

func (b *BackgroundModel) buildMystifyGrid() {
	b.charGrid = make(map[int]map[int]bgCell)

	colors := [3][3]uint8{}
	r, g, bv := hexToRGB(string(b.theme.Palette.Primary))
	colors[0] = [3]uint8{uint8(r), uint8(g), uint8(bv)}
	r, g, bv = hexToRGB(string(b.theme.Palette.Secondary))
	colors[1] = [3]uint8{uint8(r), uint8(g), uint8(bv)}
	r, g, bv = hexToRGB(string(b.theme.Palette.Accent))
	colors[2] = [3]uint8{uint8(r), uint8(g), uint8(bv)}

	for _, poly := range b.polygons {
		c := colors[poly.color%3]

		// Draw trail (older = dimmer)
		for hi, pos := range poly.history {
			fade := float64(hi+1) / float64(len(poly.history)+1) * 0.15
			b.drawQuad(pos, c, fade)
		}

		// Draw current polygon
		var curPos [4][2]int
		for vi, v := range poly.verts {
			curPos[vi] = [2]int{int(v.x), int(v.y)}
		}
		b.drawQuad(curPos, c, 0.35)
	}
}

func (b *BackgroundModel) drawQuad(verts [4][2]int, color [3]uint8, brightness float64) {
	// Draw lines between consecutive vertices and close the quad
	for i := 0; i < 4; i++ {
		j := (i + 1) % 4
		b.drawLine(verts[i][0], verts[i][1], verts[j][0], verts[j][1], color, brightness)
	}
}

func (b *BackgroundModel) drawLine(x0, y0, x1, y1 int, color [3]uint8, brightness float64) {
	// Bresenham's line algorithm
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
	if maxSteps > 500 {
		maxSteps = 500
	}

	for steps < maxSteps {
		if x0 >= 0 && x0 < b.width && y0 >= 0 && y0 < b.height {
			// Choose character based on line angle
			var ch rune
			if dx > dy*2 {
				ch = '─'
			} else if dy > dx*2 {
				ch = '│'
			} else if (sx > 0) == (sy > 0) {
				ch = '\\'
			} else {
				ch = '/'
			}

			if b.charGrid[y0] == nil {
				b.charGrid[y0] = make(map[int]bgCell)
			}
			b.charGrid[y0][x0] = bgCell{
				ch: ch,
				fg: [3]uint8{
					uint8(float64(color[0]) * brightness),
					uint8(float64(color[1]) * brightness),
					uint8(float64(color[2]) * brightness),
				},
			}
		}

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

// --- Maze ---

func (b *BackgroundModel) initMaze() {
	if b.width == 0 || b.height == 0 {
		return
	}
	// Maze cells are 2 chars wide, 1 char tall for display
	b.maze.mw = b.width / 3
	b.maze.mh = b.height / 2
	if b.maze.mw < 2 {
		b.maze.mw = 2
	}
	if b.maze.mh < 2 {
		b.maze.mh = 2
	}

	b.maze.cells = make([][]mazeCell, b.maze.mh)
	for y := range b.maze.cells {
		b.maze.cells[y] = make([]mazeCell, b.maze.mw)
		for x := range b.maze.cells[y] {
			b.maze.cells[y][x].walls = [4]bool{true, true, true, true}
		}
	}

	startX := b.rng.Intn(b.maze.mw)
	startY := b.rng.Intn(b.maze.mh)
	b.maze.curX = startX
	b.maze.curY = startY
	b.maze.cells[startY][startX].visited = true
	b.maze.stack = [][2]int{{startX, startY}}
	b.maze.done = false
	b.maze.fadeOut = 0
}

func (b *BackgroundModel) updateMaze() {
	if b.maze.done {
		b.maze.fadeOut += 0.02
		if b.maze.fadeOut >= 1.0 {
			b.initMaze()
		}
		return
	}

	// Grow 2-3 cells per tick
	for steps := 0; steps < 3; steps++ {
		if len(b.maze.stack) == 0 {
			b.maze.done = true
			return
		}

		cur := b.maze.stack[len(b.maze.stack)-1]
		cx, cy := cur[0], cur[1]
		b.maze.curX = cx
		b.maze.curY = cy

		// Find unvisited neighbors
		type neighbor struct {
			x, y, wall, oppWall int
		}
		dirs := []neighbor{
			{cx, cy - 1, 0, 2}, // top
			{cx + 1, cy, 1, 3}, // right
			{cx, cy + 1, 2, 0}, // bottom
			{cx - 1, cy, 3, 1}, // left
		}

		var unvisited []neighbor
		for _, d := range dirs {
			if d.x >= 0 && d.x < b.maze.mw && d.y >= 0 && d.y < b.maze.mh && !b.maze.cells[d.y][d.x].visited {
				unvisited = append(unvisited, d)
			}
		}

		if len(unvisited) > 0 {
			n := unvisited[b.rng.Intn(len(unvisited))]
			b.maze.cells[cy][cx].walls[n.wall] = false
			b.maze.cells[n.y][n.x].walls[n.oppWall] = false
			b.maze.cells[n.y][n.x].visited = true
			b.maze.stack = append(b.maze.stack, [2]int{n.x, n.y})
		} else {
			b.maze.stack = b.maze.stack[:len(b.maze.stack)-1]
		}
	}
}

func (b *BackgroundModel) buildMazeGrid() {
	b.charGrid = make(map[int]map[int]bgCell)

	br, bg, bb := hexToRGB(string(b.theme.Palette.FgMuted))
	wallScale := 0.25
	if b.maze.done {
		wallScale = 0.25 * (1.0 - b.maze.fadeOut)
	}

	pr, ppg, pb := hexToRGB(string(b.theme.Palette.Primary))

	for my := 0; my < b.maze.mh; my++ {
		for mx := 0; mx < b.maze.mw; mx++ {
			if !b.maze.cells[my][mx].visited {
				continue
			}

			// Screen position: each maze cell is 3 wide, 2 tall
			sx := mx * 3
			sy := my * 2

			cell := b.maze.cells[my][mx]

			// Draw walls
			wallColor := [3]uint8{
				uint8(float64(br) * wallScale),
				uint8(float64(bg) * wallScale),
				uint8(float64(bb) * wallScale),
			}

			// Top wall
			if cell.walls[0] {
				b.setGridCell(sx, sy, '┌', wallColor)
				b.setGridCell(sx+1, sy, '─', wallColor)
				b.setGridCell(sx+2, sy, '┐', wallColor)
			} else {
				b.setGridCell(sx, sy, '┌', wallColor)
				b.setGridCell(sx+2, sy, '┐', wallColor)
			}
			// Left wall
			if cell.walls[3] {
				b.setGridCell(sx, sy+1, '│', wallColor)
			}
			// Right wall
			if cell.walls[1] {
				b.setGridCell(sx+2, sy+1, '│', wallColor)
			}
			// Bottom wall
			if cell.walls[2] {
				b.setGridCell(sx, sy+1, '└', wallColor)
				b.setGridCell(sx+1, sy+1, '─', wallColor)
				b.setGridCell(sx+2, sy+1, '┘', wallColor)
			}
		}
	}

	// Draw current cell as bright dot
	if !b.maze.done {
		csx := b.maze.curX*3 + 1
		csy := b.maze.curY*2 + 1
		curColor := [3]uint8{uint8(pr), uint8(ppg), uint8(pb)}
		b.setGridCell(csx, csy, '●', curColor)
	}
}

func (b *BackgroundModel) setGridCell(x, y int, ch rune, color [3]uint8) {
	if x < 0 || x >= b.width || y < 0 || y >= b.height {
		return
	}
	if b.charGrid[y] == nil {
		b.charGrid[y] = make(map[int]bgCell)
	}
	b.charGrid[y][x] = bgCell{ch: ch, fg: color}
}

// --- Toasters ---

var toasterShape = [2]string{
	"[╦]",
	"[╩]",
}
var toasterWidth = 3
var toasterHeight = 2

func (b *BackgroundModel) initToasters() {
	if b.width == 0 || b.height == 0 {
		return
	}
	count := 5 + b.rng.Intn(4) // 5-8
	b.toasters = make([]toaster, count)
	for i := range b.toasters {
		b.toasters[i] = b.newToaster()
		// Scatter initial positions
		b.toasters[i].x = b.rng.Float64() * float64(b.width)
		b.toasters[i].y = b.rng.Float64() * float64(b.height)
	}
}

func (b *BackgroundModel) newToaster() toaster {
	return toaster{
		x:     float64(b.width) + float64(b.rng.Intn(20)),
		y:     float64(-b.rng.Intn(b.height)),
		speed: 0.5 + b.rng.Float64()*1.0,
		kind:  b.rng.Intn(2),
	}
}

func (b *BackgroundModel) updateToasters() {
	for i := range b.toasters {
		t := &b.toasters[i]
		// Move from upper-right to lower-left
		t.x -= t.speed
		t.y += t.speed * 0.6

		// Respawn if off screen
		if t.x < -float64(toasterWidth) || t.y > float64(b.height+toasterHeight) {
			b.toasters[i] = b.newToaster()
		}
	}
}

func (b *BackgroundModel) buildToasterGrid() {
	b.charGrid = make(map[int]map[int]bgCell)

	sr, sg, sb := hexToRGB(string(b.theme.Palette.Secondary))
	ar, ag, ab := hexToRGB(string(b.theme.Palette.Accent))

	for _, t := range b.toasters {
		ix := int(t.x)
		iy := int(t.y)

		var cr, cg, cb int
		if t.kind == 0 {
			cr, cg, cb = sr, sg, sb
		} else {
			cr, cg, cb = ar, ag, ab
		}

		// Dim
		scale := 0.25
		color := [3]uint8{
			uint8(float64(cr) * scale),
			uint8(float64(cg) * scale),
			uint8(float64(cb) * scale),
		}

		if t.kind == 0 {
			// Toaster
			for row, line := range toasterShape {
				y := iy + row
				col := 0
				for _, ch := range line {
					x := ix + col
					b.setGridCell(x, y, ch, color)
					col++
				}
			}
		} else {
			// Toast - simple block
			b.setGridCell(ix, iy, '▪', color)
			b.setGridCell(ix+1, iy, '▪', color)
		}
	}
}

// --- Matrix ---

var matrixChars = []rune("ｦｧｨｩｪｫｬｭｮｯｱｲｳｴｵｶｷｸｹｺ0123456789")

func (b *BackgroundModel) initMatrix() {
	if b.width == 0 {
		return
	}
	cols := b.width
	b.matrixCols = make([]matrixCol, cols)
	for i := range b.matrixCols {
		b.matrixCols[i] = matrixCol{
			y:      float64(-b.rng.Intn(b.height + 10)),
			speed:  0.3 + b.rng.Float64()*0.8,
			length: 5 + b.rng.Intn(15),
			active: b.rng.Float64() < 0.4,
			chars:  b.randomMatrixChars(20),
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
	for i := range b.matrixCols {
		col := &b.matrixCols[i]
		if !col.active {
			if b.rng.Float64() < 0.012 {
				col.active = true
				col.y = 0
				col.speed = 0.3 + b.rng.Float64()*0.8
				col.length = 5 + b.rng.Intn(15)
				col.chars = b.randomMatrixChars(20)
			}
			continue
		}
		col.y += col.speed
		if int(col.y)-col.length > b.height {
			col.active = false
		}

		// Glitch: randomly swap a character
		if b.rng.Float64() < 0.05 && len(col.chars) > 0 {
			idx := b.rng.Intn(len(col.chars))
			col.chars[idx] = matrixChars[b.rng.Intn(len(matrixChars))]
		}
	}
}

func (b *BackgroundModel) buildMatrixGrid() {
	b.charGrid = make(map[int]map[int]bgCell)

	for x, col := range b.matrixCols {
		if !col.active {
			continue
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
				// Head: bright white-green
				r = 80
				g = 128
				bv = 80
			} else {
				// Tail: pure green, fading to dark green
				scale := brightness * 0.35
				r = 0
				g = uint8(255.0 * scale)
				bv = 0
			}

			charIdx := y
			if len(col.chars) > 0 {
				charIdx = y % len(col.chars)
				if charIdx < 0 {
					charIdx += len(col.chars)
				}
			}
			ch := matrixChars[0]
			if len(col.chars) > 0 {
				ch = col.chars[charIdx]
			}

			if b.charGrid[y] == nil {
				b.charGrid[y] = make(map[int]bgCell)
			}
			b.charGrid[y][x] = bgCell{
				ch: ch,
				fg: [3]uint8{r, g, bv},
			}
		}
	}
}

// --- Animation updates ---

func (b *BackgroundModel) updateAnimation() {
	switch b.mode {
	case BgStarfield:
		b.updateStarfield()
		b.buildStarfieldGrid()
	case BgPipes:
		b.updatePipes()
		b.buildPipesGrid()
	case BgDVD:
		b.updateDVD()
		b.buildDVDGrid()
	case BgMystify:
		b.updateMystify()
		b.buildMystifyGrid()
	case BgMaze:
		b.updateMaze()
		b.buildMazeGrid()
	case BgToasters:
		b.updateToasters()
		b.buildToasterGrid()
	case BgMatrix:
		b.updateMatrix()
		b.buildMatrixGrid()
	}
}

// --- Rendering ---

// RenderLine renders a full background line at row y.
func (b *BackgroundModel) RenderLine(y, width int) string {
	if b.mode == BgOff || width == 0 {
		return strings.Repeat(" ", width)
	}
	return b.renderCharLine(y, width)
}

// renderCharLine renders a line for character-based effects.
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
			visWidth := lipgloss.Width(line)
			stripped := stripAnsi(line)

			if strings.TrimSpace(stripped) == "" {
				result = append(result, b.RenderLine(y, width))
			} else if visWidth < width {
				result = append(result, line+b.RenderSegment(y, visWidth, width))
			} else {
				result = append(result, line)
			}
		} else {
			result = append(result, b.RenderLine(y, width))
		}
	}

	return strings.Join(result, "\n")
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
