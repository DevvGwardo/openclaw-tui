package tui

import "math"

// --- Aquarium (underwater fish tank screensaver — half-block, color-intensive) ---

type aquariumFish struct {
	x, y      float64
	speed     float64
	dir       float64 // -1 left, +1 right
	species   int     // determines shape and color
	size      int     // 0=small, 1=medium, 2=large
	wobble    float64 // vertical oscillation phase
	wobbleAmp float64
	depth     float64 // 0.0=back, 1.0=front (affects brightness)
}

type aquariumBubble struct {
	x, y   float64
	speed  float64
	size   int // 0=tiny, 1=small, 2=medium
	wobble float64
	drift  float64
}

type aquariumWeed struct {
	x      int
	height int
	phase  float64
	hue    int // 0=green, 1=dark green, 2=olive
}

func (b *BackgroundModel) initAquarium() {
	if b.width == 0 || b.height == 0 {
		return
	}

	// Create fish
	fishCount := b.width / 8
	if fishCount < 6 {
		fishCount = 6
	}
	if fishCount > 25 {
		fishCount = 25
	}
	b.aquariumFish = make([]aquariumFish, fishCount)
	for i := range b.aquariumFish {
		b.aquariumFish[i] = b.newAquariumFish(true)
	}

	// Create bubbles
	bubbleCount := b.width / 12
	if bubbleCount < 4 {
		bubbleCount = 4
	}
	if bubbleCount > 15 {
		bubbleCount = 15
	}
	b.aquariumBubbles = make([]aquariumBubble, bubbleCount)
	for i := range b.aquariumBubbles {
		b.aquariumBubbles[i] = b.newAquariumBubble(true)
	}

	// Create seaweed
	weedCount := b.width / 10
	if weedCount < 3 {
		weedCount = 3
	}
	if weedCount > 12 {
		weedCount = 12
	}
	b.aquariumWeeds = make([]aquariumWeed, weedCount)
	for i := range b.aquariumWeeds {
		b.aquariumWeeds[i] = aquariumWeed{
			x:      b.rng.Intn(b.width),
			height: 3 + b.rng.Intn(b.height/3),
			phase:  b.rng.Float64() * math.Pi * 2,
			hue:    b.rng.Intn(3),
		}
	}
}

func (b *BackgroundModel) newAquariumFish(randomX bool) aquariumFish {
	dir := 1.0
	if b.rng.Intn(2) == 0 {
		dir = -1.0
	}

	x := 0.0
	if randomX {
		x = b.rng.Float64() * float64(b.width)
	} else {
		if dir > 0 {
			x = -4.0
		} else {
			x = float64(b.width) + 4.0
		}
	}

	size := 0
	r := b.rng.Float64()
	if r < 0.15 {
		size = 2 // large
	} else if r < 0.45 {
		size = 1 // medium
	}

	return aquariumFish{
		x:         x,
		y:         1.0 + b.rng.Float64()*float64(b.height*2-4),
		speed:     0.15 + b.rng.Float64()*0.5,
		dir:       dir,
		species:   b.rng.Intn(6),
		size:      size,
		wobble:    b.rng.Float64() * math.Pi * 2,
		wobbleAmp: 0.3 + b.rng.Float64()*0.8,
		depth:     0.3 + b.rng.Float64()*0.7,
	}
}

func (b *BackgroundModel) newAquariumBubble(randomY bool) aquariumBubble {
	y := float64(b.height*2 - 1)
	if randomY {
		y = b.rng.Float64() * float64(b.height*2)
	}
	return aquariumBubble{
		x:      b.rng.Float64() * float64(b.width),
		y:      y,
		speed:  0.3 + b.rng.Float64()*0.6,
		size:   b.rng.Intn(3),
		wobble: b.rng.Float64() * math.Pi * 2,
		drift:  (b.rng.Float64() - 0.5) * 0.3,
	}
}

