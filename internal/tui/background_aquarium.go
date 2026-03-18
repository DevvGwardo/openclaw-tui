package tui

import "math"

// Crab pixel art reference: Elthen (https://elthen.itch.io/2d-pixel-art-crab-sprites) - CC-BY compatible

// --- Aquarium (underwater fish tank screensaver — half-block, color-intensive) ---

type aquariumFood struct {
	x, y      float64
	sinkSpeed float64 // how fast it sinks
	wobble    float64 // horizontal drift phase
	size      int     // 0=small, 1=medium
	eaten     bool    // marked for removal
	age       int     // ticks since dropped
}

type aquariumCrab struct {
	x, y       float64
	speed      float64
	dir        float64 // -1 left, +1 right
	legPhase   float64 // leg animation
	task       string  // task label displayed above the crab
	variant    int     // 0=red, 1=orange, 2=purple
	animState  int     // 0=idle, 1=walk, 2=attack(pinch), 3=death
	animFrame  int     // current frame within the animation (0-3)
	animTimer  float64 // accumulates to advance frames
}

type aquariumFish struct {
	x, y      float64
	speed     float64
	dir       float64 // -1 left, +1 right
	species   int     // determines shape and color
	size      int     // 0=small, 1=medium, 2=large
	wobble    float64 // vertical oscillation phase
	wobbleAmp float64
	depth     float64 // 0.0=back, 1.0=front (affects brightness)
	finPhase  float64 // pectoral fin animation phase
	bodyPhase float64 // body undulation phase
}

type aquariumBubble struct {
	x, y      float64
	speed     float64
	size      int // 0=tiny, 1=small, 2=medium, 3=large
	wobble    float64
	drift     float64
	age       int     // ticks since spawn
	squish    float64 // deformation phase (makes bubble wobble/breathe)
	shimmer   float64 // highlight rotation phase
	opacity   float64 // 0-1, fades in on spawn and out near top
	splitting bool    // true when about to split into smaller bubbles
}

type aquariumWeed struct {
	x      int
	height int
	phase  float64
	hue    int // 0=green, 1=dark green, 2=olive
}

// fishColors holds the color palette for a fish species.
type fishColors struct {
	bodyR, bodyG, bodyB       float64 // main body
	bellyR, bellyG, bellyB   float64 // lighter belly
	finR, finG, finB          float64 // fin/tail
	accentR, accentG, accentB float64 // stripes/markings
	hasStripes                bool
	stripeCount               int
	hasDots                   bool
}

var speciesColors = []fishColors{
	// 0: Clownfish (orange with white stripes)
	{240, 130, 20, 255, 180, 80, 200, 90, 10, 240, 240, 240, true, 3, false},
	// 1: Blue tang (royal blue with yellow tail)
	{30, 80, 220, 60, 120, 240, 220, 200, 50, 20, 40, 100, false, 0, false},
	// 2: Angelfish (yellow/silver with black stripes)
	{230, 210, 80, 240, 230, 150, 200, 180, 40, 30, 30, 30, true, 2, false},
	// 3: Neon tetra (red/blue split with light stripe)
	{180, 30, 50, 200, 60, 70, 160, 20, 40, 100, 220, 255, false, 0, false},
	// 4: Green chromis (emerald green)
	{40, 190, 100, 80, 220, 150, 30, 150, 70, 100, 240, 160, false, 0, false},
	// 5: Royal gramma (purple/yellow split)
	{150, 40, 180, 220, 190, 40, 120, 30, 150, 180, 160, 30, false, 0, false},
}

func (b *BackgroundModel) initAquarium() {
	if b.width == 0 || b.height == 0 {
		return
	}

	fishCount := b.width / 10
	if fishCount < 5 {
		fishCount = 5
	}
	if fishCount > 20 {
		fishCount = 20
	}
	b.aquariumFish = make([]aquariumFish, fishCount)
	for i := range b.aquariumFish {
		b.aquariumFish[i] = b.newAquariumFish(true)
	}

	bubbleCount := b.width / 20
	if bubbleCount < 2 {
		bubbleCount = 2
	}
	if bubbleCount > 8 {
		bubbleCount = 8
	}
	b.aquariumBubbles = make([]aquariumBubble, bubbleCount)
	for i := range b.aquariumBubbles {
		b.aquariumBubbles[i] = b.newAquariumBubble(true)
	}

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
			height: 4 + b.rng.Intn(b.height/3),
			phase:  b.rng.Float64() * math.Pi * 2,
			hue:    b.rng.Intn(3),
		}
	}

	// Initialize crabs from current tasks
	b.syncCrabs()
}

func (b *BackgroundModel) newAquariumFish(randomX bool) aquariumFish {
	dir := 1.0
	if b.rng.Intn(2) == 0 {
		dir = -1.0
	}

	size := 0
	r := b.rng.Float64()
	if r < 0.20 {
		size = 2 // large
	} else if r < 0.55 {
		size = 1 // medium
	}

	// Fish positions in pixel-buffer space (height*2)
	margin := 6 + size*4
	x := 0.0
	if randomX {
		x = b.rng.Float64() * float64(b.width)
	} else {
		if dir > 0 {
			x = float64(-margin)
		} else {
			x = float64(b.width + margin)
		}
	}

	return aquariumFish{
		x:         x,
		y:         3.0 + b.rng.Float64()*float64(b.height*2-margin-6),
		speed:     0.12 + b.rng.Float64()*0.4,
		dir:       dir,
		species:   b.rng.Intn(len(speciesColors)),
		size:      size,
		wobble:    b.rng.Float64() * math.Pi * 2,
		wobbleAmp: 0.4 + b.rng.Float64()*1.0,
		depth:     0.3 + b.rng.Float64()*0.7,
		finPhase:  b.rng.Float64() * math.Pi * 2,
		bodyPhase: b.rng.Float64() * math.Pi * 2,
	}
}

