package tui

import "math"

// clampF clamps a float64 to [lo, hi].
func clampF(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// --- Brainrot state types ---

type skibidiObj struct {
	x, y   float64
	vx, vy float64
	text   string
	color  int
}

type sigmaText struct {
	x, y    int
	text    string
	life    int
	maxLife int
}

type npcChar struct {
	x           float64
	y           int
	dir         int
	speed       float64
	frame       int
	speechTimer int
	speech      string
}

type ohioGlitch struct {
	y, shift, life int
}

type rizzSparkle struct {
	x, y float64
	vy   float64
}

type gyattText struct {
	x, y    int
	life    int
	maxLife int
}

type amogusCrew struct {
	x     float64
	y     int
	dir   int
	speed float64
	color [3]uint8
	frame int
}

type bussinDrop struct {
	x     int
	y     float64
	speed float64
	ch    rune
	color [3]uint8
}

// ============================================================
// SKIBIDI — Chaotic bouncing text, flashing colors, pure chaos
// ============================================================

var skibidiStrings = []string{"SKIBIDI", "TOILET", "DOP DOP", "YES YES", "BRRR", "skibidi", "bop bop"}
var skibidiPalette = [][3]uint8{
	{120, 30, 150}, {30, 140, 130}, {150, 130, 30},
	{150, 30, 30}, {30, 150, 30}, {150, 70, 30},
}

func (b *BackgroundModel) initSkibidi() {
	if b.width == 0 || b.height == 0 {
		return
	}
	n := 8 + b.width/20
	b.skibidiObjs = make([]skibidiObj, n)
	for i := range b.skibidiObjs {
		t := skibidiStrings[b.rng.Intn(len(skibidiStrings))]
		mxW := maxInt(1, b.width-len(t))
		mxH := maxInt(1, b.height)
		b.skibidiObjs[i] = skibidiObj{
			x:     float64(b.rng.Intn(mxW)),
			y:     float64(b.rng.Intn(mxH)),
			vx:    (b.rng.Float64()-0.5)*2.5 + 0.3,
			vy:    (b.rng.Float64()-0.5)*1.8 + 0.2,
			text:  t,
			color: b.rng.Intn(len(skibidiPalette)),
		}
	}
}

func (b *BackgroundModel) updateSkibidi() {
	b.charGrid = make(map[int]map[int]bgCell)
	if b.width == 0 || b.height == 0 {
		return
	}

	for i := range b.skibidiObjs {
		o := &b.skibidiObjs[i]
		o.x += o.vx
		o.y += o.vy

		mx := float64(maxInt(1, b.width-len(o.text)))
		if o.x <= 0 || o.x >= mx {
			o.vx = -o.vx
			o.color = (o.color + 1) % len(skibidiPalette)
		}
		if o.y <= 0 || o.y >= float64(b.height-1) {
			o.vy = -o.vy
			o.color = (o.color + 1) % len(skibidiPalette)
		}
		o.x = clampF(o.x, 0, mx)
		o.y = clampF(o.y, 0, float64(b.height-1))

		// Flash colors chaotically
		ci := (o.color + b.frame/3 + i) % len(skibidiPalette)
		c := skibidiPalette[ci]

		row, col := int(o.y), int(o.x)
		for j, ch := range o.text {
			x := col + j
			if x >= 0 && x < b.width && row >= 0 && row < b.height {
				if b.charGrid[row] == nil {
					b.charGrid[row] = make(map[int]bgCell)
				}
				b.charGrid[row][x] = bgCell{ch: ch, fg: c}
			}
		}
	}

	// Scatter toilet-ish 'T' chars randomly for extra chaos
	if b.frame%5 == 0 {
		for k := 0; k < 3; k++ {
			row := b.rng.Intn(maxInt(1, b.height))
			col := b.rng.Intn(maxInt(1, b.width))
			if b.charGrid[row] == nil {
				b.charGrid[row] = make(map[int]bgCell)
			}
			ci := b.rng.Intn(len(skibidiPalette))
			c := skibidiPalette[ci]
			// dim it
			c = [3]uint8{c[0] / 3, c[1] / 3, c[2] / 3}
			b.charGrid[row][col] = bgCell{ch: 'T', fg: c}
		}
	}
}

// ============================================================
// SIGMA — Dark moody. Deep red/blue plasma. Lightning. Edgy.
// ============================================================

func (b *BackgroundModel) initSigma() {
	b.sigmaTexts = nil
}

func (b *BackgroundModel) updateSigma() {
	b.ensurePixelBuffer()
	b.pb.clear()

	if b.width == 0 || b.height == 0 {
		return
	}

	t := float64(b.frame) * 0.02

	// Dark moody ambient — slow plasma in deep reds and dark navy
	for py := 0; py < b.height*2; py++ {
		fy := float64(py) / float64(b.height*2)
		for x := 0; x < b.width; x++ {
			fx := float64(x) / float64(b.width)

			v1 := b.fastSin(fx*3.0*math.Pi + t*0.5)
			v2 := b.fastSin(fy*4.0*math.Pi + t*0.3)
			v3 := b.fastSin((fx+fy)*2.5*math.Pi + t*0.7)
			val := ((v1 + v2 + v3) / 3.0) // -1..1
			val = (val + 1.0) / 2.0        // 0..1

			// Blood red to dark navy
			var cr, cg, cb uint8
			if val < 0.5 {
				f := val / 0.5
				cr = uint8(30 * f)
				cb = uint8(12 + 15*f)
			} else {
				f := (val - 0.5) / 0.5
				cr = uint8(30 + 15*f)
				cb = uint8(27 - 10*f)
			}
			cg = 2
			b.pb.set(x, py, cr, cg, cb)
		}
	}

	// Lightning bolt (rare, dramatic)
	if b.rng.Float64() < 0.03 {
		lx := b.rng.Intn(b.width)
		for py := 0; py < b.height*2; py++ {
			lx += b.rng.Intn(3) - 1
			lx = maxInt(0, minInt(b.width-1, lx))
			v := uint8(40 + b.rng.Intn(30))
			bv := uint8(minInt(255, int(float64(v)*1.3)))
			b.pb.set(lx, py, v, v, bv)
		}
	}
}

// ============================================================
// NPC — Stick figures walking back and forth with speech bubbles
// ============================================================

var npcSpeeches = []string{"huh?", "...", "ok", "hmm", "bruh", "what", "same", "lol"}

func (b *BackgroundModel) initNpc() {
	if b.width == 0 || b.height == 0 {
		return
	}
	n := 5 + b.width/25
	b.npcs = make([]npcChar, n)
	for i := range b.npcs {
		dir := 1
		if b.rng.Intn(2) == 0 {
			dir = -1
		}
		b.npcs[i] = npcChar{
			x:     float64(b.rng.Intn(maxInt(1, b.width))),
			y:     b.rng.Intn(maxInt(1, b.height-4)) + 3,
			dir:   dir,
			speed: 0.15 + b.rng.Float64()*0.3,
			frame: b.rng.Intn(20),
		}
	}
}

func (b *BackgroundModel) updateNpc() {
	b.charGrid = make(map[int]map[int]bgCell)
	if b.width == 0 || b.height == 0 {
		return
	}

	gray := [3]uint8{50, 50, 45}
	speechClr := [3]uint8{65, 60, 55}

	putChar := func(row, c int, ch rune, clr [3]uint8) {
		if c >= 0 && c < b.width && row >= 0 && row < b.height {
			if b.charGrid[row] == nil {
				b.charGrid[row] = make(map[int]bgCell)
			}
			b.charGrid[row][c] = bgCell{ch: ch, fg: clr}
		}
	}

	for i := range b.npcs {
		n := &b.npcs[i]
		n.x += float64(n.dir) * n.speed
		n.frame++

		if n.x < 0 {
			n.x = 0
			n.dir = 1
		}
		if n.x >= float64(b.width-1) {
			n.x = float64(b.width - 2)
			n.dir = -1
		}

		// Speech bubble logic
		if n.speechTimer <= 0 && b.rng.Float64() < 0.005 {
			n.speech = npcSpeeches[b.rng.Intn(len(npcSpeeches))]
			n.speechTimer = 20 + b.rng.Intn(15)
		}
		if n.speechTimer > 0 {
			n.speechTimer--
		}

		col := int(n.x)

		// Stick figure: O head, | body, arms, legs
		putChar(n.y-2, col, 'O', gray)
		putChar(n.y-1, col, '|', gray)

		// Arms based on direction
		if n.dir > 0 {
			putChar(n.y-1, col-1, '-', gray)
			putChar(n.y-1, col+1, '-', gray)
		} else {
			putChar(n.y-1, col-1, '-', gray)
			putChar(n.y-1, col+1, '-', gray)
		}

		// Legs walking animation
		if (n.frame/8)%2 == 0 {
			putChar(n.y, col-1, '/', gray)
			putChar(n.y, col+1, '\\', gray)
		} else {
			putChar(n.y, col, '|', gray)
		}

		// Speech bubble
		if n.speechTimer > 0 && n.speech != "" {
			row := n.y - 3
			startX := col - len(n.speech)/2
			for j, ch := range n.speech {
				putChar(row, startX+j, ch, speechClr)
			}
		}
	}
}

// ============================================================
// OHIO — Chaotic glitch energy. Noise, tears, scary faces, flashes.
// ============================================================

func (b *BackgroundModel) initOhio() {
	b.ohioGlitches = nil
	b.ohioFlash = 0
}

func (b *BackgroundModel) updateOhio() {
	b.ensurePixelBuffer()
	b.pb.clear()

	if b.width == 0 || b.height == 0 {
		return
	}

	pH := b.height * 2

	// Dark base with random noise
	for py := 0; py < pH; py++ {
		for x := 0; x < b.width; x++ {
			noise := uint8(b.rng.Intn(8))
			b.pb.set(x, py, noise, noise, noise+3)
		}
	}

	// Spawn horizontal glitch tears
	if b.rng.Float64() < 0.08 {
		b.ohioGlitches = append(b.ohioGlitches, ohioGlitch{
			y:     b.rng.Intn(pH),
			shift: b.rng.Intn(20) - 10,
			life:  3 + b.rng.Intn(5),
		})
	}

	// Render glitch tears
	alive := b.ohioGlitches[:0]
	for _, g := range b.ohioGlitches {
		g.life--
		if g.life <= 0 {
			continue
		}
		alive = append(alive, g)

		// Bright colored tear stripe
		for x := 0; x < b.width; x++ {
			r := uint8(30 + b.rng.Intn(40))
			b.pb.set(x, g.y, r, uint8(b.rng.Intn(15)), uint8(b.rng.Intn(20)))
			if g.y+1 < pH {
				b.pb.set(x, g.y+1, r/2, 0, uint8(b.rng.Intn(10)))
			}
		}
	}
	b.ohioGlitches = alive

	// Flash effect (rare but dramatic)
	if b.ohioFlash > 0 {
		b.ohioFlash--
		v := uint8(20 + b.ohioFlash*8)
		for py := 0; py < pH; py++ {
			for x := 0; x < b.width; x++ {
				p := b.pb.get(x, py)
				b.pb.set(x, py,
					uint8(minInt(255, int(p.r)+int(v))),
					uint8(minInt(255, int(p.g)+int(v))),
					uint8(minInt(255, int(p.b)+int(v))))
			}
		}
	} else if b.rng.Float64() < 0.01 {
		b.ohioFlash = 3
	}

	// Scary face-like patterns (occasional)
	if b.rng.Float64() < 0.02 {
		fx := b.rng.Intn(maxInt(1, b.width-6)) + 3
		fy := b.rng.Intn(maxInt(1, pH-8)) + 4
		red := uint8(60 + b.rng.Intn(40))
		// Eyes
		b.pb.set(fx-1, fy, red, 0, 0)
		b.pb.set(fx+1, fy, red, 0, 0)
		b.pb.set(fx-1, fy+1, red/2, 0, 0)
		b.pb.set(fx+1, fy+1, red/2, 0, 0)
		// Mouth
		for dx := -1; dx <= 1; dx++ {
			b.pb.set(fx+dx, fy+3, red/2, 0, 0)
		}
		b.pb.set(fx-2, fy+2, red/3, 0, 0)
		b.pb.set(fx+2, fy+2, red/3, 0, 0)
	}

	// Random inverted color blocks
	if b.rng.Float64() < 0.05 {
		bx := b.rng.Intn(maxInt(1, b.width-4))
		by := b.rng.Intn(maxInt(1, pH-4))
		for dy := 0; dy < 4; dy++ {
			for dx := 0; dx < 4; dx++ {
				v := uint8(20 + b.rng.Intn(30))
				b.pb.set(bx+dx, by+dy, v, v/2, v)
			}
		}
	}
}

// ============================================================
// RIZZ — Sparkles floating upward. Pink/purple gradient. Romantic.
// ============================================================

func (b *BackgroundModel) initRizz() {
	if b.width == 0 || b.height == 0 {
		return
	}
	n := 40 + b.width/3
	b.rizzSparkles = make([]rizzSparkle, n)
	for i := range b.rizzSparkles {
		b.rizzSparkles[i] = rizzSparkle{
			x:  float64(b.rng.Intn(b.width)),
			y:  float64(b.rng.Intn(b.height * 2)),
			vy: -(0.3 + b.rng.Float64()*0.8),
		}
	}
}

func (b *BackgroundModel) updateRizz() {
	b.ensurePixelBuffer()
	b.pb.clear()

	if b.width == 0 || b.height == 0 {
		return
	}

	pH := b.height * 2

	// Pink/purple gradient background
	for py := 0; py < pH; py++ {
		fy := float64(py) / float64(pH)
		for x := 0; x < b.width; x++ {
			// Gradient: deep purple (top) → warm pink (bottom)
			cr := uint8(15 + 10*fy)
			cg := uint8(5 + 3*fy)
			cb := uint8(22 - 8*fy)
			b.pb.set(x, py, cr, cg, cb)
		}
	}

	// Sparkles floating upward
	for i := range b.rizzSparkles {
		s := &b.rizzSparkles[i]
		s.y += s.vy

		// Respawn at bottom
		if s.y < 0 {
			s.y = float64(pH - 1)
			s.x = float64(b.rng.Intn(b.width))
			s.vy = -(0.3 + b.rng.Float64()*0.8)
		}

		ix, iy := int(s.x), int(s.y)

		// Shimmer effect using sin wave
		shimmer := (b.fastSin(float64(b.frame)*0.2+float64(i)*0.5) + 1.0) / 2.0
		bright := shimmer * 0.5

		cr := uint8(180 * bright)
		cg := uint8(80 * bright)
		cb := uint8(160 * bright)
		b.pb.set(ix, iy, cr, cg, cb)

		// Glow on bright sparkles
		if shimmer > 0.7 {
			dim := bright * 0.4
			dr := uint8(180 * dim)
			dg := uint8(80 * dim)
			db := uint8(160 * dim)
			b.pb.set(ix+1, iy, dr, dg, db)
			b.pb.set(ix-1, iy, dr, dg, db)
			b.pb.set(ix, iy+1, dr, dg, db)
			b.pb.set(ix, iy-1, dr, dg, db)
		}
	}

	// Rose petals drifting down (slower, larger, warmer)
	for i := 0; i < 8; i++ {
		px := int(b.fastSin(float64(b.frame)*0.03+float64(i)*1.2)*float64(b.width)/3.0) + b.width/2
		py := (b.frame*2/3 + i*b.height*2/8) % pH
		r := uint8(60 + i*5)
		g := uint8(15)
		bv := uint8(25)
		b.pb.set(px, py, r, g, bv)
		b.pb.set(px+1, py, r, g, bv)
		b.pb.set(px, py+1, r-10, g, bv)
	}
}

// ============================================================
// GYATT — Impact text. Screen shake. Brief white flash. Bass energy.
// ============================================================

var gyattWords = []string{"GYATT", "SHEESH", "DAYUM", "GYATT"}

func (b *BackgroundModel) initGyatt() {
	b.gyattTexts = nil
}

func (b *BackgroundModel) updateGyatt() {
	b.charGrid = make(map[int]map[int]bgCell)
	if b.width == 0 || b.height == 0 {
		return
	}

	// Spawn new impact text
	if b.rng.Float64() < 0.06 {
		word := gyattWords[b.rng.Intn(len(gyattWords))]
		b.gyattTexts = append(b.gyattTexts, gyattText{
			x:       b.rng.Intn(maxInt(1, b.width-len(word))),
			y:       b.rng.Intn(maxInt(1, b.height)),
			life:    0,
			maxLife: 15 + b.rng.Intn(10),
		})
	}

	alive := b.gyattTexts[:0]
	for _, gt := range b.gyattTexts {
		gt.life++
		if gt.life >= gt.maxLife {
			continue
		}
		alive = append(alive, gt)

		fade := 1.0 - float64(gt.life)/float64(gt.maxLife)

		// Screen shake offset on fresh impacts
		shakeX, shakeY := 0, 0
		if gt.life < 5 {
			shakeX = b.rng.Intn(3) - 1
			shakeY = b.rng.Intn(3) - 1
		}

		text := gyattWords[0] // "GYATT"
		var intensity uint8
		if gt.life < gt.maxLife/3 {
			// Impact flash phase — bright white
			intensity = uint8(fade * 80)
			for j, ch := range text {
				x := gt.x + j + shakeX
				y := gt.y + shakeY
				if x >= 0 && x < b.width && y >= 0 && y < b.height {
					if b.charGrid[y] == nil {
						b.charGrid[y] = make(map[int]bgCell)
					}
					b.charGrid[y][x] = bgCell{ch: ch, fg: [3]uint8{intensity, intensity, intensity}}
				}
			}
		} else {
			// Fade out — warm red
			intensity = uint8(fade * 50)
			for j, ch := range text {
				x := gt.x + j + shakeX
				y := gt.y + shakeY
				if x >= 0 && x < b.width && y >= 0 && y < b.height {
					if b.charGrid[y] == nil {
						b.charGrid[y] = make(map[int]bgCell)
					}
					b.charGrid[y][x] = bgCell{ch: ch, fg: [3]uint8{intensity, intensity / 3, intensity / 4}}
				}
			}
		}

		// Impact frame: brief horizontal line flash
		if gt.life == 1 {
			y := gt.y
			if y >= 0 && y < b.height {
				if b.charGrid[y] == nil {
					b.charGrid[y] = make(map[int]bgCell)
				}
				for x := 0; x < b.width; x++ {
					if _, exists := b.charGrid[y][x]; !exists {
						b.charGrid[y][x] = bgCell{ch: '-', fg: [3]uint8{35, 35, 35}}
					}
				}
			}
		}
	}
	b.gyattTexts = alive
}

// ============================================================
// AMOGUS — Among Us crewmates walking. "SUS" text. Emergency flash.
// ============================================================

var crewColors = [][3]uint8{
	{140, 30, 30}, {30, 30, 140}, {30, 140, 30}, {140, 140, 30},
	{140, 30, 140}, {30, 140, 140}, {140, 80, 30}, {180, 50, 50},
}

func (b *BackgroundModel) initAmogus() {
	if b.width == 0 || b.height == 0 {
		return
	}
	n := 4 + b.width/30
	b.amogusCrews = make([]amogusCrew, n)
	for i := range b.amogusCrews {
		dir := 1
		if b.rng.Intn(2) == 0 {
			dir = -1
		}
		b.amogusCrews[i] = amogusCrew{
			x:     float64(b.rng.Intn(maxInt(1, b.bb.pixW))),
			y:     b.rng.Intn(maxInt(1, b.height)),
			dir:   dir,
			speed: 0.3 + b.rng.Float64()*0.5,
			color: crewColors[b.rng.Intn(len(crewColors))],
			frame: b.rng.Intn(10),
		}
	}
	b.amogusSus = 0
}

func (b *BackgroundModel) updateAmogus() {
	b.ensureBrailleBuffer()
	b.bb.clear()

	if b.width == 0 || b.height == 0 {
		return
	}

	for i := range b.amogusCrews {
		c := &b.amogusCrews[i]
		c.x += float64(c.dir) * c.speed
		c.frame++

		// Wrap around
		maxX := float64(b.bb.pixW + 10)
		if c.x < -10 {
			c.x = maxX
		}
		if c.x > maxX {
			c.x = -10
		}

		ix := int(c.x)
		iy := c.y * 4

		// Dim the crew color for readability
		dr := uint8(float64(c.color[0]) * 0.35)
		dg := uint8(float64(c.color[1]) * 0.35)
		db := uint8(float64(c.color[2]) * 0.35)

		// Visor color (lighter blue)
		vr := uint8(30)
		vg := uint8(50)
		vb := uint8(65)

		// Draw crewmate bean shape (~4 wide, ~6 tall in braille pixels)
		// Head
		b.bb.set(ix+1, iy, dr, dg, db)
		b.bb.set(ix+2, iy, dr, dg, db)

		// Visor
		b.bb.set(ix+2, iy+1, vr, vg, vb)
		b.bb.set(ix+3, iy+1, vr, vg, vb)

		// Body
		for dy := 1; dy <= 4; dy++ {
			b.bb.set(ix, iy+dy, dr, dg, db)
			b.bb.set(ix+1, iy+dy, dr, dg, db)
			b.bb.set(ix+2, iy+dy, dr, dg, db)
		}

		// Backpack
		b.bb.set(ix-1, iy+2, dr, dg, db)
		b.bb.set(ix-1, iy+3, dr, dg, db)

		// Legs (walking animation)
		if (c.frame/5)%2 == 0 {
			b.bb.set(ix, iy+5, dr, dg, db)
			b.bb.set(ix+2, iy+5, dr, dg, db)
		} else {
			b.bb.set(ix+1, iy+5, dr, dg, db)
		}
	}

	// "SUS" text in braille (occasional)
	b.amogusSus--
	if b.amogusSus <= 0 && b.rng.Float64() < 0.01 {
		b.amogusSus = 30
	}
	if b.amogusSus > 20 {
		sx := b.bb.pixW/2 - 6
		sy := b.bb.pixH/2 - 2
		red := uint8(50)
		// S
		b.bb.set(sx, sy, red, 0, 0)
		b.bb.set(sx+1, sy, red, 0, 0)
		b.bb.set(sx, sy+1, red, 0, 0)
		b.bb.set(sx, sy+2, red, 0, 0)
		b.bb.set(sx+1, sy+2, red, 0, 0)
		b.bb.set(sx+1, sy+3, red, 0, 0)
		b.bb.set(sx, sy+4, red, 0, 0)
		b.bb.set(sx+1, sy+4, red, 0, 0)
		// U
		b.bb.set(sx+3, sy, red, 0, 0)
		b.bb.set(sx+5, sy, red, 0, 0)
		b.bb.set(sx+3, sy+1, red, 0, 0)
		b.bb.set(sx+5, sy+1, red, 0, 0)
		b.bb.set(sx+3, sy+2, red, 0, 0)
		b.bb.set(sx+5, sy+2, red, 0, 0)
		b.bb.set(sx+3, sy+3, red, 0, 0)
		b.bb.set(sx+5, sy+3, red, 0, 0)
		b.bb.set(sx+3, sy+4, red, 0, 0)
		b.bb.set(sx+4, sy+4, red, 0, 0)
		b.bb.set(sx+5, sy+4, red, 0, 0)
		// S
		b.bb.set(sx+7, sy, red, 0, 0)
		b.bb.set(sx+8, sy, red, 0, 0)
		b.bb.set(sx+7, sy+1, red, 0, 0)
		b.bb.set(sx+7, sy+2, red, 0, 0)
		b.bb.set(sx+8, sy+2, red, 0, 0)
		b.bb.set(sx+8, sy+3, red, 0, 0)
		b.bb.set(sx+7, sy+4, red, 0, 0)
		b.bb.set(sx+8, sy+4, red, 0, 0)
	}

	// Emergency meeting flash (rare)
	if b.rng.Float64() < 0.003 {
		for py := 0; py < b.bb.pixH; py += 3 {
			for px := 0; px < b.bb.pixW; px += 2 {
				b.bb.set(px, py, 40, 0, 0)
			}
		}
	}
}

// ============================================================
// BUSSIN — Fire at the bottom, warm orange/red. Actually kinda fire.
// ============================================================

func (b *BackgroundModel) initBussin() {
	if b.width == 0 || b.height == 0 {
		return
	}
	pH := b.height * 2
	b.fireHeat = make([]float64, b.width*pH)
	// Hot source at very bottom
	for x := 0; x < b.width; x++ {
		b.fireHeat[(pH-1)*b.width+x] = 1.0
	}
}

func bussinFireColor(heat float64) (uint8, uint8, uint8) {
	if heat < 0 {
		heat = 0
	}
	if heat > 1 {
		heat = 1
	}
	dim := 0.45
	if heat < 0.2 {
		t := heat / 0.2
		return uint8(50 * t * dim), uint8(20 * t * dim), 0
	} else if heat < 0.5 {
		t := (heat - 0.2) / 0.3
		return uint8((50 + 150*t) * dim), uint8((20 + 60*t) * dim), 0
	}
	t := (heat - 0.5) / 0.5
	return uint8((200 + 55*t) * dim), uint8((80 + 100*t) * dim), uint8(30 * t * dim)
}

func (b *BackgroundModel) updateBussin() {
	b.ensurePixelBuffer()

	if b.width == 0 || b.height == 0 {
		return
	}

	pH := b.height * 2
	w := b.width

	if len(b.fireHeat) != w*pH {
		b.initBussin()
	}

	// Hot source at bottom
	for x := 0; x < w; x++ {
		b.fireHeat[(pH-1)*w+x] = 0.6 + b.rng.Float64()*0.4
	}

	// Extra hot spots
	for k := 0; k < w/10; k++ {
		x := b.rng.Intn(w)
		b.fireHeat[(pH-1)*w+x] = 1.0
	}

	// Diffuse — only bottom 60% for a "floor fire" look
	startPy := pH * 2 / 5
	for py := startPy; py < pH-1; py++ {
		for x := 0; x < w; x++ {
			left := maxInt(0, x-1)
			right := minInt(w-1, x+1)
			below := minInt(pH-1, py+1)
			below2 := minInt(pH-1, py+2)
			avg := (b.fireHeat[below*w+left] + b.fireHeat[below*w+x] +
				b.fireHeat[below*w+right] + b.fireHeat[below2*w+x]) / 4.0
			avg -= 0.015 + 0.008*b.rng.Float64()
			if avg < 0 {
				avg = 0
			}
			b.fireHeat[py*w+x] = avg
		}
	}

	// Render with warm orange palette
	b.pb.clear()
	for py := 0; py < pH; py++ {
		for x := 0; x < w; x++ {
			heat := b.fireHeat[py*w+x]
			if heat > 0.01 {
				r, g, bv := bussinFireColor(heat)
				b.pb.set(x, py, r, g, bv)
			}
		}
	}
}