func (b *BackgroundModel) updateAquarium() {
	b.ensurePixelBuffer()
	b.pb.clear()

	if b.width == 0 || b.height == 0 {
		return
	}

	t := float64(b.frame) * 0.04
	pH := b.height * 2
	w := b.width

	// --- Water background with light caustics ---
	for py := 0; py < pH; py++ {
		fy := float64(py) / float64(pH)
		for x := 0; x < w; x++ {
			fx := float64(x) / float64(w)

			// Deep blue gradient getting darker toward bottom
			baseR := 2.0 + 8.0*(1.0-fy)
			baseG := 15.0 + 30.0*(1.0-fy)
			baseB := 40.0 + 50.0*(1.0-fy)

			// Caustic light patterns (rippling light on water surface)
			c1 := b.fastSin(fx*8.0*math.Pi+t*1.2) * b.fastSin(fy*6.0*math.Pi+t*0.8)
			c2 := b.fastSin((fx+fy)*5.0*math.Pi+t*0.6) * b.fastSin(fx*12.0*math.Pi-t*1.0)
			c3 := b.fastSin(fx*3.0*math.Pi+b.fastSin(fy*4.0*math.Pi+t)*2.0) * 0.5

			caustic := (c1 + c2*0.6 + c3*0.4) / 2.0
			if caustic < 0 {
				caustic = 0
			}

			// Caustics are stronger near the top (closer to light source)
			causticStrength := (1.0 - fy) * 0.5
			caustic *= causticStrength

			// Volumetric light shafts from above
			shaft := b.fastSin(fx*4.0*math.Pi+t*0.3) * 0.5
			if shaft > 0.2 {
				shaftIntensity := (shaft - 0.2) * (1.0 - fy) * 0.15
				baseR += shaftIntensity * 60
				baseG += shaftIntensity * 90
				baseB += shaftIntensity * 50
			}

			cr := uint8(clampF(baseR+caustic*50, 0, 255))
			cg := uint8(clampF(baseG+caustic*80, 0, 255))
			cb := uint8(clampF(baseB+caustic*40, 0, 255))

			b.pb.set(x, py, cr, cg, cb)
		}
	}

	// --- Sandy bottom ---
	sandStart := pH - 3
	for py := sandStart; py < pH; py++ {
		sandDepth := float64(py-sandStart) / 3.0
		for x := 0; x < w; x++ {
			noise := b.fastSin(float64(x)*0.5+float64(py)*0.3) * 0.15
			sr := uint8(clampF(50.0+35.0*sandDepth+noise*30, 0, 255))
			sg := uint8(clampF(40.0+25.0*sandDepth+noise*20, 0, 255))
			sb := uint8(clampF(15.0+10.0*sandDepth+noise*10, 0, 255))
			b.pb.set(x, py, sr, sg, sb)
		}
	}

	// --- Seaweed ---
	for _, weed := range b.aquariumWeeds {
		baseY := pH - 3
		for seg := 0; seg < weed.height; seg++ {
			sway := b.fastSin(weed.phase+t*0.8+float64(seg)*0.4) * float64(seg) * 0.3
			py := baseY - seg
			px := weed.x + int(sway)
			if px < 0 || px >= w || py < 0 || py >= pH {
				continue
			}

			var wr, wg, wb uint8
			fade := 1.0 - float64(seg)/float64(weed.height)*0.4
			switch weed.hue {
			case 0: // bright green
				wr = uint8(10 * fade)
				wg = uint8(clampF(80*fade+30*b.fastSin(float64(seg)*0.5+t), 0, 255))
				wb = uint8(15 * fade)
			case 1: // dark green
				wr = uint8(5 * fade)
				wg = uint8(clampF(55*fade+20*b.fastSin(float64(seg)*0.6+t*1.1), 0, 255))
				wb = uint8(20 * fade)
			default: // olive
				wr = uint8(30 * fade)
				wg = uint8(clampF(60*fade+20*b.fastSin(float64(seg)*0.4+t*0.9), 0, 255))
				wb = uint8(10 * fade)
			}
			b.pb.set(px, py, wr, wg, wb)
			// Thicken at base
			if seg < weed.height/2 {
				b.pb.set(px+1, py, wr, wg, wb)
			}
		}
	}

	// --- Fish ---
	for i := range b.aquariumFish {
		fish := &b.aquariumFish[i]

		fish.x += fish.speed * fish.dir
		fish.wobble += 0.08
		yOff := b.fastSin(fish.wobble) * fish.wobbleAmp

		// Respawn if off-screen
		if fish.dir > 0 && fish.x > float64(w)+8 {
			b.aquariumFish[i] = b.newAquariumFish(false)
			continue
		}
		if fish.dir < 0 && fish.x < -8 {
			b.aquariumFish[i] = b.newAquariumFish(false)
			continue
		}

		fx := int(fish.x)
		fy := int(fish.y + yOff)
		brightness := 0.5 + fish.depth*0.5

		// Fish colors based on species
		var fr, fg, fb uint8
		switch fish.species {
		case 0: // clownfish (orange/white)
			fr = uint8(clampF(240*brightness, 0, 255))
			fg = uint8(clampF(130*brightness, 0, 255))
			fb = uint8(clampF(30*brightness, 0, 255))
		case 1: // blue tang
			fr = uint8(clampF(30*brightness, 0, 255))
			fg = uint8(clampF(100*brightness, 0, 255))
			fb = uint8(clampF(230*brightness, 0, 255))
		case 2: // angelfish (yellow)
			fr = uint8(clampF(230*brightness, 0, 255))
			fg = uint8(clampF(210*brightness, 0, 255))
			fb = uint8(clampF(50*brightness, 0, 255))
		case 3: // neon tetra (red/blue)
			fr = uint8(clampF(200*brightness, 0, 255))
			fg = uint8(clampF(40*brightness, 0, 255))
			fb = uint8(clampF(80*brightness, 0, 255))
		case 4: // green chromis
			fr = uint8(clampF(60*brightness, 0, 255))
			fg = uint8(clampF(200*brightness, 0, 255))
			fb = uint8(clampF(120*brightness, 0, 255))
		default: // purple fish
			fr = uint8(clampF(160*brightness, 0, 255))
			fg = uint8(clampF(70*brightness, 0, 255))
			fb = uint8(clampF(200*brightness, 0, 255))
		}

		// Draw fish based on size and direction
		b.drawFish(fx, fy, fish.size, fish.dir, fr, fg, fb, t, fish.species)
	}

	// --- Bubbles ---
	for i := range b.aquariumBubbles {
		bub := &b.aquariumBubbles[i]

		bub.y -= bub.speed
		bub.wobble += 0.06
		xOff := b.fastSin(bub.wobble) * 0.8

		// Respawn if off-screen
		if bub.y < -2 {
			b.aquariumBubbles[i] = b.newAquariumBubble(false)
			continue
		}

		bx := int(bub.x + xOff + bub.drift*float64(b.frame))
		by := int(bub.y)

		// Wrap horizontal
		bx = ((bx % w) + w) % w

		// Bubble brightness based on depth
		depthFrac := bub.y / float64(pH)
		bright := 0.3 + (1.0-depthFrac)*0.4

		br := uint8(clampF(120*bright, 0, 255))
		bg := uint8(clampF(200*bright, 0, 255))
		bb := uint8(clampF(255*bright, 0, 255))

		switch bub.size {
		case 0: // tiny - single pixel
			b.pb.set(bx, by, br, bg, bb)
		case 1: // small - 2 pixels
			b.pb.set(bx, by, br, bg, bb)
			b.pb.set(bx+1, by, br, bg, bb)
		case 2: // medium - small circle
			b.pb.set(bx, by, br, bg, bb)
			b.pb.set(bx+1, by, br, bg, bb)
			b.pb.set(bx, by-1, br, bg, bb)
			b.pb.set(bx+1, by-1, br, bg, bb)
			// highlight
			hr := uint8(clampF(200*bright, 0, 255))
			hg := uint8(clampF(240*bright, 0, 255))
			hb := uint8(clampF(255*bright, 0, 255))
			b.pb.set(bx, by-1, hr, hg, hb)
		}
	}

	// --- Occasional new bubble from bottom ---
	if b.rng.Float64() < 0.08 {
		b.aquariumBubbles = append(b.aquariumBubbles, b.newAquariumBubble(false))
		// Cap bubble count
		if len(b.aquariumBubbles) > 30 {
			b.aquariumBubbles = b.aquariumBubbles[1:]
		}
	}
}