func (b *BackgroundModel) newAquariumBubble(randomY bool) aquariumBubble {
	y := float64(b.height*2 - 1)
	if randomY {
		y = b.rng.Float64() * float64(b.height*2)
	}
	size := b.rng.Intn(4) // 0=tiny, 1=small, 2=medium, 3=large
	// Larger bubbles are rarer
	if size == 3 && b.rng.Float64() > 0.15 {
		size = 2
	}
	return aquariumBubble{
		x:       b.rng.Float64() * float64(b.width),
		y:       y,
		speed:   0.2 + b.rng.Float64()*0.5 + float64(size)*0.08,
		size:    size,
		wobble:  b.rng.Float64() * math.Pi * 2,
		drift:   (b.rng.Float64() - 0.5) * 0.3,
		squish:  b.rng.Float64() * math.Pi * 2,
		shimmer: b.rng.Float64() * math.Pi * 2,
		opacity: 0.0, // fades in
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
			baseR := 2.0 + 10.0*(1.0-fy)
			baseG := 12.0 + 28.0*(1.0-fy)
			baseB := 35.0 + 55.0*(1.0-fy)

			// Caustic light patterns
			c1 := b.fastSin(fx*8.0*math.Pi+t*1.2) * b.fastSin(fy*6.0*math.Pi+t*0.8)
			c2 := b.fastSin((fx+fy)*5.0*math.Pi+t*0.6) * b.fastSin(fx*12.0*math.Pi-t*1.0)
			c3 := b.fastSin(fx*3.0*math.Pi+b.fastSin(fy*4.0*math.Pi+t)*2.0) * 0.5

			caustic := (c1 + c2*0.6 + c3*0.4) / 2.0
			if caustic < 0 {
				caustic = 0
			}
			causticStrength := (1.0 - fy) * 0.5
			caustic *= causticStrength

			// Volumetric light shafts
			shaft := b.fastSin(fx*4.0*math.Pi+t*0.3) * 0.5
			if shaft > 0.2 {
				si := (shaft - 0.2) * (1.0 - fy) * 0.15
				baseR += si * 60
				baseG += si * 90
				baseB += si * 50
			}

			cr := uint8(clampF(baseR+caustic*50, 0, 255))
			cg := uint8(clampF(baseG+caustic*80, 0, 255))
			cb := uint8(clampF(baseB+caustic*40, 0, 255))
			b.pb.set(x, py, cr, cg, cb)
		}
	}

	// --- Sandy bottom with pebbles ---
	sandStart := pH - 4
	for py := sandStart; py < pH; py++ {
		sandDepth := float64(py-sandStart) / 4.0
		for x := 0; x < w; x++ {
			noise := b.fastSin(float64(x)*0.5+float64(py)*0.3) * 0.15
			pebble := b.fastSin(float64(x)*2.3+float64(py)*1.7) * 0.1
			sr := uint8(clampF(45.0+40.0*sandDepth+noise*30+pebble*20, 0, 255))
			sg := uint8(clampF(35.0+30.0*sandDepth+noise*20+pebble*15, 0, 255))
			sb := uint8(clampF(12.0+12.0*sandDepth+noise*10+pebble*5, 0, 255))
			b.pb.set(x, py, sr, sg, sb)
		}
	}

	// --- Seaweed ---
	for _, weed := range b.aquariumWeeds {
		baseY := pH - 4
		for seg := 0; seg < weed.height; seg++ {
			sway := b.fastSin(weed.phase+t*0.8+float64(seg)*0.4) * float64(seg) * 0.3
			py := baseY - seg
			px := weed.x + int(sway)
			if px < 0 || px >= w || py < 0 || py >= pH {
				continue
			}
			fade := 1.0 - float64(seg)/float64(weed.height)*0.4
			var wr, wg, wb uint8
			switch weed.hue {
			case 0:
				wr = uint8(10 * fade)
				wg = uint8(clampF(85*fade+30*b.fastSin(float64(seg)*0.5+t), 0, 255))
				wb = uint8(15 * fade)
			case 1:
				wr = uint8(5 * fade)
				wg = uint8(clampF(60*fade+20*b.fastSin(float64(seg)*0.6+t*1.1), 0, 255))
				wb = uint8(22 * fade)
			default:
				wr = uint8(30 * fade)
				wg = uint8(clampF(65*fade+20*b.fastSin(float64(seg)*0.4+t*0.9), 0, 255))
				wb = uint8(10 * fade)
			}
			b.pb.set(px, py, wr, wg, wb)
			if seg < weed.height*2/3 {
				b.pb.set(px+1, py, wr, wg, wb)
			}
			if seg < weed.height/3 {
				b.pb.set(px-1, py, wr, wg, wb)
			}
		}
	}

	// --- Food particles (update and draw BEFORE fish) ---
	sandBottom := float64(pH - 4)
	for i := range b.aquariumFood {
		food := &b.aquariumFood[i]
		if food.eaten {
			continue
		}
		food.age++
		// Dissolve after 300 ticks
		if food.age > 300 {
			food.eaten = true
			continue
		}
		// Sink and wobble
		if food.y < sandBottom {
			food.y += food.sinkSpeed
			food.wobble += 0.08
			food.x += b.fastSin(food.wobble) * 0.3
		}
		// Draw food
		fx := int(food.x)
		fy := int(food.y)
		if fx >= 0 && fx < w && fy >= 0 && fy < pH {
			if food.size == 0 {
				b.pb.set(fx, fy, 180, 140, 60) // tan pellet
			} else {
				b.pb.set(fx, fy, 160, 80, 40)   // reddish flake
				b.pb.set(fx+1, fy, 160, 80, 40)
			}
		}
	}
	// Remove eaten food
	n := 0
	for i := range b.aquariumFood {
		if !b.aquariumFood[i].eaten {
			b.aquariumFood[n] = b.aquariumFood[i]
			n++
		}
	}
	b.aquariumFood = b.aquariumFood[:n]

	// --- Fish (sorted: far fish first, close fish on top) ---
	for i := range b.aquariumFish {
		fish := &b.aquariumFish[i]

		// Check for nearby food and steer toward it
		nearestDist := math.MaxFloat64
		nearestFood := -1
		for fi := range b.aquariumFood {
			if b.aquariumFood[fi].eaten {
				continue
			}
			dx := b.aquariumFood[fi].x - fish.x
			dy := b.aquariumFood[fi].y - fish.y
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist < 20.0 && dist < nearestDist {
				nearestDist = dist
				nearestFood = fi
			}
		}

		if nearestFood >= 0 {
			// Steer toward food at 1.5x speed
			food := &b.aquariumFood[nearestFood]
			dx := food.x - fish.x
			dy := food.y - fish.y
			spd := fish.speed * 1.5
			if dx > 0 {
				fish.dir = 1
			} else if dx < 0 {
				fish.dir = -1
			}
			// Move toward food
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist > 0 {
				fish.x += (dx / dist) * spd
				fish.y += (dy / dist) * spd
			}
			// Eat if close enough
			if dist < 2.0 {
				food.eaten = true
			}
		} else {
			fish.x += fish.speed * fish.dir
		}

		fish.wobble += 0.07
		fish.finPhase += 0.12
		fish.bodyPhase += 0.05
		yOff := b.fastSin(fish.wobble) * fish.wobbleAmp

		margin := 10 + fish.size*6
		if fish.dir > 0 && fish.x > float64(w+margin) {
			b.aquariumFish[i] = b.newAquariumFish(false)
			continue
		}
		if fish.dir < 0 && fish.x < float64(-margin) {
			b.aquariumFish[i] = b.newAquariumFish(false)
			continue
		}

		actualY := fish.y + yOff
		if nearestFood >= 0 {
			actualY = fish.y // suppress wobble when chasing food
		}
		b.drawDetailedFish(fish, fish.x, actualY, t)
	}

	// --- Crabs ---
	for i := range b.aquariumCrabs {
		crab := &b.aquariumCrabs[i]

		crab.x += crab.speed * crab.dir
		crab.legPhase += 0.15

		// Bounce off edges (crab is ~8px from center to claw tip)
		if crab.x > float64(w-9) {
			crab.dir = -1
		}
		if crab.x < 9 {
			crab.dir = 1
		}

		// Set animation state based on movement
		if crab.speed > 0 && crab.task != "" {
			crab.animState = 1 // walk
		} else {
			crab.animState = 0 // idle
		}

		// Advance animation frame every ~8 ticks
		crab.animTimer += 1
		if crab.animTimer >= 8 {
			crab.animTimer = 0
			crab.animFrame = (crab.animFrame + 1) % 4
		}

		b.drawCrab(crab, t)
	}

	// --- Bubbles (animated: wobble, shimmer, fade, split, pop) ---
	var newBubbles []aquariumBubble
	for i := range b.aquariumBubbles {
		bub := &b.aquariumBubbles[i]
		bub.age++
		bub.y -= bub.speed
		bub.wobble += 0.06 + float64(bub.size)*0.01
		bub.squish += 0.09 + float64(bub.size)*0.02
		bub.shimmer += 0.14

		// Fade in over first 10 ticks
		if bub.opacity < 1.0 {
			bub.opacity += 0.1
			if bub.opacity > 1.0 {
				bub.opacity = 1.0
			}
		}
		// Fade out near top (last 15% of height)
		topFade := 1.0
		fadeZone := float64(pH) * 0.15
		if bub.y < fadeZone {
			topFade = bub.y / fadeZone
			if topFade < 0 {
				topFade = 0
			}
		}

		// Speed varies slightly — larger bubbles accelerate as they rise
		bub.speed += float64(bub.size) * 0.002

		// Only large bubbles (size 3) can split into 2 smaller ones
		if bub.size >= 3 && bub.y < float64(pH)*0.3 && !bub.splitting && b.rng.Float64() < 0.005 {
			bub.splitting = true
			// Spawn two smaller bubbles
			for s := 0; s < 2; s++ {
				child := b.newAquariumBubble(false)
				child.x = bub.x + (b.rng.Float64()-0.5)*3.0
				child.y = bub.y
				child.size = bub.size - 1
				child.speed = bub.speed * (0.8 + b.rng.Float64()*0.4)
				child.opacity = bub.opacity * 0.8
				newBubbles = append(newBubbles, child)
			}
		}

		// Remove if off screen, fully faded, or just split
		if bub.y < -3 || (bub.splitting && bub.age > 2) || topFade < 0.05 {
			// Don't replace bubbles that just split - just remove them
			if bub.splitting && bub.age > 2 {
				b.aquariumBubbles[i] = aquariumBubble{}
			} else {
				b.aquariumBubbles[i] = b.newAquariumBubble(false)
			}
			continue
		}

		// Sinusoidal drift path
		xOff := b.fastSin(bub.wobble) * (0.8 + float64(bub.size)*0.3)
		bx := int(bub.x + xOff + bub.drift*float64(b.frame))
		by := int(bub.y)
		bx = ((bx % w) + w) % w

		depthFrac := bub.y / float64(pH)
		bright := 0.3 + (1.0-depthFrac)*0.5
		alpha := bub.opacity * topFade

		// Squish deformation: bubble breathing/wobbling
		squishX := 1.0 + b.fastSin(bub.squish)*0.15
		squishY := 1.0 - b.fastSin(bub.squish)*0.15

		// Shimmer highlight position rotates around the bubble
		shimX := b.fastSin(bub.shimmer) * 0.6
		shimY := b.fastSin(bub.shimmer+math.Pi*0.5) * 0.6

		// Base bubble color (translucent blue-white)
		baseR := 120.0 * bright
		baseG := 200.0 * bright
		baseB := 255.0 * bright
		// Highlight color (brighter white-blue)
		hiR := 200.0 * bright
		hiG := 240.0 * bright
		hiB := 255.0 * bright
		// Edge/rim color (slightly darker, more blue)
		rimR := 80.0 * bright
		rimG := 160.0 * bright
		rimB := 240.0 * bright

		// Blend a bubble pixel onto the background
		bubbleBlend := func(px, py int, r, g, bv, a float64) {
			a *= alpha
			if a <= 0 || px < 0 || px >= w || py < 0 || py >= pH {
				return
			}
			existing := b.pb.get(px, py)
			nr := float64(existing.r)*(1-a) + r*a
			ng := float64(existing.g)*(1-a) + g*a
			nb := float64(existing.b)*(1-a) + bv*a
			b.pb.set(px, py, uint8(clampF(nr, 0, 255)), uint8(clampF(ng, 0, 255)), uint8(clampF(nb, 0, 255)))
		}

		switch bub.size {
		case 0: // tiny: single pixel, pulses
			pulse := 0.7 + b.fastSin(bub.shimmer*2)*0.3
			bubbleBlend(bx, by, baseR*pulse, baseG*pulse, baseB*pulse, 0.6)

		case 1: // small: 2-3 pixels with shimmer highlight
			bubbleBlend(bx, by, baseR, baseG, baseB, 0.5)
			sx := bx + int(squishX)
			bubbleBlend(sx, by, baseR, baseG, baseB, 0.5)
			// Shimmer highlight
			hx := bx + int(shimX+0.5)
			hy := by + int(shimY-0.5)
			bubbleBlend(hx, hy, hiR, hiG, hiB, 0.35)

		case 2: // medium: ~5px circle, deforms, has rim + highlight
			// Core (squished ellipse)
			bubbleBlend(bx, by, baseR, baseG, baseB, 0.4)
			bubbleBlend(bx+int(squishX), by, baseR, baseG, baseB, 0.4)
			bubbleBlend(bx-int(squishX), by, baseR, baseG, baseB, 0.35)
			bubbleBlend(bx, by-int(squishY), baseR, baseG, baseB, 0.4)
			bubbleBlend(bx, by+int(squishY), baseR, baseG, baseB, 0.35)
			// Additional body pixel
			bubbleBlend(bx+int(squishX), by-int(squishY), baseR, baseG, baseB, 0.3)
			// Rim (edge pixels, slightly darker)
			bubbleBlend(bx-int(squishX)-1, by, rimR, rimG, rimB, 0.2)
			bubbleBlend(bx+int(squishX)+1, by, rimR, rimG, rimB, 0.2)
			bubbleBlend(bx, by+int(squishY)+1, rimR, rimG, rimB, 0.2)
			// Animated highlight (rotates)
			hx := bx + int(shimX*1.2)
			hy := by + int(shimY*1.2) - 1
			bubbleBlend(hx, hy, hiR, hiG, hiB, 0.5)

		case 3: // large: ~8px circle, pronounced wobble, dual highlights
			// Outer rim ring
			for angle := 0.0; angle < math.Pi*2; angle += math.Pi / 4 {
				rx := int(math.Cos(angle)*2.5*squishX + 0.5)
				ry := int(math.Sin(angle)*2.5*squishY + 0.5)
				bubbleBlend(bx+rx, by+ry, rimR, rimG, rimB, 0.25)
			}
			// Inner body fill
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					sx := int(float64(dx) * squishX)
					sy := int(float64(dy) * squishY)
					a := 0.35
					if dx == 0 && dy == 0 {
						a = 0.45
					}
					bubbleBlend(bx+sx, by+sy, baseR, baseG, baseB, a)
				}
			}
			// Extended body pixels along squish axes
			bubbleBlend(bx+int(squishX*2), by, baseR, baseG, baseB, 0.3)
			bubbleBlend(bx-int(squishX*2), by, baseR, baseG, baseB, 0.3)
			bubbleBlend(bx, by-int(squishY*2), baseR, baseG, baseB, 0.3)
			bubbleBlend(bx, by+int(squishY*2), baseR, baseG, baseB, 0.25)
			// Primary highlight (top-left area, rotates)
			hx1 := bx + int(shimX*1.5) - 1
			hy1 := by + int(shimY*1.5) - 1
			bubbleBlend(hx1, hy1, hiR, hiG, hiB, 0.55)
			bubbleBlend(hx1+1, hy1, hiR, hiG, hiB, 0.35)
			// Secondary smaller highlight (opposite side)
			hx2 := bx - int(shimX*0.8) + 1
			hy2 := by - int(shimY*0.8) + 1
			bubbleBlend(hx2, hy2, hiR, hiG, hiB, 0.25)
		}
	}

	// Add child bubbles from splits
	b.aquariumBubbles = append(b.aquariumBubbles, newBubbles...)

	// Remove empty slots (bubbles that split and were zeroed out)
	bubbleCount := 0
	for i := range b.aquariumBubbles {
		if b.aquariumBubbles[i].size != 0 || b.aquariumBubbles[i].speed != 0 {
			b.aquariumBubbles[bubbleCount] = b.aquariumBubbles[i]
			bubbleCount++
		}
	}
	b.aquariumBubbles = b.aquariumBubbles[:bubbleCount]

	// Spawn new bubbles (from seaweed bases, fish, and random)
	spawnRate := 0.015
	if b.rng.Float64() < spawnRate {
		nb := b.newAquariumBubble(false)
		// 30% chance to spawn from a seaweed position
		if len(b.aquariumWeeds) > 0 && b.rng.Float64() < 0.3 {
			weed := b.aquariumWeeds[b.rng.Intn(len(b.aquariumWeeds))]
			nb.x = float64(weed.x) + (b.rng.Float64()-0.5)*2.0
			nb.y = float64(pH-4-weed.height) + b.rng.Float64()*2.0
		}
		// 15% chance to spawn from a fish position (fish exhale)
		if len(b.aquariumFish) > 0 && b.rng.Float64() < 0.15 {
			fish := b.aquariumFish[b.rng.Intn(len(b.aquariumFish))]
			nb.x = fish.x + fish.dir*3.0
			nb.y = fish.y - 1.0
			nb.size = 0 // fish bubbles are always tiny
		}
		b.aquariumBubbles = append(b.aquariumBubbles, nb)
		if len(b.aquariumBubbles) > 25 {
			b.aquariumBubbles = b.aquariumBubbles[1:]
		}
	}
}

