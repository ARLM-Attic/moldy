package world

import (
	"fmt"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Subscriber func(ev interface{}) error

type mold struct {
	world      *world
	Name       string
	Bits       posBoolMap
	roomy      posBoolMap
	threatened map[pos]uint8
	power      map[pos]uint8
	Targets    posUint16Map
}

func (self *mold) addTarget(precision uint16, p pos) {
	self.Targets[p] = precision
}

func (self *mold) size() uint16 {
	return uint16(len(self.Bits))
}

func (self *mold) set(p pos) {
	if _, found := self.Bits[p]; !found {
		self.Bits[p] = true
		var threat uint8
		p.eachNeighbour(self.world.Dimensions, func(p2 pos) bool {
			self.power[p2]++
			if owner, found := self.world.owner(p2); found && owner != self.Name {
				threat++
			}
			return false
		})
		if threat > 0 {
			self.threatened[p] = threat
		}
		self.world.set(p, self.Name)
		if self.world.hasSpace(p) {
			self.roomy[p] = true
		}
	} else {
		panic(fmt.Errorf("Tried to set %v in %v, but it was already set", p, self.Name))
	}
}

func (self *mold) noSpace(p pos) {
	delete(self.roomy, p)
}

func (self *mold) space(p pos) {
	if self.Bits[p] {
		self.roomy[p] = true
	}
}

func (self *mold) threaten(p pos) {
	if self.Bits[p] {
		self.threatened[p]++
	}
}

func (self *mold) secure(p pos) {
	if pow, found := self.threatened[p]; found {
		if pow == 1 {
			delete(self.threatened, p)
		} else {
			self.threatened[p] = pow - 1
		}
	}
}

func (self *mold) pow(p pos) uint8 {
	return self.power[p]
}

func (self *mold) unset(p pos) {
	if _, found := self.Bits[p]; found {
		delete(self.Bits, p)
		p.eachNeighbour(self.world.Dimensions, func(p2 pos) bool {
			self.power[p2]--
			return false
		})
		delete(self.roomy, p)
		delete(self.threatened, p)
		self.world.unset(p)
	} else {
		panic(fmt.Errorf("Tried to unset %v in %v, but it was never set", p, self.Name))
	}
}

func (self *mold) get(p pos) bool {
	return self.Bits[p]
}

func (self *mold) shrink(delta *Delta) {
	for bit, threat := range self.threatened {
		if self.power[bit] < threat {
			delta.Removed[bit] = self.Name
			self.unset(bit)
		}
	}
}

func (self *mold) grow(delta *Delta, k float32) {
	num := int(k * float32(self.world.MoldGrowth))
	for bit, _ := range self.roomy {
		bit.eachNeighbour(self.world.Dimensions, func(neigh pos) bool {
			if !self.world.hasMold(neigh) {
				delta.Created[neigh] = self.Name
				self.set(neigh)
				num--
				return true
			}
			return false
		})
		if num < 1 {
			break
		}
	}
}

func (self *mold) moveTowards(delta *Delta, target pos, precision uint16) {
	var bestPos *pos
	var bestDistance int64
	var tries uint16
	for n := 0; n < int(self.world.MoldMovement); n++ {
		bestPos = nil
		bestDistance = 0
		tries = precision
		for p, _ := range self.roomy {
			if p2 := p.neighbourTowards(self.world.Dimensions, target); p2 != nil {
				if !self.world.hasMold(*p2) {
					if dist := p2.distance(target); bestPos == nil || dist < bestDistance {
						bestPos = p2
						bestDistance = dist
					}
					tries--
				}
				if tries < 1 {
					break
				}
			}
		}
		if bestPos != nil {
			delta.Created[*bestPos] = self.Name
			self.set(*bestPos)
			for p, _ := range self.roomy {
				if p2 := p.neighbourTowards(self.world.Dimensions, target); p2 != nil {
					if self.world.hasMold(*p2) {
						delta.Removed[p] = self.Name
						self.unset(p)
						break
					}
				}
			}
		}
	}
}

func (self *mold) move(delta *Delta) {
	for target, precision := range self.Targets {
		self.moveTowards(delta, target, precision)
	}
}

type world struct {
	Dimensions   pos
	Molds        map[string]*mold
	MaxMoldSize  uint16
	MoldGrowth   uint8
	MoldMovement uint8
	neighbours   map[pos]uint8
	moldBits     map[pos]string
	cmd          CmdChan
	subscribers  map[*Subscriber]bool
}

func New(width, height, maxMoldSize uint16, moldGrowth, moldMovement uint8) CmdChan {
	w := &world{
		Molds:        make(map[string]*mold),
		Dimensions:   pos{width, height},
		MaxMoldSize:  maxMoldSize,
		neighbours:   make(map[pos]uint8),
		moldBits:     make(map[pos]string),
		cmd:          make(CmdChan),
		subscribers:  make(map[*Subscriber]bool),
		MoldGrowth:   moldGrowth,
		MoldMovement: moldMovement,
	}
	go w.mainLoop()
	return w.cmd
}

func (self *world) hasSpace(p pos) bool {
	return self.neighbours[p] < 8
}

func (self *world) owner(p pos) (result string, found bool) {
	result, found = self.moldBits[p]
	return
}

func (self *world) set(p pos, name string) {
	if n, found := self.moldBits[p]; !found {
		self.moldBits[p] = name
		p.eachNeighbour(self.Dimensions, func(p2 pos) bool {
			for _, mold := range self.Molds {
				if mold.Name != name {
					mold.threaten(p2)
				}
			}
			self.neighbours[p2]++
			if self.neighbours[p2] > 7 {
				for _, mold := range self.Molds {
					mold.noSpace(p2)
				}
			}
			return false
		})
	} else {
		panic(fmt.Errorf("Tried to set %v which was already occupied by %v", p, n))
	}
}

func (self *world) unset(p pos) {
	if owner, found := self.moldBits[p]; found {
		delete(self.moldBits, p)
		p.eachNeighbour(self.Dimensions, func(p2 pos) bool {
			for _, mold := range self.Molds {
				if mold.Name != owner {
					mold.secure(p2)
				}
			}
			if self.neighbours[p2] == 8 {
				for _, mold := range self.Molds {
					mold.space(p2)
				}
			}
			self.neighbours[p2]--
			return false
		})
	} else {
		panic(fmt.Errorf("Tried to unset %v which was not occupied by anyone", p))
	}
}

func (self *world) hasMold(p pos) (result bool) {
	_, result = self.moldBits[p]
	return
}

func (self *world) rand() pos {
	return pos{uint16(rand.Int()) % self.Dimensions.X, uint16(rand.Int()) % self.Dimensions.Y}
}

func (self *world) newMold(name string) {
	p := self.rand()
	for self.hasMold(p) {
		p = self.rand()
	}
	m := &mold{
		world:      self,
		Bits:       make(posBoolMap),
		Name:       name,
		roomy:      make(posBoolMap),
		threatened: make(map[pos]uint8),
		power:      make(map[pos]uint8),
		Targets:    make(posUint16Map),
	}
	m.set(p)
	self.Molds[name] = m
	return
}

func (self *world) growMolds(delta *Delta) {
	var k float32
	var diff int64
	for _, mold := range self.Molds {
		if diff = int64(self.MaxMoldSize) - int64(mold.size()); diff > 0 {
			if k = float32(diff) / float32(self.MaxMoldSize); k > rand.Float32() {
				mold.grow(delta, k)
			}
		}
	}
}

func (self *world) shrinkMolds(delta *Delta) {
	for _, mold := range self.Molds {
		mold.shrink(delta)
	}
}

func (self *world) moveMolds(delta *Delta) {
	for _, mold := range self.Molds {
		mold.move(delta)
	}
}

func (self *world) emit(ev interface{}) {
	for subscriber, _ := range self.subscribers {
		if err := (*subscriber)(ev); err != nil {
			fmt.Println(err)
			delete(self.subscribers, subscriber)
		}
	}
}

func (self *world) addTarget(name string, precision uint16, p pos) {
	self.Molds[name].addTarget(precision, p)
}

func (self *world) tick() {
	delta := &Delta{
		Created: make(posStringMap),
		Removed: make(posStringMap),
	}
	self.growMolds(delta)
	self.shrinkMolds(delta)
	self.moveMolds(delta)
	self.emit(delta)
}

func (self *world) handleCommand(c cmd) {
	switch c.typ {
	case cmdGetState:
		c.ret <- self
	case cmdSubscribe:
		sub := c.arg.(Subscriber)
		self.subscribers[&sub] = true
		c.ret <- nil
	case cmdNewMold:
		name := c.arg.(string)
		self.newMold(name)
		c.ret <- nil
	case cmdAddTarget:
		t := c.arg.(target)
		self.addTarget(t.name, uint16(t.precision), t.pos)
		c.ret <- nil
	}
}

func (self *world) mainLoop() {
	timer := time.Tick(time.Millisecond * 20)
	for {
		select {
		case <-timer:
			self.tick()
		case c := <-self.cmd:
			self.handleCommand(c)
		}
	}
}