// drawFish renders a fish at the given position.
func (b *BackgroundModel) drawFish(x, y, size int, dir float64, r, g, bv uint8, t float64, species int) {
	pH := b.height * 2
	w := b.width
	if y < 0 || y >= pH {
		return
	}

	// Dimmer body color for body variation
	br := uint8(clampF(float64(r)*0.7, 0, 255))
	bg := uint8(clampF(float64(g)*0.7, 0, 255))
	bb := uint8(clampF(float64(bv)*0.7, 0, 255))

	// Tail color
	tr := uint8(clampF(float64(r)*0.5, 0, 255))
	tg := uint8(clampF(float64(g)*0.5, 0, 255))
	tb := uint8(clampF(float64(bv)*0.5, 0, 255))

	// Eye
	er, eg, eb := uint8(220), uint8(220), uint8(230)

	// Tail animation
	tailWag := b.fastSin(t*4.0+float64(x)*0.1) * 0.5

	setP := func(dx, dy int, cr, cg, cb uint8) {
		px := x + dx
		py := y + dy
		if px >= 0 && px < w && py >= 0 && py < pH {
			b.pb.set(px, py, cr, cg, cb)
		}
	}

	d := int(dir)

	switch size {
	case 0: // small fish: 3 wide
		// Body
		setP(0, 0, r, g, bv)
		setP(d, 0, r, g, bv)
		// Tail
		setP(-d, 0, tr, tg, tb)
		// Eye
		setP(d, 0, er, eg, eb)

	case 1: // medium fish: 5 wide
		// Body
		setP(0, 0, r, g, bv)
		setP(d, 0, r, g, bv)
		setP(d*2, 0, r, g, bv)
		setP(0, -1, br, bg, bb)
		setP(d, -1, br, bg, bb)
		setP(0, 1, br, bg, bb)
		setP(d, 1, br, bg, bb)
		// Tail
		tailOff := int(tailWag)
		setP(-d, 0, tr, tg, tb)
		setP(-d*2, -1+tailOff, tr, tg, tb)
		setP(-d*2, 1+tailOff, tr, tg, tb)
		// Eye
		setP(d*2, -1, er, eg, eb)
		// Stripe for clownfish
		if species == 0 {
			setP(0, 0, 230, 230, 230)
			setP(0, -1, 230, 230, 230)
			setP(0, 1, 230, 230, 230)
		}

	case 2: // large fish: 7 wide
		// Body center
		for dx := -1; dx <= 2; dx++ {
			setP(d*dx, 0, r, g, bv)
			setP(d*dx, -1, br, bg, bb)
			setP(d*dx, 1, br, bg, bb)
		}
		// Top/bottom fins
		setP(0, -2, tr, tg, tb)
		setP(d, -2, tr, tg, tb)
		setP(0, 2, tr, tg, tb)
		// Head
		setP(d*3, 0, r, g, bv)
		setP(d*3, -1, br, bg, bb)
		// Tail
		tailOff := int(tailWag)
		setP(-d*2, 0, tr, tg, tb)
		setP(-d*3, -1+tailOff, tr, tg, tb)
		setP(-d*3, 0, tr, tg, tb)
		setP(-d*3, 1+tailOff, tr, tg, tb)
		setP(-d*4, -2+tailOff, tr, tg, tb)
		setP(-d*4, 2+tailOff, tr, tg, tb)
		// Eye
		setP(d*3, -1, er, eg, eb)
		// Stripes for clownfish
		if species == 0 {
			for dy := -1; dy <= 1; dy++ {
				setP(0, dy, 230, 230, 230)
				setP(d*2, dy, 230, 230, 230)
			}
		}
	}
}