// drawDetailedFish renders a detailed fish sprite using the pixel buffer.
func (b *BackgroundModel) drawDetailedFish(fish *aquariumFish, fx, fy float64, t float64) {
	pH := b.height * 2
	w := b.width
	sc := speciesColors[fish.species%len(speciesColors)]
	bright := 0.45 + fish.depth*0.55

	// Pixel setter with bounds check
	setP := func(px, py int, r, g, bv float64) {
		if px >= 0 && px < w && py >= 0 && py < pH {
			cr := uint8(clampF(r*bright, 0, 255))
			cg := uint8(clampF(g*bright, 0, 255))
			cb := uint8(clampF(bv*bright, 0, 255))
			b.pb.set(px, py, cr, cg, cb)
		}
	}

	// Blend setter for smoother edges (blends with existing pixel)
	blendP := func(px, py int, r, g, bv, alpha float64) {
		if px >= 0 && px < w && py >= 0 && py < pH {
			existing := b.pb.get(px, py)
			nr := float64(existing.r)*(1-alpha) + r*bright*alpha
			ng := float64(existing.g)*(1-alpha) + g*bright*alpha
			nb := float64(existing.b)*(1-alpha) + bv*bright*alpha
			b.pb.set(px, py, uint8(clampF(nr, 0, 255)), uint8(clampF(ng, 0, 255)), uint8(clampF(nb, 0, 255)))
		}
	}

	cx := int(fx)
	cy := int(fy)
	d := 1
	if fish.dir < 0 {
		d = -1
	}

	tailWag := b.fastSin(t*5.0+fish.bodyPhase) * 1.2
	finFlutter := b.fastSin(fish.finPhase)

	switch fish.size {
	case 0: // small fish: ~5x3 pixels
		// Body ellipse (3 wide, 3 tall)
		setP(cx, cy, sc.bodyR, sc.bodyG, sc.bodyB)
		setP(cx+d, cy, sc.bodyR, sc.bodyG, sc.bodyB)
		setP(cx, cy-1, sc.bodyR*0.9, sc.bodyG*0.9, sc.bodyB*0.9)
		setP(cx, cy+1, sc.bellyR, sc.bellyG, sc.bellyB)
		// Head
		setP(cx+d*2, cy, sc.bodyR, sc.bodyG, sc.bodyB)
		// Eye
		setP(cx+d*2, cy-1, 230, 230, 240)
		// Tail
		tw := int(tailWag * 0.5)
		setP(cx-d, cy+tw, sc.finR, sc.finG, sc.finB)
		setP(cx-d*2, cy-1+tw, sc.finR, sc.finG, sc.finB)
		setP(cx-d*2, cy+1+tw, sc.finR, sc.finG, sc.finB)

	case 1: // medium fish: ~8x5 pixels
		// Body: tapered ellipse
		// Center row (widest)
		for dx := -1; dx <= 3; dx++ {
			setP(cx+d*dx, cy, sc.bodyR, sc.bodyG, sc.bodyB)
		}
		// Upper body
		for dx := 0; dx <= 3; dx++ {
			shade := 0.85 + float64(dx)*0.03
			setP(cx+d*dx, cy-1, sc.bodyR*shade, sc.bodyG*shade, sc.bodyB*shade)
		}
		// Lower body (belly)
		for dx := 0; dx <= 2; dx++ {
			setP(cx+d*dx, cy+1, sc.bellyR, sc.bellyG, sc.bellyB)
		}
		// Top contour
		setP(cx+d, cy-2, sc.bodyR*0.7, sc.bodyG*0.7, sc.bodyB*0.7)
		setP(cx+d*2, cy-2, sc.bodyR*0.7, sc.bodyG*0.7, sc.bodyB*0.7)
		// Bottom contour
		setP(cx+d, cy+2, sc.bellyR*0.8, sc.bellyG*0.8, sc.bellyB*0.8)

		// Dorsal fin
		finOff := int(finFlutter * 0.3)
		setP(cx+d, cy-3+finOff, sc.finR, sc.finG, sc.finB)
		blendP(cx+d*2, cy-3+finOff, sc.finR, sc.finG, sc.finB, 0.6)

		// Pectoral fin
		pfinY := cy + 2 + int(finFlutter*0.5)
		blendP(cx+d, pfinY, sc.finR*0.8, sc.finG*0.8, sc.finB*0.8, 0.7)

		// Head/nose
		setP(cx+d*4, cy, sc.bodyR, sc.bodyG, sc.bodyB)
		setP(cx+d*4, cy-1, sc.bodyR*0.9, sc.bodyG*0.9, sc.bodyB*0.9)

		// Eye (white + dark pupil)
		setP(cx+d*3, cy-1, 230, 230, 240)
		blendP(cx+d*4, cy-1, 20, 20, 30, 0.5) // pupil hint

		// Tail fin (forked)
		tw := int(tailWag)
		setP(cx-d*2, cy+tw, sc.finR, sc.finG, sc.finB)
		setP(cx-d*3, cy-1+tw, sc.finR, sc.finG, sc.finB)
		setP(cx-d*3, cy+tw, sc.finR*0.8, sc.finG*0.8, sc.finB*0.8)
		setP(cx-d*3, cy+1+tw, sc.finR, sc.finG, sc.finB)
		setP(cx-d*4, cy-2+tw, sc.finR*0.7, sc.finG*0.7, sc.finB*0.7)
		setP(cx-d*4, cy+2+tw, sc.finR*0.7, sc.finG*0.7, sc.finB*0.7)

		// Species markings
		if sc.hasStripes {
			for s := 0; s < sc.stripeCount; s++ {
				sx := cx + d*(s*2)
				for dy := -1; dy <= 1; dy++ {
					blendP(sx, cy+dy, sc.accentR, sc.accentG, sc.accentB, 0.7)
				}
			}
		}
		// Neon tetra light stripe
		if fish.species == 3 {
			for dx := 0; dx <= 3; dx++ {
				blendP(cx+d*dx, cy, sc.accentR, sc.accentG, sc.accentB, 0.5)
			}
		}

	case 2: // large fish: ~12x8 pixels
		// Body ellipse: wider center, tapered head and tail
		// Row cy (widest): 8 pixels
		for dx := -2; dx <= 5; dx++ {
			setP(cx+d*dx, cy, sc.bodyR, sc.bodyG, sc.bodyB)
		}
		// Row cy-1: 7 pixels
		for dx := -1; dx <= 5; dx++ {
			shade := 0.9 + float64(dx)*0.015
			setP(cx+d*dx, cy-1, sc.bodyR*shade, sc.bodyG*shade, sc.bodyB*shade)
		}
		// Row cy+1: 7 pixels (belly)
		for dx := -1; dx <= 5; dx++ {
			mix := float64(dx+2) / 8.0
			r := sc.bodyR*(1-mix) + sc.bellyR*mix
			g := sc.bodyG*(1-mix) + sc.bellyG*mix
			bv := sc.bodyB*(1-mix) + sc.bellyB*mix
			setP(cx+d*dx, cy+1, r, g, bv)
		}
		// Row cy-2: 5 pixels (upper body contour)
		for dx := 0; dx <= 4; dx++ {
			shade := 0.8
			setP(cx+d*dx, cy-2, sc.bodyR*shade, sc.bodyG*shade, sc.bodyB*shade)
		}
		// Row cy+2: 5 pixels (lower body)
		for dx := 0; dx <= 4; dx++ {
			setP(cx+d*dx, cy+2, sc.bellyR*0.9, sc.bellyG*0.9, sc.bellyB*0.9)
		}
		// Row cy-3: 3 pixels (top contour)
		for dx := 1; dx <= 3; dx++ {
			blendP(cx+d*dx, cy-3, sc.bodyR*0.7, sc.bodyG*0.7, sc.bodyB*0.7, 0.8)
		}
		// Row cy+3: 2 pixels (bottom contour)
		blendP(cx+d, cy+3, sc.bellyR*0.7, sc.bellyG*0.7, sc.bellyB*0.7, 0.7)
		blendP(cx+d*2, cy+3, sc.bellyR*0.7, sc.bellyG*0.7, sc.bellyB*0.7, 0.7)

		// Head (tapered)
		setP(cx+d*6, cy, sc.bodyR, sc.bodyG, sc.bodyB)
		setP(cx+d*6, cy-1, sc.bodyR*0.9, sc.bodyG*0.9, sc.bodyB*0.9)
		setP(cx+d*6, cy+1, sc.bellyR, sc.bellyG, sc.bellyB)
		setP(cx+d*7, cy, sc.bodyR*0.95, sc.bodyG*0.95, sc.bodyB*0.95)

		// Mouth
		blendP(cx+d*7, cy+1, 40, 20, 30, 0.4)

		// Eye: white sclera + iris + pupil
		setP(cx+d*5, cy-2, 230, 230, 240) // sclera
		setP(cx+d*6, cy-2, 230, 230, 240) // sclera
		setP(cx+d*5, cy-1, 230, 230, 240) // sclera lower
		// Iris (species-tinted)
		irisR := clampF(sc.bodyR*0.3+50, 0, 255)
		irisG := clampF(sc.bodyG*0.3+50, 0, 255)
		irisB := clampF(sc.bodyB*0.3+50, 0, 255)
		setP(cx+d*6, cy-2, irisR, irisG, irisB)
		// Pupil (dark)
		blendP(cx+d*6, cy-2, 10, 10, 15, 0.6)
		// Eye highlight
		blendP(cx+d*5, cy-2, 255, 255, 255, 0.3)

		// Gill line
		blendP(cx+d*4, cy-1, sc.bodyR*0.5, sc.bodyG*0.5, sc.bodyB*0.5, 0.4)
		blendP(cx+d*4, cy, sc.bodyR*0.5, sc.bodyG*0.5, sc.bodyB*0.5, 0.4)
		blendP(cx+d*4, cy+1, sc.bodyR*0.5, sc.bodyG*0.5, sc.bodyB*0.5, 0.3)

		// Dorsal fin (tall, flowing)
		finOff := int(finFlutter * 0.5)
		for dx := 1; dx <= 4; dx++ {
			finHeight := 2 - (dx-1)/2
			for dy := 1; dy <= finHeight; dy++ {
				fade := 1.0 - float64(dy)*0.3
				setP(cx+d*dx, cy-3-dy+finOff, sc.finR*fade, sc.finG*fade, sc.finB*fade)
			}
		}
		// Fin membrane (semi-transparent between rays)
		blendP(cx+d*2, cy-4+finOff, sc.finR*0.6, sc.finG*0.6, sc.finB*0.6, 0.5)
		blendP(cx+d*3, cy-4+finOff, sc.finR*0.5, sc.finG*0.5, sc.finB*0.5, 0.4)

		// Pectoral fin (side fin)
		pfinY := cy + 3 + int(finFlutter*0.6)
		setP(cx+d*3, pfinY, sc.finR*0.8, sc.finG*0.8, sc.finB*0.8)
		blendP(cx+d*4, pfinY, sc.finR*0.6, sc.finG*0.6, sc.finB*0.6, 0.6)
		blendP(cx+d*3, pfinY+1, sc.finR*0.5, sc.finG*0.5, sc.finB*0.5, 0.4)

		// Anal fin (bottom rear)
		blendP(cx, cy+3, sc.finR*0.7, sc.finG*0.7, sc.finB*0.7, 0.6)
		blendP(cx+d, cy+3, sc.finR*0.6, sc.finG*0.6, sc.finB*0.6, 0.5)

		// Caudal fin (tail) — forked shape with wag animation
		tw := int(tailWag)
		// Tail peduncle (narrow connection)
		setP(cx-d*3, cy+tw, sc.finR, sc.finG, sc.finB)
		setP(cx-d*3, cy-1+tw, sc.finR*0.8, sc.finG*0.8, sc.finB*0.8)
		setP(cx-d*3, cy+1+tw, sc.finR*0.8, sc.finG*0.8, sc.finB*0.8)
		// Upper fork
		setP(cx-d*4, cy-1+tw, sc.finR, sc.finG, sc.finB)
		setP(cx-d*4, cy-2+tw, sc.finR, sc.finG, sc.finB)
		setP(cx-d*5, cy-2+tw, sc.finR*0.8, sc.finG*0.8, sc.finB*0.8)
		setP(cx-d*5, cy-3+tw, sc.finR*0.7, sc.finG*0.7, sc.finB*0.7)
		setP(cx-d*6, cy-3+tw, sc.finR*0.5, sc.finG*0.5, sc.finB*0.5)
		// Lower fork
		setP(cx-d*4, cy+1+tw, sc.finR, sc.finG, sc.finB)
		setP(cx-d*4, cy+2+tw, sc.finR, sc.finG, sc.finB)
		setP(cx-d*5, cy+2+tw, sc.finR*0.8, sc.finG*0.8, sc.finB*0.8)
		setP(cx-d*5, cy+3+tw, sc.finR*0.7, sc.finG*0.7, sc.finB*0.7)
		setP(cx-d*6, cy+3+tw, sc.finR*0.5, sc.finG*0.5, sc.finB*0.5)
		// Tail membrane
		blendP(cx-d*4, cy+tw, sc.finR*0.6, sc.finG*0.6, sc.finB*0.6, 0.5)
		blendP(cx-d*5, cy-1+tw, sc.finR*0.5, sc.finG*0.5, sc.finB*0.5, 0.4)
		blendP(cx-d*5, cy+1+tw, sc.finR*0.5, sc.finG*0.5, sc.finB*0.5, 0.4)

		// Species-specific markings
		if sc.hasStripes {
			for s := 0; s < sc.stripeCount; s++ {
				sx := cx + d*(s*3)
				for dy := -2; dy <= 2; dy++ {
					blendP(sx, cy+dy, sc.accentR, sc.accentG, sc.accentB, 0.65)
				}
			}
		}
		// Neon tetra: horizontal light stripe
		if fish.species == 3 {
			for dx := -1; dx <= 5; dx++ {
				blendP(cx+d*dx, cy, sc.accentR, sc.accentG, sc.accentB, 0.45)
			}
			// Red front half
			for dx := -1; dx <= 2; dx++ {
				blendP(cx+d*dx, cy-1, 200, 30, 50, 0.3)
				blendP(cx+d*dx, cy+1, 200, 30, 50, 0.3)
			}
		}
		// Royal gramma: front purple, back yellow split
		if fish.species == 5 {
			for dx := -2; dx <= 1; dx++ {
				for dy := -2; dy <= 2; dy++ {
					blendP(cx+d*dx, cy+dy, sc.bellyR, sc.bellyG, sc.bellyB, 0.4)
				}
			}
		}
		// Blue tang: dark accent band through body
		if fish.species == 1 {
			for dx := 0; dx <= 4; dx++ {
				blendP(cx+d*dx, cy, sc.accentR, sc.accentG, sc.accentB, 0.35)
			}
			// Yellow tail accent
			blendP(cx-d*4, cy-1+tw, 220, 200, 50, 0.5)
			blendP(cx-d*4, cy+1+tw, 220, 200, 50, 0.5)
		}

		// Subtle scale pattern (body shimmer)
		for dx := 0; dx <= 4; dx++ {
			for dy := -1; dy <= 1; dy++ {
				if (dx+dy)%3 == 0 {
					blendP(cx+d*dx, cy+dy, 255, 255, 255, 0.06)
				}
			}
		}
	}
}

