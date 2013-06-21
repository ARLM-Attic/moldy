package world

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type pos struct {
	x int16
	y int16
}

func (self pos) neighbours() (result []pos) {
	result = make([]pos, 0, 8)
	var xd int16
	var yd int16
	for xd = -1; xd < 1; xd++ {
		for yd = -1; yd < 1; yd++ {
			if xd != 0 || yd != 0 {
        result = append(result, pos{self.x + xd, self.y + yd})
			}
		}
	}
	return
}

type mold struct {
	world *World
	bits map[pos]bool
}

func (self *mold) set(p pos) {
	self.bits[pos] = true
}

func (self *mold) unset(p pos) {
	delete(self.bits, p)
}

func (self *mold) get(p pos) bool {
	return self.bits[p]
}

func (self *mold) grow() {
	for bit, _ := range self.bits {
    if !self.World.hasMold(bit) {
			self.bits[bit] = true
			return
		}
	}
}

type World struct {
	width  int
	height int
	molds  map[string]*mold
	stop   chan bool
	maxMoldSize int16
}

func New(width, height, maxMoldSize int16) (result *World) {
	result = &World{
		molds:  make(map[string]*mold),
		stop:   make(chan bool),
		width:  width,
		height: height,
		maxMoldSize: maxMoldSize,
	}
	go result.mainLoop()
	return
}

func (self *World) hasMold(p pos) bool {
	for _, mold := range self.molds {
		if mold.get(p) {
			return true
		}
	}
	return false
}

func (self *World) newMold(name string) {
	newX := int16(rand.Int() % self.width)
	newY := int16(rand.Int() % self.height)
	for self.hasMold(newX, newY) {
		newX = int16(rand.Int() % self.width)
		newY = int16(rand.Int() % self.height)
	}
	m:= &mold{
		bits: make(map[bit]bool),
	}
	m.set(newX, newY)
	self.molds[name] = m
	return
}

func (self *World) growMolds() {
	for _, mold := range self.molds {
		if mold.size() < maxMoldSize {
			if rand.Float32 > float32(mold.Size()) / float32(self.maxMoldSize) {
				mold.grow()
			}
		}
	}
}

func (self *World) tick() {
	self.growMolds()
}

func (self *World) mainLoop() {
	timer := time.Tick(time.Millisecond * 20)
	for {
		select {
		case <- timer:
			self.tick()
		case <- self.stop:
      break
	}
}