// --- Crab types and colors (Elthen-style pixel art palette) ---

type crabColors struct {
	outR, outG, outB       float64 // dark outline
	shellR, shellG, shellB float64 // main shell body
	highR, highG, highB    float64 // shell highlights (golden)
	bellyR, bellyG, bellyB float64 // lighter belly/underside
	legR, legG, legB       float64 // legs/claws (dark)
}

var crabVariants = []crabColors{
	// 0: Classic orange-brown (Elthen original)
	{40, 24, 8, 200, 120, 40, 224, 160, 48, 232, 176, 80, 48, 32, 16},
	// 1: Reddish variant
	{48, 16, 8, 190, 64, 32, 220, 100, 40, 224, 140, 72, 56, 24, 12},
	// 2: Blue/teal variant
	{12, 32, 48, 48, 140, 180, 80, 180, 210, 120, 200, 220, 16, 40, 56},
}

// SetTasks updates the task labels for aquarium crabs.
func (b *BackgroundModel) SetTasks(tasks []string) {
	b.aquariumTasks = tasks
	if b.mode == BgAquarium {
		b.syncCrabs()
	}
}

// syncCrabs ensures we have one crab per task, reusing existing crabs where possible.
// Also updates crab Y positions when the terminal is resized to keep them on the seafloor.
func (b *BackgroundModel) syncCrabs() {
	if b.width == 0 || b.height == 0 {
		return
	}

	// Calculate the sand Y position based on current height
	pH := float64(b.height * 2)
	sandY := pH - 6.0

	tasks := b.aquariumTasks

	// Always have at least one idle crab even with no tasks
	if len(tasks) == 0 {
		if len(b.aquariumCrabs) == 0 {
			b.aquariumCrabs = []aquariumCrab{b.newCrab("")}
		} else {
			// Clear task labels on existing crabs, keep one
			b.aquariumCrabs = b.aquariumCrabs[:1]
			b.aquariumCrabs[0].task = ""
			// Update Y position to match new height
			b.aquariumCrabs[0].y = sandY
		}
		return
	}

	// Match tasks to existing crabs by index, add/remove as needed
	for i, task := range tasks {
		if i < len(b.aquariumCrabs) {
			b.aquariumCrabs[i].task = task
			// Update Y position to keep crab on seafloor after resize
			b.aquariumCrabs[i].y = sandY
			if task != "" {
				b.aquariumCrabs[i].animState = 1 // walk when has task
			} else {
				b.aquariumCrabs[i].animState = 0 // idle when no task
			}
		} else {
			b.aquariumCrabs = append(b.aquariumCrabs, b.newCrab(task))
		}
	}
	// Trim excess crabs
	if len(b.aquariumCrabs) > len(tasks) {
		b.aquariumCrabs = b.aquariumCrabs[:len(tasks)]
	}
}

func (b *BackgroundModel) newCrab(task string) aquariumCrab {
	dir := 1.0
	if b.rng.Intn(2) == 0 {
		dir = -1.0
	}
	pH := float64(b.height * 2)
	// Crabs sit on the sandy bottom; cy is shell bottom, legs extend 2px below
	sandY := pH - 6.0

	return aquariumCrab{
		x:        b.rng.Float64() * float64(b.width),
		y:        sandY,
		speed:    0.06 + b.rng.Float64()*0.12,
		dir:      dir,
		legPhase: b.rng.Float64() * math.Pi * 2,
		task:     task,
		variant:  b.rng.Intn(len(crabVariants)),
	}
}

// drawCrab renders an Elthen-style pixel art crab on the pixel buffer.
// Dome/arc shell with bumpy segments, small claws, stubby legs, eyes at shell-top.
func (b *BackgroundModel) drawCrab(crab *aquariumCrab, t float64) {
	pH := b.height * 2
	w := b.width
	cc := crabVariants[crab.variant%len(crabVariants)]

	// Body bob animation (1px up/down)
	bob := int(b.fastSin(crab.legPhase*0.5) * 0.6)
	cx := int(crab.x)
	cy := int(crab.y) + bob
	d := 1
	if crab.dir < 0 {
		d = -1
	}

	setP := func(px, py int, r, g, bv float64) {
		if px >= 0 && px < w && py >= 0 && py < pH {
			b.pb.set(px, py, uint8(clampF(r, 0, 255)), uint8(clampF(g, 0, 255)), uint8(clampF(bv, 0, 255)))
		}
	}
	blendP := func(px, py int, r, g, bv, alpha float64) {
		if px >= 0 && px < w && py >= 0 && py < pH {
			existing := b.pb.get(px, py)
			nr := float64(existing.r)*(1-alpha) + r*alpha
			ng := float64(existing.g)*(1-alpha) + g*alpha
			nb := float64(existing.b)*(1-alpha) + bv*alpha
			b.pb.set(px, py, uint8(clampF(nr, 0, 255)), uint8(clampF(ng, 0, 255)), uint8(clampF(nb, 0, 255)))
		}
	}

	// Animation-state-dependent parameters
	var bodyBob float64   // extra body bob amplitude
	var clawScale float64 // claw size multiplier
	var clawExtend int    // extra claw reach
	var legAmp float64    // leg movement amplitude

	switch crab.animState {
	case 0: // idle: subtle claw movement, minimal body bob
		bodyBob = 0.0
		clawScale = 1.0
		clawExtend = 0
		legAmp = 0.3
	case 1: // walk: more leg movement, slight body bob, medium claws
		bodyBob = 0.8
		clawScale = 1.2
		clawExtend = 0
		legAmp = 0.8
	case 2: // attack: claws extend wide, dramatic
		bodyBob = 0.3
		clawScale = 1.6
		clawExtend = 2
		legAmp = 0.4
	default: // death or other
		bodyBob = 0.0
		clawScale = 0.5
		clawExtend = 0
		legAmp = 0.1
	}

	// Frame-based claw animation: idle subtly opens/closes, attack snaps
	framePhase := float64(crab.animFrame) / 4.0 * math.Pi * 2
	clawAnim := b.fastSin(framePhase) * clawScale

	// Apply extra body bob for walk state
	if bodyBob > 0 {
		walkBob := int(b.fastSin(framePhase*2) * bodyBob)
		cy += walkBob
	}

	// Shorthand colors
	oR, oG, oB := cc.outR, cc.outG, cc.outB       // outline
	sR, sG, sB := cc.shellR, cc.shellG, cc.shellB  // shell body
	hR, hG, hB := cc.highR, cc.highG, cc.highB     // highlight
	bR, bG, bB := cc.bellyR, cc.bellyG, cc.bellyB  // belly
	lR, lG, lB := cc.legR, cc.legG, cc.legB        // legs

	// --- Dome shell (flat bottom, rounded top, ~13px wide × 7px tall) ---
	// cy is the bottom of the shell; shell extends upward

	// Row cy-6: bumpy top ridge — 3 scalloped segments with outline
	setP(cx-3, cy-6, oR, oG, oB)
	setP(cx-2, cy-6, hR, hG, hB)
	setP(cx-1, cy-6, oR, oG, oB)
	setP(cx, cy-6, hR, hG, hB)
	setP(cx+1, cy-6, oR, oG, oB)
	setP(cx+2, cy-6, hR, hG, hB)
	setP(cx+3, cy-6, oR, oG, oB)

	// Row cy-5: upper dome outline + fill
	setP(cx-4, cy-5, oR, oG, oB)
	for dx := -3; dx <= 3; dx++ {
		setP(cx+dx, cy-5, sR, sG, sB)
	}
	setP(cx+4, cy-5, oR, oG, oB)
	// Highlight on upper dome
	setP(cx-1, cy-5, hR, hG, hB)
	setP(cx, cy-5, hR, hG, hB)
	setP(cx+1, cy-5, hR, hG, hB)

	// Row cy-4: wider shell
	setP(cx-5, cy-4, oR, oG, oB)
	for dx := -4; dx <= 4; dx++ {
		setP(cx+dx, cy-4, sR, sG, sB)
	}
	setP(cx+5, cy-4, oR, oG, oB)
	// Highlight band
	setP(cx-2, cy-4, hR, hG, hB)
	setP(cx-1, cy-4, hR, hG, hB)
	setP(cx, cy-4, hR, hG, hB)
	setP(cx+1, cy-4, hR, hG, hB)

	// Row cy-3: full width shell
	setP(cx-6, cy-3, oR, oG, oB)
	for dx := -5; dx <= 5; dx++ {
		setP(cx+dx, cy-3, sR, sG, sB)
	}
	setP(cx+6, cy-3, oR, oG, oB)
	// Subtle highlight
	setP(cx-1, cy-3, hR, hG, hB)
	setP(cx, cy-3, hR, hG, hB)

	// Row cy-2: widest shell body
	setP(cx-6, cy-2, oR, oG, oB)
	for dx := -5; dx <= 5; dx++ {
		setP(cx+dx, cy-2, sR, sG, sB)
	}
	setP(cx+6, cy-2, oR, oG, oB)

	// Row cy-1: lower shell, belly tint
	setP(cx-6, cy-1, oR, oG, oB)
	for dx := -5; dx <= 5; dx++ {
		mix := 0.4
		r := sR*(1-mix) + bR*mix
		g := sG*(1-mix) + bG*mix
		bv := sB*(1-mix) + bB*mix
		setP(cx+dx, cy-1, r, g, bv)
	}
	setP(cx+6, cy-1, oR, oG, oB)

	// Row cy: bottom of shell (flat base, outlined)
	setP(cx-6, cy, oR, oG, oB)
	for dx := -5; dx <= 5; dx++ {
		setP(cx+dx, cy, bR, bG, bB)
	}
	setP(cx+6, cy, oR, oG, oB)
	// Bottom outline
	for dx := -5; dx <= 5; dx++ {
		blendP(cx+dx, cy+1, oR, oG, oB, 0.5)
	}

	// --- Eyes (at front-top of shell, NOT on stalks) ---
	// Place eyes on the shell surface near the front
	eyeX1 := cx + d*2
	eyeX2 := cx + d*4
	eyeY := cy - 5
	// White eyeball
	setP(eyeX1, eyeY, 255, 255, 255)
	setP(eyeX2, eyeY, 255, 255, 255)
	// Dark pupil (toward the direction of movement)
	setP(eyeX1+d, eyeY, 10, 10, 15)
	setP(eyeX2+d, eyeY, 10, 10, 15)

	// --- Claws (size/position varies by animation state) ---
	clawOff := int(clawAnim * 0.5)
	// Front claw (direction crab faces)
	setP(cx+d*(7+clawExtend), cy-2+clawOff, lR, lG, lB)
	setP(cx+d*(7+clawExtend), cy-3+clawOff, lR, lG, lB)
	setP(cx+d*(8+clawExtend), cy-3+clawOff, lR, lG, lB) // pincer tip
	// Extra pincer pixel for attack state
	if crab.animState == 2 {
		setP(cx+d*(9+clawExtend), cy-3+clawOff, lR, lG, lB)
		setP(cx+d*(9+clawExtend), cy-4+clawOff, lR, lG, lB)
	}
	// Rear claw (opposite side)
	setP(cx-d*(7+clawExtend), cy-2-clawOff, lR, lG, lB)
	setP(cx-d*(7+clawExtend), cy-3-clawOff, lR, lG, lB)
	setP(cx-d*(8+clawExtend), cy-3-clawOff, lR, lG, lB) // pincer tip
	if crab.animState == 2 {
		setP(cx-d*(9+clawExtend), cy-3-clawOff, lR, lG, lB)
		setP(cx-d*(9+clawExtend), cy-4-clawOff, lR, lG, lB)
	}

	// --- Legs (2-3 stubby nubs per side, amplitude varies by animation state) ---
	for leg := 0; leg < 3; leg++ {
		// Use animFrame to offset leg phases for more varied walk cycles
		phase := crab.legPhase + float64(leg)*1.0 + float64(crab.animFrame)*0.4
		legOff := int(b.fastSin(phase) * legAmp)

		// Right-side legs
		lx := cx + 2 + leg*2
		setP(lx, cy+1+legOff, lR, lG, lB)
		setP(lx, cy+2+legOff, lR*0.7, lG*0.7, lB*0.7)
		// Left-side legs (mirrored)
		lx = cx - 2 - leg*2
		setP(lx, cy+1-legOff, lR, lG, lB)
		setP(lx, cy+2-legOff, lR*0.7, lG*0.7, lB*0.7)
	}
}


// DropFood drops count food particles near the top of the water.
func (b *BackgroundModel) DropFood(count int) {
	if b.width == 0 || b.height == 0 {
		return
	}
	for i := 0; i < count; i++ {
		size := 0
		if b.rng.Float64() < 0.35 {
			size = 1
		}
		b.aquariumFood = append(b.aquariumFood, aquariumFood{
			x:         b.rng.Float64() * float64(b.width),
			y:         2.0 + b.rng.Float64()*2.0,
			sinkSpeed: 0.15 + b.rng.Float64()*0.20,
			wobble:    b.rng.Float64() * math.Pi * 2,
			size:      size,
		})
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// crabLabel represents a task label to overlay as text above a crab.
type crabLabel struct {
	col  int    // terminal column
	row  int    // terminal row (above the crab)
	text string // task text
}

// CrabLabels returns the current crab task labels with their screen positions.
func (b *BackgroundModel) CrabLabels() []crabLabel {
	if b.mode != BgAquarium {
		return nil
	}
	var labels []crabLabel
	for _, crab := range b.aquariumCrabs {
		if crab.task == "" {
			continue
		}
		// Convert pixel coords to terminal row/col
		col := int(crab.x)
		row := int(crab.y)/2 - 3 // 3 rows above the crab body
		if row < 0 {
			row = 0
		}
		// Truncate long tasks
		task := crab.task
		if len(task) > 30 {
			task = task[:27] + "..."
		}
		// Add crab emoji prefix
		task = "🦀 " + task
		labels = append(labels, crabLabel{col: col, row: row, text: task})
	}
	return labels
}
